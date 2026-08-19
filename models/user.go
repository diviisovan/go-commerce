package models

import (
	"time"

	"gorm.io/gorm"
)

// User roles. Role is stored as a plain string so it survives schema changes.
const (
	RoleCustomer = "customer"
	RoleAdmin    = "admin"
)

// User represents a customer in the eCommerce system
// @Description User account information
type User struct {
	ID        uint           `gorm:"primaryKey" json:"id" example:"1"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Name      string         `gorm:"type:varchar(255);not null" json:"name" example:"Kiran Kumar"`
	Email     string         `gorm:"type:varchar(191);uniqueIndex;not null" json:"email" example:"kiran@example.com"`
	Address   string         `gorm:"type:text" json:"address"`
	Phone     string         `gorm:"type:varchar(50)" json:"phone"`
	IsActive  bool           `gorm:"default:true" json:"is_active"`

	// PasswordHash holds the bcrypt hash of the password, never the password itself.
	// The json:"-" tag is what keeps it out of every API response — do not remove it.
	PasswordHash string `gorm:"type:varchar(255);not null;default:''" json:"-"`

	// Role drives authorization checks in middleware.RequireRole.
	Role string `gorm:"type:varchar(20);not null;default:'customer'" json:"role" example:"customer"`

	LastLoginAt *time.Time `json:"last_login_at,omitempty"`

	// Brute-force protection state. Both are hidden from JSON so they never leak
	// how close an attacker is to a lockout.
	FailedLoginAttempts int        `gorm:"not null;default:0" json:"-"`
	LockedUntil         *time.Time `json:"-"`
}

// TableName specifies the table name for User
func (User) TableName() string {
	return "users"
}

// IsLocked reports whether the account is currently locked out after too many
// failed login attempts.
func (u *User) IsLocked() bool {
	return u.LockedUntil != nil && u.LockedUntil.After(time.Now())
}

// HasUsablePassword reports whether the account can be logged into with a
// password. Accounts created before auth existed (e.g. seeded users) have an
// empty hash and must go through a password reset instead.
func (u *User) HasUsablePassword() bool {
	return u.PasswordHash != ""
}
