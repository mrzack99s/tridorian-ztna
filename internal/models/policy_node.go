package models

import (
	"time"

	"github.com/google/uuid"
)

type PolicyNode struct {
	BaseModel
	BaseTenant

	// AND / OR
	Operator string `gorm:"size:10" json:"operator,omitempty"` // "AND" | "OR"

	// Self reference (Tree)
	ParentID *uuid.UUID   `json:"parent_id,omitempty"`
	Parent   *PolicyNode  `json:"parent,omitempty"`
	Children []PolicyNode `gorm:"foreignKey:ParentID" json:"children,omitempty"`

	// ถ้าเป็น leaf
	Condition *PolicyCondition `gorm:"foreignKey:NodeID" json:"condition,omitempty"`

	CreatedAt time.Time `json:"created_at,omitempty"`
}
