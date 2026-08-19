package models

import "time"

// RefreshToken is a long-lived credential used to mint new access tokens.
//
// Only the SHA-256 hash of the token is stored: a leaked database dump then
// cannot be replayed against the API. Tokens are rotated on every use, and a
// reused token is treated as theft, which revokes the whole family.
type RefreshToken struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`

	UserID uint `gorm:"not null;index" json:"user_id"`
	User   User `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`

	// TokenHash is the hex-encoded SHA-256 of the opaque token string.
	TokenHash string `gorm:"type:char(64);uniqueIndex;not null" json:"-"`

	ExpiresAt time.Time  `gorm:"not null;index" json:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`

	// ReplacedBy links a rotated token to its successor, so a reuse attempt can
	// be traced back through the chain.
	ReplacedBy string `gorm:"type:char(64)" json:"-"`

	UserAgent string `gorm:"type:varchar(255)" json:"-"`
	IP        string `gorm:"type:varchar(45)" json:"-"`
}

// TableName specifies the table name for RefreshToken
func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

// IsActive reports whether the token can still be exchanged.
func (t *RefreshToken) IsActive() bool {
	return t.RevokedAt == nil && t.ExpiresAt.After(time.Now())
}
