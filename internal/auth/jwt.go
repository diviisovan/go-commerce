package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// signingMethod is fixed at HS256. Pinning it and passing it to
// jwt.WithValidMethods below is what closes the "algorithm confusion" attack,
// where a forged token claims alg:none or swaps HMAC for RSA verification.
var signingMethod = jwt.SigningMethodHS256

// MinSecretLength mirrors the HMAC-SHA256 block size. A shorter secret weakens
// the signature and is rejected at startup rather than at request time.
const MinSecretLength = 32

var (
	ErrInvalidToken = errors.New("invalid or expired token")
	ErrWeakSecret   = fmt.Errorf("JWT secret must be at least %d characters", MinSecretLength)
)

// Claims is the access-token payload. Keep it small: it is sent on every
// request, it is only base64-encoded rather than encrypted, and it cannot be
// revoked before it expires.
type Claims struct {
	UserID uint   `json:"uid"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// TokenManager issues and verifies tokens.
type TokenManager struct {
	secret     []byte
	issuer     string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// NewTokenManager validates the secret and returns a manager.
func NewTokenManager(secret, issuer string, accessTTL, refreshTTL time.Duration) (*TokenManager, error) {
	if len(secret) < MinSecretLength {
		return nil, ErrWeakSecret
	}
	return &TokenManager{
		secret:     []byte(secret),
		issuer:     issuer,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}, nil
}

// AccessTTL exposes the access-token lifetime for the expires_in field.
func (tm *TokenManager) AccessTTL() time.Duration { return tm.accessTTL }

// RefreshTTL exposes the refresh-token lifetime.
func (tm *TokenManager) RefreshTTL() time.Duration { return tm.refreshTTL }

// GenerateAccessToken returns a signed JWT and the moment it expires.
func (tm *TokenManager) GenerateAccessToken(userID uint, role string) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(tm.accessTTL)

	jti, err := randomString(16)
	if err != nil {
		return "", time.Time{}, err
	}

	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatUint(uint64(userID), 10),
			Issuer:    tm.issuer,
			ID:        jti,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	signed, err := jwt.NewWithClaims(signingMethod, claims).SignedString(tm.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign access token: %w", err)
	}
	return signed, expiresAt, nil
}

// ParseAccessToken verifies the signature, algorithm, issuer and expiry, then
// returns the claims. Every failure collapses into ErrInvalidToken so the
// response never tells a caller why their token was rejected.
func (tm *TokenManager) ParseAccessToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) { return tm.secret, nil },
		jwt.WithValidMethods([]string{signingMethod.Alg()}),
		jwt.WithIssuer(tm.issuer),
		jwt.WithExpirationRequired(),
	)
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}
	if claims.UserID == 0 {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

// GenerateRefreshToken returns an opaque token to hand to the client, the
// SHA-256 hash to persist, and the expiry.
//
// Refresh tokens are deliberately NOT JWTs. They must be revocable, and a
// random string checked against a database row can be revoked instantly,
// whereas a self-contained JWT is valid until it expires.
func (tm *TokenManager) GenerateRefreshToken() (plain, hash string, expiresAt time.Time, err error) {
	plain, err = randomString(32)
	if err != nil {
		return "", "", time.Time{}, err
	}
	return plain, HashRefreshToken(plain), time.Now().Add(tm.refreshTTL), nil
}

// HashRefreshToken returns the hex SHA-256 of a refresh token.
//
// Plain SHA-256 is correct here, unlike for passwords: the token is 256 bits of
// cryptographic randomness, so there is no dictionary to attack and no need for
// a slow hash.
func HashRefreshToken(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}

// randomString returns n cryptographically random bytes, URL-safe base64 encoded.
func randomString(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("read random bytes: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
