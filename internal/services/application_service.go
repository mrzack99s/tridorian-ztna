package services

import (
	"errors"
	"tridorian-ztna/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ApplicationService struct {
	db *gorm.DB
}

func NewApplicationService(db *gorm.DB) *ApplicationService {
	return &ApplicationService{db: db}
}

func (s *ApplicationService) ListApplications(tenantID uuid.UUID) ([]models.Application, error) {
	var apps []models.Application
	if err := s.db.Scopes(models.TenantScope(tenantID)).Find(&apps).Error; err != nil {
		return nil, err
	}
	return apps, nil
}

func (s *ApplicationService) GetApplicationWithCIDRs(tenantID uuid.UUID, appID uuid.UUID) (*models.Application, []models.ApplicationCIDR, error) {
	var app models.Application
	if err := s.db.Scopes(models.TenantScope(tenantID)).First(&app, "id = ?", appID).Error; err != nil {
		return nil, nil, err
	}

	var cidrs []models.ApplicationCIDR
	if err := s.db.Where("application_id = ?", appID).Find(&cidrs).Error; err != nil {
		return nil, nil, err
	}

	return &app, cidrs, nil
}

func (s *ApplicationService) CreateApplication(tenantID uuid.UUID, name, description string, cidrs []string) (*models.Application, error) {
	app := &models.Application{
		BaseTenant:  models.BaseTenant{TenantID: tenantID},
		Name:        name,
		Description: description,
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(app).Error; err != nil {
			return err
		}

		for _, cidr := range cidrs {
			appCIDR := &models.ApplicationCIDR{
				BaseTenant:    models.BaseTenant{TenantID: tenantID},
				ApplicationID: app.ID,
				CIDR:          cidr,
			}
			if err := tx.Create(appCIDR).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return app, nil
}

func (s *ApplicationService) UpdateApplication(tenantID uuid.UUID, appID uuid.UUID, name, description string, cidrs []string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Update application
		if err := tx.Scopes(models.TenantScope(tenantID)).
			Model(&models.Application{}).
			Where("id = ?", appID).
			Updates(map[string]interface{}{
				"name":        name,
				"description": description,
			}).Error; err != nil {
			return err
		}

		// Delete old CIDRs
		if err := tx.Where("application_id = ?", appID).Delete(&models.ApplicationCIDR{}).Error; err != nil {
			return err
		}

		// Create new CIDRs
		for _, cidr := range cidrs {
			appCIDR := &models.ApplicationCIDR{
				BaseTenant:    models.BaseTenant{TenantID: tenantID},
				ApplicationID: appID,
				CIDR:          cidr,
			}
			if err := tx.Create(appCIDR).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *ApplicationService) DeleteApplication(tenantID uuid.UUID, appID uuid.UUID) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Delete CIDRs first
		if err := tx.Where("application_id = ?", appID).Delete(&models.ApplicationCIDR{}).Error; err != nil {
			return err
		}

		// Delete application
		result := tx.Scopes(models.TenantScope(tenantID)).Delete(&models.Application{}, "id = ?", appID)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("application not found")
		}

		return nil
	})
}
