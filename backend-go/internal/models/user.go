package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID               int32          `gorm:"primaryKey;autoIncrement" json:"id"`
	Username         string         `gorm:"uniqueIndex;not null" json:"username"`
	Email            string         `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash     string         `gorm:"not null" json:"-"`
	UType            string         `gorm:"default:user" json:"utype"`
	IsActive         bool           `gorm:"default:true" json:"is_active"`
	EmailVerified    bool           `gorm:"default:false" json:"email_verified"`
	EmailVerifiedAt  *time.Time     `json:"email_verified_at,omitempty"`
	VerifyOTP        string         `json:"-"`
	VerifyOTPExpires *time.Time     `json:"-"`
	LastLogin        *time.Time     `json:"last_login,omitempty"`
	LastLoginIP      string         `json:"last_login_ip,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}
