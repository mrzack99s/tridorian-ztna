package services

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
	"tridorian-ztna/internal/models"
	"tridorian-ztna/pkg/encryption"
	"tridorian-ztna/pkg/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TenantService struct {
	db        *gorm.DB
	masterKey string
}

func NewTenantService(db *gorm.DB) *TenantService {
	return &TenantService{
		db:        db,
		masterKey: utils.GetEnv("MASTER_KEY", "default-master-key-32-chars-long"),
	}
}

const FreeDomainSuffix = ".devztna.rattanaburi.ac.th"

func (s *TenantService) ListTenants() ([]models.Tenant, error) {
	var tenants []models.Tenant
	if err := s.db.Find(&tenants).Error; err != nil {
		return nil, err
	}
	return tenants, nil
}

// FindBySlug searches for a tenant based on its slug.
func (s *TenantService) FindBySlug(slug string) (*models.Tenant, error) {
	var tenant models.Tenant
	err := s.db.First(&tenant, "slug = ?", slug).Error
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

// FindByDomain searches for a tenant based on a custom domain or free domain.
func (s *TenantService) FindByDomain(domain string) (*models.Tenant, error) {
	// 1. Check if it's a free domain (slug.devztna.rattanaburi.ac.th)
	if strings.HasSuffix(domain, FreeDomainSuffix) {
		slug := strings.TrimSuffix(domain, FreeDomainSuffix)
		return s.FindBySlug(slug)
	}

	// 2. Check CustomDomain table
	var customDomain models.CustomDomain
	err := s.db.Where("domain = ? AND is_verified = ?", domain, true).First(&customDomain).Error
	if err == nil {
		return s.GetTenantByID(customDomain.TenantID)
	}

	return nil, errors.New("tenant not found for domain: " + domain)
}

// ActivateDomain sets the primary domain for a tenant.
// The domain must be either the free domain or a verified custom domain.
func (s *TenantService) ActivateDomain(tenantID uuid.UUID, domain string) error {
	domain = strings.ToLower(strings.TrimSpace(domain))
	var tenant models.Tenant
	if err := s.db.First(&tenant, "id = ?", tenantID).Error; err != nil {
		return err
	}

	// Validate: Is it the free domain?
	freeDomain := strings.ToLower(fmt.Sprintf("%s%s", tenant.Slug, FreeDomainSuffix))
	if domain == freeDomain {
		tenant.PrimaryDomain = domain
		log.Printf("🔹 Activating free domain: %s for tenant: %s", domain, tenant.Name)
		return s.db.Save(&tenant).Error
	}

	// Validate: Is it a verified custom domain?
	var customDomain models.CustomDomain
	// Use case-sensitive match but input is lowercased.
	// Domains should generally be stored lowercased.
	err := s.db.Where("tenant_id = ? AND LOWER(domain) = ? AND is_verified = ?", tenantID, domain, true).First(&customDomain).Error
	if err != nil {
		log.Printf("❌ Domain activation failed: %s not found or not verified for tenant %s", domain, tenant.Name)
		return errors.New("domain is not verified or does not belong to this tenant")
	}

	tenant.PrimaryDomain = domain
	log.Printf("✅ Activating custom domain: %s for tenant: %s", domain, tenant.Name)
	return s.db.Save(&tenant).Error
}

// GetTenantByID retrieves a tenant by its ID.
func (s *TenantService) GetTenantByID(id uuid.UUID) (*models.Tenant, error) {
	var tenant models.Tenant
	err := s.db.First(&tenant, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

// CreateTenantWithAdmin creates a new tenant and its first local administrator.
// Returns the tenant, administrator, and the plaintext password (for one-time display).
func (s *TenantService) CreateTenantWithAdmin(name, adminEmail, adminPassword string) (*models.Tenant, *models.Administrator, string, error) {
	slug := utils.GenerateSlug(name)

	var tenant models.Tenant
	var admin models.Administrator

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 1. Create Tenant
		tenant = models.Tenant{
			Name: name,
			Slug: slug,
		}

		// Auto-generate credentials if not provided
		if adminEmail == "" {
			adminEmail = "admin@" + slug + FreeDomainSuffix
		}
		if adminPassword == "" {
			adminPassword = utils.GenerateRandomPassword(12)
		}

		if err := tx.Create(&tenant).Error; err != nil {
			return err
		}

		// Set primary domain to free domain by default
		tenant.PrimaryDomain = fmt.Sprintf("%s%s", tenant.Slug, FreeDomainSuffix)
		if err := tx.Save(&tenant).Error; err != nil {
			return err
		}

		// 2. Create Admin
		admin = models.Administrator{
			BaseTenant:             models.BaseTenant{TenantID: tenant.ID},
			Name:                   "System Administrator",
			Email:                  adminEmail,
			Role:                   models.RoleAdmin,
			ChangePasswordRequired: true,
		}
		if err := admin.SetPassword(adminPassword); err != nil {
			return err
		}

		if err := tx.Create(&admin).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, nil, "", err
	}

	return &tenant, &admin, adminPassword, nil
}

// UpdateTenant updates tenant basic info.
func (s *TenantService) UpdateTenant(id uuid.UUID, name string) error {
	return s.db.Model(&models.Tenant{}).Where("id = ?", id).Update("name", name).Error
}

// UpdateGoogleIdentity updates and encrypts Google Identity configuration
func (s *TenantService) UpdateGoogleIdentity(id uuid.UUID, clientID, clientSecret, saKey, adminEmail string) error {
	secret, err := encryption.EncryptString(clientSecret, s.masterKey)
	if err != nil {
		return err
	}
	key, err := encryption.EncryptString(saKey, s.masterKey)
	if err != nil {
		return err
	}

	return s.db.Model(&models.Tenant{}).Where("id = ?", id).Updates(map[string]interface{}{
		"google_client_id":           clientID,
		"google_client_secret":       secret,
		"google_service_account_key": key,
		"google_admin_email":         adminEmail,
	}).Error
}

// DecryptTenantConfig decrypts sensitive fields of a tenant
func (s *TenantService) DecryptTenantConfig(tenant *models.Tenant) error {
	secret, err := encryption.DecryptString(tenant.GoogleClientSecret, s.masterKey)
	if err != nil {
		return err
	}
	tenant.GoogleClientSecret = secret

	key, err := encryption.DecryptString(tenant.GoogleServiceAccountKey, s.masterKey)
	if err != nil {
		return err
	}
	tenant.GoogleServiceAccountKey = key

	return nil
}

// RegisterCustomDomain creates a record for a custom domain and generates a verification token.
func (s *TenantService) RegisterCustomDomain(tenantID uuid.UUID, domain string) (*models.CustomDomain, error) {
	tenant, err := s.GetTenantByID(tenantID)
	if err != nil {
		return nil, err
	}

	// 1. Restriction: System domains
	if domain == "auth.tridorian.com" || domain == "tridorian.com" {
		return nil, errors.New("cannot use system reserved domain")
	}

	// 2. Restriction: Free subdomain (.devztna.rattanaburi.ac.th)
	// Must be exactly tenant.Slug + FreeDomainSuffix
	expectedFreeDomain := tenant.Slug + FreeDomainSuffix
	if strings.HasSuffix(domain, FreeDomainSuffix) {
		if domain != expectedFreeDomain {
			return nil, fmt.Errorf("you can only use your assigned subdomain: %s", expectedFreeDomain)
		}
	}

	// Check if already exists
	var existing models.CustomDomain
	if err := s.db.Where("domain = ?", domain).First(&existing).Error; err == nil {
		if existing.TenantID != tenantID {
			return nil, errors.New("domain is already in use by another tenant")
		}
		// Return existing if already there (idempotent-ish)
		return &existing, nil
	}

	token := fmt.Sprintf("tridorian-verification=%s", uuid.New().String())
	isVerified := strings.HasSuffix(domain, FreeDomainSuffix)

	customDomain := models.CustomDomain{
		BaseTenant:        models.BaseTenant{TenantID: tenantID},
		Domain:            domain,
		VerificationToken: token,
		IsVerified:        isVerified,
	}

	if err := s.db.Create(&customDomain).Error; err != nil {
		return nil, err
	}

	return &customDomain, nil
}

// VerifyCustomDomain checks the DNS TXT records for the verification token.
func (s *TenantService) VerifyCustomDomain(tenantID uuid.UUID, domainID uuid.UUID) error {
	var customDomain models.CustomDomain
	if err := s.db.Where("id = ? AND tenant_id = ?", domainID, tenantID).First(&customDomain).Error; err != nil {
		return errors.New("custom domain not found")
	}

	if customDomain.IsVerified {
		return nil
	}

	// Host to check: _tridorian-challenge.<domain>
	// or just the domain itself? Usually subdomains are cleaner to avoid root clutter.
	// usage: _tridorian-challenge.example.com TXT "tridorian-verification=..."
	host := fmt.Sprintf("_tridorian-challenge.%s", customDomain.Domain)

	txtRecords, err := net.LookupTXT(host)
	if err != nil {
		// Fallback: try root domain just in case user put it there
		txtRecords, err = net.LookupTXT(customDomain.Domain)
		if err != nil {
			return errors.New("failed to lookup TXT records: " + err.Error())
		}
	}

	log.Printf("TXT records for %s: %v", host, txtRecords)

	verified := false
	for _, record := range txtRecords {
		if record == customDomain.VerificationToken {
			verified = true
			break
		}
	}

	if !verified {
		return errors.New("verification token not found in DNS records")
	}

	now := time.Now()
	customDomain.IsVerified = true
	customDomain.VerifiedAt = &now

	if err := s.db.Save(&customDomain).Error; err != nil {
		return err
	}

	// NEW: Automatically make it the primary domain upon verification
	return s.ActivateDomain(tenantID, customDomain.Domain)
}
func (s *TenantService) DeleteTenant(id uuid.UUID) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Delete related data first
		if err := tx.Where("tenant_id = ?", id).Delete(&models.Administrator{}).Error; err != nil {
			return err
		}
		if err := tx.Where("tenant_id = ?", id).Delete(&models.CustomDomain{}).Error; err != nil {
			return err
		}
		// Finally delete tenant
		return tx.Delete(&models.Tenant{}, "id = ?", id).Error
	})
}
