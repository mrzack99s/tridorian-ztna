package models

import "github.com/google/uuid"

type Application struct {
	BaseModel
	BaseTenant

	Name        string            `gorm:"size:255;not null" json:"name,omitempty"`
	Description string            `gorm:"size:500" json:"description,omitempty"`
	CIDRs       []ApplicationCIDR `gorm:"foreignKey:ApplicationID" json:"cidrs,omitempty"`
}

type ApplicationCIDR struct {
	BaseModel
	BaseTenant

	ApplicationID uuid.UUID `gorm:"type:uuid;not null;index" json:"application_id,omitempty"`
	CIDR          string    `gorm:"size:100;not null;column:cidr" json:"cidr,omitempty"`
}

func (ApplicationCIDR) TableName() string {
	return "application_cidrs"
}
