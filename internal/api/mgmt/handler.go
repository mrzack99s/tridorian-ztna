package mgmt

import (
	"encoding/json"
	"log"
	"net/http"
	"tridorian-ztna/internal/api/common"
	"tridorian-ztna/internal/api/middleware"
	"tridorian-ztna/internal/models"
	"tridorian-ztna/internal/services"

	"github.com/google/uuid"
)

type Handler struct {
	adminService  *services.AdminService
	tenantService *services.TenantService
	policyService *services.PolicyService
	nodeService   *services.NodeService
	jwtSecret     string
}

func NewHandler(adminService *services.AdminService, tenantService *services.TenantService, policyService *services.PolicyService, nodeService *services.NodeService, jwtSecret string) *Handler {
	return &Handler{
		adminService:  adminService,
		tenantService: tenantService,
		policyService: policyService,
		nodeService:   nodeService,
		jwtSecret:     jwtSecret,
	}
}

func (h *Handler) ListTenants(w http.ResponseWriter, r *http.Request) {
	tenants, err := h.tenantService.ListTenants()
	if err != nil {
		common.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	common.Success(w, http.StatusOK, tenants)
}

func (h *Handler) ListAdmins(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	admins, err := h.adminService.ListAdmins(tenantID)
	if err != nil {
		common.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	common.Success(w, http.StatusOK, admins)
}

func (h *Handler) CreateAdmin(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		common.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	admin, err := h.adminService.CreateAdmin(tenantID, input.Name, input.Email, input.Password)
	if err != nil {
		common.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	common.Success(w, http.StatusCreated, admin)
}

func (h *Handler) DeleteAdmin(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var input struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		common.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	adminID, err := uuid.Parse(input.ID)
	if err != nil {
		common.Error(w, http.StatusBadRequest, "invalid admin id")
		return
	}

	// Prevent self-deletion?
	currentAdminIDStr := r.Context().Value(middleware.AdminIDKey).(string)
	if currentAdminIDStr == input.ID {
		common.Error(w, http.StatusBadRequest, "cannot delete yourself")
		return
	}

	if err := h.adminService.DeleteAdmin(tenantID, adminID); err != nil {
		common.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	common.Success(w, http.StatusOK, map[string]string{"message": "admin deleted"})
}

func (h *Handler) UpdateAdmin(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var input struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		common.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	adminID, err := uuid.Parse(input.ID)
	if err != nil {
		common.Error(w, http.StatusBadRequest, "invalid admin id")
		return
	}

	if err := h.adminService.UpdateAdmin(tenantID, adminID, input.Name); err != nil {
		common.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	common.Success(w, http.StatusOK, map[string]string{"message": "admin updated"})
}

func (h *Handler) CreateTenant(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name          string `json:"name"`
		AdminEmail    string `json:"admin_email"`
		AdminPassword string `json:"admin_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		common.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenant, admin, password, err := h.tenantService.CreateTenantWithAdmin(input.Name, input.AdminEmail, input.AdminPassword)
	if err != nil {
		common.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Log for visibility
	log.Printf("🏢 Created Tenant: %s (Admin: %s, Pass: %s)", tenant.Name, admin.Email, password)

	common.Success(w, http.StatusCreated, map[string]interface{}{
		"tenant":         tenant,
		"admin_email":    admin.Email,
		"admin_password": password,
	})
}

func (h *Handler) DeleteTenant(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		common.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID, err := uuid.Parse(input.ID)
	if err != nil {
		common.Error(w, http.StatusBadRequest, "invalid tenant id")
		return
	}

	if err := h.tenantService.DeleteTenant(tenantID); err != nil {
		common.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	common.Success(w, http.StatusOK, map[string]string{"message": "tenant deleted"})
}

func (h *Handler) ActivateDomain(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var input struct {
		Domain string `json:"domain"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		common.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.tenantService.ActivateDomain(tenantID, input.Domain); err != nil {
		common.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	common.Success(w, http.StatusOK, map[string]string{"message": "domain activated"})
}

func (h *Handler) UpdateIdentity(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var input struct {
		ClientID     string `json:"google_client_id"`
		ClientSecret string `json:"google_client_secret"`
		SAKey        string `json:"google_service_account_key"`
		AdminEmail   string `json:"google_admin_email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		common.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err := h.tenantService.UpdateGoogleIdentity(tenantID, input.ClientID, input.ClientSecret, input.SAKey, input.AdminEmail)
	if err != nil {
		common.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	common.Success(w, http.StatusOK, map[string]string{"message": "identity configuration updated and encrypted"})
}

func (h *Handler) UpdateTenant(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var input struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		common.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.tenantService.UpdateTenant(tenantID, input.Name); err != nil {
		common.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	common.Success(w, http.StatusOK, map[string]string{"message": "tenant updated"})
}

func (h *Handler) GetMyTenant(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	tenant, err := h.tenantService.GetTenantByID(tenantID)
	if err != nil {
		common.Error(w, http.StatusNotFound, "tenant not found")
		return
	}

	common.Success(w, http.StatusOK, tenant)
}

func (h *Handler) RegisterCustomDomain(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var input struct {
		Domain string `json:"domain"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		common.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	customDomain, err := h.tenantService.RegisterCustomDomain(tenantID, input.Domain)
	if err != nil {
		common.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	common.Success(w, http.StatusOK, customDomain)
}

func (h *Handler) VerifyCustomDomain(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var input struct {
		DomainID string `json:"domain_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		common.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	domainUUID, err := uuid.Parse(input.DomainID)
	if err != nil {
		common.Error(w, http.StatusBadRequest, "invalid domain id")
		return
	}

	if err := h.tenantService.VerifyCustomDomain(tenantID, domainUUID); err != nil {
		common.Error(w, http.StatusBadRequest, "verification failed: "+err.Error())
		return
	}
	common.Success(w, http.StatusOK, map[string]string{"message": "domain verified"})
}

func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	adminIDStr := r.Context().Value(middleware.AdminIDKey).(string)
	id, _ := uuid.Parse(adminIDStr)

	var input struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		common.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.adminService.ChangePassword(id, input.OldPassword, input.NewPassword); err != nil {
		common.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	common.Success(w, http.StatusOK, map[string]string{"message": "password changed successfully"})
}

func (h *Handler) ListAccessPolicies(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	policies, err := h.policyService.ListAccessPolicies(tenantID)
	if err != nil {
		common.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	common.Success(w, http.StatusOK, policies)
}

func (h *Handler) CreateAccessPolicy(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var input struct {
		Name     string `json:"name"`
		Effect   string `json:"effect"`
		Priority int    `json:"priority"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		common.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	policy, err := h.policyService.CreateAccessPolicy(tenantID, input.Name, input.Effect, input.Priority)
	if err != nil {
		common.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	common.Success(w, http.StatusCreated, policy)
}

func (h *Handler) DeleteAccessPolicy(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var input struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		common.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	policyID, err := uuid.Parse(input.ID)
	if err != nil {
		common.Error(w, http.StatusBadRequest, "invalid policy id")
		return
	}

	if err := h.policyService.DeleteAccessPolicy(tenantID, policyID); err != nil {
		common.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	common.Success(w, http.StatusOK, map[string]string{"message": "policy deleted"})
}

func (h *Handler) ListSignInPolicies(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	policies, err := h.policyService.ListSignInPolicies(tenantID)
	if err != nil {
		common.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	common.Success(w, http.StatusOK, policies)
}

func (h *Handler) CreateSignInPolicy(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var policy models.SignInPolicy
	if err := json.NewDecoder(r.Body).Decode(&policy); err != nil {
		common.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	created, err := h.policyService.CreateSignInPolicy(tenantID, &policy)
	if err != nil {
		common.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	common.Success(w, http.StatusCreated, created)
}

func (h *Handler) UpdateSignInPolicy(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var policy models.SignInPolicy
	if err := json.NewDecoder(r.Body).Decode(&policy); err != nil {
		common.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	updated, err := h.policyService.UpdateSignInPolicy(tenantID, &policy)
	if err != nil {
		common.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	common.Success(w, http.StatusOK, updated)
}

func (h *Handler) DeleteSignInPolicy(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var input struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		common.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	policyID, err := uuid.Parse(input.ID)
	if err != nil {
		common.Error(w, http.StatusBadRequest, "invalid policy id")
		return
	}

	if err := h.policyService.DeleteSignInPolicy(tenantID, policyID); err != nil {
		common.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	common.Success(w, http.StatusOK, map[string]string{"message": "policy deleted"})
}

func (h *Handler) ListNodes(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	nodes, err := h.nodeService.ListNodes(tenantID)
	if err != nil {
		common.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	common.Success(w, http.StatusOK, nodes)
}

func (h *Handler) CreateNode(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var input struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		common.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	node, err := h.nodeService.CreateNode(tenantID, input.Name)
	if err != nil {
		common.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	common.Success(w, http.StatusCreated, node)
}

func (h *Handler) DeleteNode(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var input struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		common.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	nodeID, err := uuid.Parse(input.ID)
	if err != nil {
		common.Error(w, http.StatusBadRequest, "invalid node id")
		return
	}

	if err := h.nodeService.DeleteNode(tenantID, nodeID); err != nil {
		common.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	common.Success(w, http.StatusOK, map[string]string{"message": "node deleted"})
}
