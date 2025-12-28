package models

import "time"

type CustomDomain struct {
	BaseModel
	BaseTenant

	Domain            string     `gorm:"uniqueIndex:idx_domain_tenant;not null" json:"domain,omitempty"`
	VerificationToken string     `json:"verification_token,omitempty"`
	IsVerified        bool       `gorm:"default:false" json:"is_verified,omitempty"`
	VerifiedAt        *time.Time `json:"verified_at,omitempty"`
}
