package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BaseTenant struct {
	TenantID uuid.UUID `gorm:"type:uuid;index;not null" json:"tenant_id,omitempty"`
}

// TenantScope returns a GORM scope that filters results by TenantID.
func TenantScope(tenantID uuid.UUID) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("tenant_id = ?", tenantID)
	}
}

type BaseModel struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id,omitempty"`
	CreatedAt time.Time      `json:"created_at,omitempty"`
	UpdatedAt time.Time      `json:"updated_at,omitempty"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

type BasePolicy struct {
	BaseModel

	Name     string `json:"name,omitempty"`
	Priority int    `json:"priority"`
	Enabled  bool   `json:"enabled"`
}
