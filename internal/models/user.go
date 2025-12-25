package models

import "golang.org/x/crypto/bcrypt"

type AdminRole string

const (
	RoleAdmin      AdminRole = "admin"
	RoleSuperAdmin AdminRole = "super_admin"
)

// Administrator is a local account for managing the system.
// These are created and stored in our database.
type Administrator struct {
	BaseModel
	BaseTenant

	Name                   string    `json:"name,omitempty"`
	Email                  string    `gorm:"uniqueIndex:idx_admin_email_tenant;not null" json:"email,omitempty"`
	Password               string    `json:"-"` // Locally managed password
	ChangePasswordRequired bool      `gorm:"default:true" json:"change_password_required"`
	Role                   AdminRole `gorm:"type:varchar(20);default:'admin'" json:"role,omitempty"`
}

func (a *Administrator) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	a.Password = string(hashedPassword)
	return nil
}

func (a *Administrator) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(a.Password), []byte(password))
	return err == nil
}

// ExternalIdentity represents a user from Google Identity.
// NOT stored in our database. Used only for authorization logic.
type ExternalIdentity struct {
	Name       string `json:"name"`
	Email      string `json:"email"`
	ExternalID string `json:"external_id"`
	IsAdmin    bool   `json:"is_admin"` // Google Workspace Admin status
}

// BackofficeUser is for the system-wide administration.
// SEPARATE table from tenant administrators.
type BackofficeUser struct {
	BaseModel

	Name     string `json:"name,omitempty"`
	Email    string `gorm:"uniqueIndex;not null" json:"email,omitempty"`
	Password string `json:"-"`
}

func (u *BackofficeUser) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

func (u *BackofficeUser) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
