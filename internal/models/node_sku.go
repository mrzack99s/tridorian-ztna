package models

type NodeSku struct {
	BaseModel

	Name        string `gorm:"size:255;not null" json:"name,omitempty"`
	Description string `gorm:"type:text" json:"description,omitempty"`

	MaxUsers   int   `gorm:"not null;default:0" json:"max_users,omitempty"`
	Bandwidth  int64 `gorm:"not null;default:0" json:"bandwidth,omitempty"`
	PriceCents int64 `gorm:"not null;default:0" json:"price_cents,omitempty"` // in cents
}
