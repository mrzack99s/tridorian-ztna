package models

type Tenant struct {
	BaseModel
	Name          string `json:"name,omitempty"`
	Slug          string `gorm:"uniqueIndex;not null" json:"slug,omitempty"`
	PrimaryDomain string `json:"primary_domain,omitempty"` // The domain currently chosen as active

	// Google Identity Configuration (Per Tenant)
	GoogleClientID          string `json:"google_client_id,omitempty"`
	GoogleClientSecret      string `json:"-"`                  // sensitive
	GoogleServiceAccountKey string `gorm:"type:text" json:"-"` // sensitive JSON
	GoogleAdminEmail        string `json:"google_admin_email,omitempty"`
}
