// Package auth holds the security primitives shared by the auth service and the
// authentication middleware: password hashing and JWT / refresh-token handling.
package auth

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

const (
	// BcryptCost of 12 is roughly 250ms per hash on current hardware. High
	// enough to make offline cracking expensive, low enough to keep login
	// responsive. Raise it as hardware improves.
	BcryptCost = 12

	// MinPasswordLength follows NIST SP 800-63B, which recommends a length
	// floor over forced composition rules.
	MinPasswordLength = 8

	// MaxPasswordBytes is a hard limit of the algorithm, not a policy choice:
	// bcrypt only considers the first 72 bytes, so anything longer would be
	// silently truncated. Rejecting is safer than truncating.
	MaxPasswordBytes = 72
)

var (
	ErrPasswordTooShort  = fmt.Errorf("password must be at least %d characters", MinPasswordLength)
	ErrPasswordTooLong   = fmt.Errorf("password must be at most %d bytes", MaxPasswordBytes)
	ErrPasswordTooCommon = errors.New("password is too common, please choose another")
	ErrPasswordTooSimple = errors.New("password must contain both letters and at least one number or symbol")
	ErrPasswordEchoesID  = errors.New("password must not contain your name or email")
)

// commonPasswords blocks the credentials that show up first in every breach
// corpus. A production system should check against a full list such as the
// Pwned Passwords k-anonymity API instead of an inline slice.
var commonPasswords = map[string]struct{}{
	"password": {}, "password1": {}, "password123": {}, "12345678": {},
	"123456789": {}, "1234567890": {}, "qwerty123": {}, "qwertyuiop": {},
	"letmein": {}, "welcome1": {}, "admin123": {}, "iloveyou": {},
	"sunshine": {}, "princess": {}, "football": {}, "baseball": {},
	"passw0rd": {}, "p@ssw0rd": {}, "abc12345": {}, "changeme": {},
}

// HashPassword returns the bcrypt hash of a plaintext password.
func HashPassword(plain string) (string, error) {
	if len(plain) > MaxPasswordBytes {
		return "", ErrPasswordTooLong
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), BcryptCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

// CheckPassword reports whether plain matches the stored bcrypt hash.
// bcrypt's comparison is constant-time with respect to the hash contents.
func CheckPassword(hash, plain string) bool {
	if hash == "" {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

var (
	dummyOnce sync.Once
	dummyHash []byte
)

// BurnPasswordCompare performs a throwaway bcrypt comparison.
//
// Call it on the "user not found" path of login. Without it, a missing account
// returns in microseconds while a real one takes ~250ms, and that timing gap
// alone lets an attacker enumerate which email addresses are registered.
func BurnPasswordCompare() {
	dummyOnce.Do(func() {
		dummyHash, _ = bcrypt.GenerateFromPassword([]byte("timing-equalizer"), BcryptCost)
	})
	_ = bcrypt.CompareHashAndPassword(dummyHash, []byte("timing-equalizer-miss"))
}

// ValidatePasswordStrength enforces the password policy. The identifiers
// argument takes context-specific strings (name, email) that the password must
// not simply echo back.
func ValidatePasswordStrength(password string, identifiers ...string) error {
	if len([]rune(password)) < MinPasswordLength {
		return ErrPasswordTooShort
	}
	// Measured in bytes, not runes: a 20-character password of multi-byte
	// characters can still exceed bcrypt's 72-byte window.
	if len(password) > MaxPasswordBytes {
		return ErrPasswordTooLong
	}

	lower := strings.ToLower(password)
	if _, blocked := commonPasswords[lower]; blocked {
		return ErrPasswordTooCommon
	}

	var hasLetter, hasOther bool
	for _, r := range password {
		switch {
		case unicode.IsLetter(r):
			hasLetter = true
		case unicode.IsDigit(r), unicode.IsPunct(r), unicode.IsSymbol(r), unicode.IsSpace(r):
			hasOther = true
		}
	}
	if !hasLetter || !hasOther {
		return ErrPasswordTooSimple
	}

	for _, word := range identifierWords(identifiers) {
		// Both directions matter: "kiranecho123" contains the word "kiranecho",
		// and the password "kiran" is contained in the word "kirankumar".
		if strings.Contains(lower, word) || strings.Contains(word, lower) {
			return ErrPasswordEchoesID
		}
	}

	return nil
}

// identifierWords breaks names and email addresses into the comparable words a
// password must not echo.
//
// A whole-string comparison is far too easy to slip past: the email
// "kiran166@example.com" has the local part "kiran166", which the password
// "kiran166!" contains but "kiranSecret1" does not, even though both give away
// the same name. Splitting on separators and stripping trailing digits reduces
// each identifier to its meaningful stem.
func identifierWords(identifiers []string) []string {
	var words []string

	for _, id := range identifiers {
		id = strings.ToLower(strings.TrimSpace(id))
		if at := strings.IndexByte(id, '@'); at > 0 {
			id = id[:at] // an email's domain is not personal information
		}

		for _, part := range strings.FieldsFunc(id, func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsDigit(r)
		}) {
			part = strings.TrimRight(part, "0123456789")
			// Words shorter than 4 characters produce too many false positives
			// to be worth blocking.
			if len(part) >= 4 {
				words = append(words, part)
			}
		}
	}

	return words
}
