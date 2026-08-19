package services

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"go-ecommerce/config"
	"go-ecommerce/database"
	"go-ecommerce/internal/auth"
	"go-ecommerce/models"

	"gorm.io/gorm"
)

// Sentinel errors. The controller maps these to HTTP status codes, which keeps
// status-code decisions out of the service and HTTP concerns out of the domain.
var (
	ErrEmailTaken          = errors.New("email already registered")
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrAccountLocked       = errors.New("account temporarily locked due to too many failed login attempts")
	ErrAccountInactive     = errors.New("account is inactive")
	ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")
	ErrUserNotFound        = errors.New("user not found")
	ErrSamePassword        = errors.New("new password must differ from the current one")

	// errTokenReplay is internal. It signals that a rotated token was presented
	// again, so the caller can revoke the family after the transaction unwinds.
	errTokenReplay = errors.New("refresh token replay detected")
)

// AuthService handles authentication and session business logic
type AuthService struct {
	tokens *auth.TokenManager
	cfg    config.AuthConfig
}

// NewAuthService creates a new auth service. It returns an error because a
// misconfigured secret must stop the process at startup.
func NewAuthService(cfg config.AuthConfig) (*AuthService, error) {
	tokens, err := auth.NewTokenManager(cfg.JWTSecret, cfg.JWTIssuer, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
	if err != nil {
		return nil, err
	}
	return &AuthService{tokens: tokens, cfg: cfg}, nil
}

// TokenManager exposes the manager so the authentication middleware can verify
// tokens with the exact same configuration that issued them.
func (s *AuthService) TokenManager() *auth.TokenManager {
	return s.tokens
}

// Signup registers a new account and logs it straight in.
func (s *AuthService) Signup(req *models.SignupRequest, userAgent, ip string) (*models.AuthResponse, error) {
	email := normalizeEmail(req.Email)
	name := strings.TrimSpace(req.Name)

	if err := auth.ValidatePasswordStrength(req.Password, email, name); err != nil {
		return nil, err
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Name:         name,
		Email:        email,
		Address:      strings.TrimSpace(req.Address),
		Phone:        strings.TrimSpace(req.Phone),
		PasswordHash: hash,
		Role:         models.RoleCustomer, // never accept the role from the request body
		IsActive:     true,
	}

	// Let the unique index decide. Checking "does this email exist?" first and
	// inserting afterwards is a race: two concurrent signups can both pass the
	// check before either one inserts.
	if err := database.GetDB().Create(user).Error; err != nil {
		if isDuplicateKeyError(err) {
			return nil, ErrEmailTaken
		}
		return nil, fmt.Errorf("create user: %w", err)
	}

	return s.issueTokens(user, userAgent, ip)
}

// Login verifies credentials and issues a token pair.
func (s *AuthService) Login(req *models.LoginRequest, userAgent, ip string) (*models.AuthResponse, error) {
	email := normalizeEmail(req.Email)

	var user models.User
	err := database.GetDB().Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Spend the same ~250ms a real bcrypt check costs, then return the
			// identical error a wrong password returns. Both halves are needed:
			// matching timing and matching message. Either one alone still lets
			// an attacker discover which emails are registered.
			auth.BurnPasswordCompare()
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("lookup user: %w", err)
	}

	// Reporting the lockout does leak that the account exists. That is a
	// deliberate trade for usability: without it, a locked-out user sees
	// "wrong password" and keeps retrying a password that is actually correct.
	if user.IsLocked() {
		return nil, ErrAccountLocked
	}
	if !user.IsActive {
		return nil, ErrAccountInactive
	}
	// Accounts predating password auth (the seeded rows) have an empty hash.
	if !user.HasUsablePassword() {
		auth.BurnPasswordCompare()
		return nil, ErrInvalidCredentials
	}

	if !auth.CheckPassword(user.PasswordHash, req.Password) {
		s.recordFailedLogin(&user)
		return nil, ErrInvalidCredentials
	}

	s.clearFailedLogins(&user)
	return s.issueTokens(&user, userAgent, ip)
}

// Refresh exchanges a refresh token for a new pair, rotating the old one.
func (s *AuthService) Refresh(plainToken, userAgent, ip string) (*models.AuthResponse, error) {
	tokenHash := auth.HashRefreshToken(plainToken)
	var response *models.AuthResponse
	var replayUserID uint

	err := database.GetDB().Transaction(func(tx *gorm.DB) error {
		var stored models.RefreshToken
		if err := tx.Where("token_hash = ?", tokenHash).First(&stored).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrInvalidRefreshToken
			}
			return fmt.Errorf("lookup refresh token: %w", err)
		}

		if stored.RevokedAt != nil {
			// This token was already rotated, so someone is replaying it. We
			// cannot tell the thief from the victim, so every live session for
			// the user has to go.
			//
			// The revocation deliberately happens AFTER this function returns:
			// returning an error rolls the transaction back, which would undo
			// any revocation performed here.
			replayUserID = stored.UserID
			return errTokenReplay
		}

		if !stored.IsActive() {
			return ErrInvalidRefreshToken
		}

		var user models.User
		if err := tx.First(&user, stored.UserID).Error; err != nil {
			return ErrInvalidRefreshToken
		}
		if !user.IsActive {
			return ErrAccountInactive
		}

		newPlain, newHash, expiresAt, err := s.tokens.GenerateRefreshToken()
		if err != nil {
			return err
		}

		if err := tx.Model(&stored).Updates(map[string]interface{}{
			"revoked_at":  time.Now(),
			"replaced_by": newHash,
		}).Error; err != nil {
			return fmt.Errorf("rotate refresh token: %w", err)
		}

		if err := tx.Create(&models.RefreshToken{
			UserID:    user.ID,
			TokenHash: newHash,
			ExpiresAt: expiresAt,
			UserAgent: truncate(userAgent, 255),
			IP:        ip,
		}).Error; err != nil {
			return fmt.Errorf("store refresh token: %w", err)
		}

		accessToken, _, err := s.tokens.GenerateAccessToken(user.ID, user.Role)
		if err != nil {
			return err
		}

		response = &models.AuthResponse{
			AccessToken:  accessToken,
			RefreshToken: newPlain,
			TokenType:    "Bearer",
			ExpiresIn:    int(s.tokens.AccessTTL().Seconds()),
			User:         models.NewUserResponse(&user),
		}
		return nil
	})

	// Runs outside the rolled-back transaction so the revocation actually sticks.
	if replayUserID != 0 {
		if revokeErr := s.LogoutAll(replayUserID); revokeErr != nil {
			log.Printf("auth: failed to revoke token family for user %d: %v", replayUserID, revokeErr)
		} else {
			log.Printf("auth: refresh token replay detected for user %d, all sessions revoked", replayUserID)
		}
	}

	if err != nil {
		// Do not tell the caller that replay specifically was detected.
		if errors.Is(err, errTokenReplay) {
			return nil, ErrInvalidRefreshToken
		}
		return nil, err
	}
	return response, nil
}

// Logout revokes a single refresh token, ending that one session.
//
// The access token stays technically valid until it expires; that is the
// inherent cost of stateless JWTs and the reason AccessTokenTTL is short.
func (s *AuthService) Logout(plainToken string) error {
	tokenHash := auth.HashRefreshToken(plainToken)
	if err := database.GetDB().Model(&models.RefreshToken{}).
		Where("token_hash = ? AND revoked_at IS NULL", tokenHash).
		Update("revoked_at", time.Now()).Error; err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}
	// An unknown or already-revoked token is not an error: logout is idempotent
	// and must not confirm whether a token was real.
	return nil
}

// LogoutAll revokes every refresh token for a user ("sign out everywhere").
func (s *AuthService) LogoutAll(userID uint) error {
	if err := database.GetDB().Model(&models.RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", time.Now()).Error; err != nil {
		return fmt.Errorf("revoke user sessions: %w", err)
	}
	return nil
}

// GetProfile returns the account for an authenticated user.
func (s *AuthService) GetProfile(userID uint) (*models.User, error) {
	var user models.User
	if err := database.GetDB().First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("lookup user: %w", err)
	}
	return &user, nil
}

// ChangePassword replaces a password after re-verifying the current one.
func (s *AuthService) ChangePassword(userID uint, req *models.ChangePasswordRequest) error {
	var user models.User
	if err := database.GetDB().First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("lookup user: %w", err)
	}

	// Re-checking the current password stops an attacker holding only a stolen
	// access token from locking the real owner out of their own account.
	if !auth.CheckPassword(user.PasswordHash, req.CurrentPassword) {
		return ErrInvalidCredentials
	}
	if req.CurrentPassword == req.NewPassword {
		return ErrSamePassword
	}
	if err := auth.ValidatePasswordStrength(req.NewPassword, user.Email, user.Name); err != nil {
		return err
	}

	hash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	if err := database.GetDB().Model(&user).Update("password_hash", hash).Error; err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	// A password change must invalidate sessions elsewhere, otherwise changing
	// it after a compromise does not actually evict the attacker.
	return s.LogoutAll(userID)
}

// issueTokens mints an access/refresh pair and records the refresh token.
func (s *AuthService) issueTokens(user *models.User, userAgent, ip string) (*models.AuthResponse, error) {
	accessToken, _, err := s.tokens.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return nil, err
	}

	plain, hash, expiresAt, err := s.tokens.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	if err := database.GetDB().Create(&models.RefreshToken{
		UserID:    user.ID,
		TokenHash: hash,
		ExpiresAt: expiresAt,
		UserAgent: truncate(userAgent, 255),
		IP:        ip,
	}).Error; err != nil {
		return nil, fmt.Errorf("store refresh token: %w", err)
	}

	return &models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: plain,
		TokenType:    "Bearer",
		ExpiresIn:    int(s.tokens.AccessTTL().Seconds()),
		User:         models.NewUserResponse(user),
	}, nil
}

// recordFailedLogin increments the failure counter and locks the account once it
// crosses the configured threshold.
func (s *AuthService) recordFailedLogin(user *models.User) {
	attempts := user.FailedLoginAttempts + 1
	updates := map[string]interface{}{"failed_login_attempts": attempts}

	if attempts >= s.cfg.MaxFailedLogins {
		updates["locked_until"] = time.Now().Add(s.cfg.LockoutDuration)
		updates["failed_login_attempts"] = 0 // reset so the next lock needs a fresh run
	}

	// Best-effort: a bookkeeping failure must not turn a wrong password into a
	// 500, since that alone would distinguish real accounts from missing ones.
	_ = database.GetDB().Model(user).Updates(updates).Error
}

// clearFailedLogins resets brute-force state after a successful login.
func (s *AuthService) clearFailedLogins(user *models.User) {
	now := time.Now()
	_ = database.GetDB().Model(user).Updates(map[string]interface{}{
		"failed_login_attempts": 0,
		"locked_until":          nil,
		"last_login_at":         now,
	}).Error
	user.LastLoginAt = &now
}

// normalizeEmail lowercases and trims an address so that "Kiran@Example.com"
// and "kiran@example.com" cannot become two separate accounts.
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// truncate caps a string at n bytes so it fits its column.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

// isDuplicateKeyError reports whether err is a unique-constraint violation.
// GORM only returns the typed ErrDuplicatedKey when TranslateError is enabled,
// so the MySQL 1062 code is checked as well.
func isDuplicateKeyError(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "1062") || strings.Contains(msg, "duplicate entry")
}
