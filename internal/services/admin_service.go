package services

import (
	"errors"
	"tridorian-ztna/internal/models"
	"tridorian-ztna/pkg/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AdminService struct {
	db *gorm.DB
}

func NewAdminService(db *gorm.DB) *AdminService {
	return &AdminService{db: db}
}

func (s *AdminService) ListAdmins(tenantID uuid.UUID) ([]models.Administrator, error) {
	var admins []models.Administrator
	if err := s.db.Scopes(models.TenantScope(tenantID)).Find(&admins).Error; err != nil {
		return nil, err
	}
	return admins, nil
}

func (s *AdminService) Authenticate(tenantID uuid.UUID, email, password string) (*models.Administrator, error) {
	var admin models.Administrator
	err := s.db.Scopes(models.TenantScope(tenantID)).Where("email = ?", email).First(&admin).Error
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if !admin.CheckPassword(password) {
		return nil, errors.New("invalid email or password")
	}

	return &admin, nil
}
func (s *AdminService) GetByID(id uuid.UUID) (*models.Administrator, error) {
	var admin models.Administrator
	err := s.db.First(&admin, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

func (s *AdminService) ChangePassword(id uuid.UUID, oldPassword, newPassword string) error {
	var admin models.Administrator
	if err := s.db.First(&admin, "id = ?", id).Error; err != nil {
		return err
	}

	if !admin.CheckPassword(oldPassword) {
		return errors.New("incorrect old password")
	}

	if err := admin.SetPassword(newPassword); err != nil {
		return err
	}

	// Reset flag
	admin.ChangePasswordRequired = false
	return s.db.Save(&admin).Error
}
func (s *AdminService) CreateAdmin(tenantID uuid.UUID, name, email, password string, role models.AdminRole) (*models.Administrator, string, error) {
	admin := &models.Administrator{
		BaseTenant: models.BaseTenant{TenantID: tenantID},
		Name:       name,
		Email:      email,
		Role:       role,
	}

	if password == "" {
		password = utils.GenerateRandomPassword(12)
	}

	if err := admin.SetPassword(password); err != nil {
		return nil, "", err
	}

	if err := s.db.Create(admin).Error; err != nil {
		return nil, "", err
	}

	return admin, password, nil
}

func (s *AdminService) DeleteAdmin(tenantID uuid.UUID, adminID uuid.UUID) error {
	// Don't allow deleting the last admin? Or maybe just check if it exists in this tenant.
	return s.db.Scopes(models.TenantScope(tenantID)).Delete(&models.Administrator{}, "id = ?", adminID).Error
}

func (s *AdminService) UpdateAdmin(tenantID uuid.UUID, adminID uuid.UUID, name string, role models.AdminRole) error {
	updates := map[string]interface{}{
		"name": name,
	}
	if role != "" {
		updates["role"] = role
	}
	return s.db.Scopes(models.TenantScope(tenantID)).Model(&models.Administrator{}).Where("id = ?", adminID).Updates(updates).Error
}
