package mgmt

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"tridorian-ztna/internal/api/common"
	"tridorian-ztna/internal/api/middleware"
	"tridorian-ztna/internal/models"
	"tridorian-ztna/internal/services"

	"github.com/google/uuid"
)

type Handler struct {
	adminService       *services.AdminService
	tenantService      *services.TenantService
	policyService      *services.PolicyService
	nodeService        *services.NodeService
	identityService    *services.IdentityService
	applicationService *services.ApplicationService
}

func NewHandler(adminService *services.AdminService, tenantService *services.TenantService, policyService *services.PolicyService, nodeService *services.NodeService, identityService *services.IdentityService, applicationService *services.ApplicationService) *Handler {
	return &Handler{
		adminService:       adminService,
		tenantService:      tenantService,
		policyService:      policyService,
		nodeService:        nodeService,
		identityService:    identityService,
		applicationService: applicationService,
	}
}

func (h *Handler) ListTenants(w http.ResponseWriter, r *http.Request) {
	tenants, err := h.tenantService.ListTenants()
	if err != nil {
		common.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	common.Success(w, http.StatusOK, map[string]interface{}{
		"tenants":            tenants,
		"free_domain_suffix": h.tenantService.GetFreeDomainSuffix(),
	})
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
		Name     string           `json:"name"`
		Email    string           `json:"email"`
		Password string           `json:"password"`
		Role     models.AdminRole `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		common.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	admin, password, err := h.adminService.CreateAdmin(tenantID, input.Name, input.Email, input.Password, input.Role)
	if err != nil {
		common.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	common.Success(w, http.StatusCreated, map[string]interface{}{
		"admin":    admin,
		"password": password,
	})
}

func (h *Handler) ListDomains(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	domains, err := h.tenantService.ListDomains(tenantID)
	if err != nil {
		common.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	common.Success(w, http.StatusOK, domains)
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
		ID   string           `json:"id"`
		Name string           `json:"name"`
		Role models.AdminRole `json:"role"`
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

	if err := h.adminService.UpdateAdmin(tenantID, adminID, input.Name, input.Role); err != nil {
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
		"tenant":             tenant,
		"admin_email":        admin.Email,
		"admin_password":     password,
		"free_domain_suffix": h.tenantService.GetFreeDomainSuffix(),
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

	// Use a map to include the global free domain suffix
	response := struct {
		*models.Tenant
		FreeDomainSuffix string `json:"free_domain_suffix"`
	}{
		Tenant:           tenant,
		FreeDomainSuffix: h.tenantService.GetFreeDomainSuffix(),
	}

	common.Success(w, http.StatusOK, response)
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
		models.AccessPolicy
		RawRootNodeID       *string  `json:"root_node_id"`
		RawDestinationAppID *string  `json:"destination_app_id"`
		NodeIDs             []string `json:"node_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		common.Error(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	policy := input.AccessPolicy

	// Handle RootNodeID
	if input.RawRootNodeID != nil && *input.RawRootNodeID != "" {
		parsedID, err := uuid.Parse(*input.RawRootNodeID)
		if err != nil {
			common.Error(w, http.StatusBadRequest, "invalid root_node_id format")
			return
		}
		policy.RootNodeID = &parsedID
	}

	// Handle DestinationAppID
	if input.RawDestinationAppID != nil && *input.RawDestinationAppID != "" {
		parsedID, err := uuid.Parse(*input.RawDestinationAppID)
		if err != nil {
			common.Error(w, http.StatusBadRequest, "invalid destination_app_id format")
			return
		}
		policy.DestinationAppID = &parsedID
	}

	// Handle NodeIDs allocation
	if len(input.NodeIDs) > 0 {
		var nodes []models.Node
		for _, idStr := range input.NodeIDs {
			uid, err := uuid.Parse(idStr)
			if err == nil {
				// We only need the ID for the association
				nodes = append(nodes, models.Node{BaseModel: models.BaseModel{ID: uid}})
			}
		}
		policy.Nodes = nodes
	}

	created, err := h.policyService.CreateAccessPolicy(tenantID, &policy)
	if err != nil {
		common.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	common.Success(w, http.StatusCreated, created)
}

func (h *Handler) UpdateAccessPolicy(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var input struct {
		models.AccessPolicy
		RawRootNodeID       *string  `json:"root_node_id"`
		RawDestinationAppID *string  `json:"destination_app_id"`
		NodeIDs             []string `json:"node_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		common.Error(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	policy := input.AccessPolicy

	// Handle RootNodeID
	if input.RawRootNodeID != nil && *input.RawRootNodeID != "" {
		parsedID, err := uuid.Parse(*input.RawRootNodeID)
		if err != nil {
			common.Error(w, http.StatusBadRequest, "invalid root_node_id format")
			return
		}
		policy.RootNodeID = &parsedID
	}

	// Handle DestinationAppID
	if input.RawDestinationAppID != nil && *input.RawDestinationAppID != "" {
		parsedID, err := uuid.Parse(*input.RawDestinationAppID)
		if err != nil {
			common.Error(w, http.StatusBadRequest, "invalid destination_app_id format")
			return
		}
		policy.DestinationAppID = &parsedID
	}

	// Handle NodeIDs allocation
	if len(input.NodeIDs) > 0 {
		var nodes []models.Node
		for _, idStr := range input.NodeIDs {
			uid, err := uuid.Parse(idStr)
			if err == nil {
				nodes = append(nodes, models.Node{BaseModel: models.BaseModel{ID: uid}})
			}
		}
		policy.Nodes = nodes
	}

	updated, err := h.policyService.UpdateAccessPolicy(tenantID, &policy)
	if err != nil {
		common.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	common.Success(w, http.StatusOK, updated)
}

func (h *Handler) DeleteAccessPolicy(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.URL.Query().Get("id")
	if id == "" {
		common.Error(w, http.StatusBadRequest, "id is required")
		return
	}

	policyID, err := uuid.Parse(id)
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
	id := r.URL.Query().Get("id")
	if id == "" {
		common.Error(w, http.StatusBadRequest, "id is required")
		return
	}
	policyID, err := uuid.Parse(id)
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

func (h *Handler) ListNodeSkus(w http.ResponseWriter, r *http.Request) {
	skus, err := h.nodeService.ListNodeSkus()
	if err != nil {
		common.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	common.Success(w, http.StatusOK, skus)
}

func (h *Handler) ListNodeSessions(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	nodeIDStr := r.URL.Query().Get("node_id")
	if nodeIDStr == "" {
		common.Error(w, http.StatusBadRequest, "node_id is required")
		return
	}

	nodeID, err := uuid.Parse(nodeIDStr)
	if err != nil {
		common.Error(w, http.StatusBadRequest, "invalid node_id")
		return
	}

	sessions, err := h.nodeService.ListSessions(tenantID, nodeID)
	if err != nil {
		common.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	common.Success(w, http.StatusOK, sessions)
}

func (h *Handler) CreateNode(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var input struct {
		Name       string `json:"name"`
		SkuID      string `json:"sku_id"`
		ClientCIDR string `json:"client_cidr"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		common.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if input.SkuID == "" {
		common.Error(w, http.StatusBadRequest, "sku_id is required")
		return
	}

	skuUUID, err := uuid.Parse(input.SkuID)
	if err != nil {
		common.Error(w, http.StatusBadRequest, "invalid sku_id")
		return
	}

	node, err := h.nodeService.CreateNode(tenantID, input.Name, skuUUID, input.ClientCIDR)
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
func (h *Handler) SearchIdentity(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	query := r.URL.Query().Get("q")
	if query == "" {
		common.Error(w, http.StatusBadRequest, "missing query parameter q")
		return
	}

	tenant, err := h.tenantService.GetTenantByID(tenantID)
	if err != nil {
		common.Error(w, http.StatusNotFound, "tenant not found")
		return
	}

	if tenant.GoogleServiceAccountKey == "" || tenant.GoogleAdminEmail == "" {
		common.Error(w, http.StatusBadRequest, "google identity not configured for this tenant")
		return
	}

	// If it doesn't look like JSON, try to decrypt it
	if !strings.HasPrefix(strings.TrimSpace(tenant.GoogleServiceAccountKey), "{") {
		if err := h.tenantService.DecryptTenantConfig(tenant); err != nil {
			log.Printf("failed to decrypt identity configuration for tenant %s: %v", tenantID, err)
			common.Error(w, http.StatusInternalServerError, "failed to decrypt identity configuration. please re-configure google identity.")
			return
		}
	}

	// Double check it's JSON now
	if !strings.HasPrefix(strings.TrimSpace(tenant.GoogleServiceAccountKey), "{") {
		log.Printf("identity configuration for tenant %s is not a valid JSON after decryption", tenantID)
		common.Error(w, http.StatusBadRequest, "invalid identity configuration format. please re-upload your service account key.")
		return
	}

	// Search Users - handle errors gracefully
	users, err := h.identityService.SearchGoogleUsers(r.Context(), []byte(tenant.GoogleServiceAccountKey), tenant.GoogleAdminEmail, query)
	if err != nil {
		log.Printf("failed to search users for tenant %s: %v", tenantID, err)
		users = []models.ExternalIdentity{} // Return empty slice instead of nil
	}

	// Search Groups - handle errors gracefully
	groups, err := h.identityService.SearchGoogleGroups(r.Context(), []byte(tenant.GoogleServiceAccountKey), tenant.GoogleAdminEmail, query)
	if err != nil {
		log.Printf("failed to search groups for tenant %s: %v", tenantID, err)
		groups = []string{} // Return empty slice instead of nil
	}

	var results = []struct {
		Type  string `json:"type"`
		Value string `json:"value"`
		Label string `json:"label"`
	}{}

	for _, u := range users {
		results = append(results, struct {
			Type  string `json:"type"`
			Value string `json:"value"`
			Label string `json:"label"`
		}{
			Type:  "user",
			Value: u.Email,
			Label: fmt.Sprintf("%s (%s)", u.Name, u.Email),
		})
	}

	for _, g := range groups {
		results = append(results, struct {
			Type  string `json:"type"`
			Value string `json:"value"`
			Label string `json:"label"`
		}{
			Type:  "group",
			Value: g,
			Label: g,
		})
	}

	// Always return success with results (even if empty)
	// This prevents white screen errors on the frontend

	common.Success(w, http.StatusOK, results)
}

// Application Management

func (h *Handler) ListApplications(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	apps, err := h.applicationService.ListApplications(tenantID)
	if err != nil {
		common.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	// Ensure we always return an array, never null
	if apps == nil {
		apps = []models.Application{}
	}
	common.Success(w, http.StatusOK, apps)
}

func (h *Handler) GetApplication(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	appIDStr := r.URL.Query().Get("id")
	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		common.Error(w, http.StatusBadRequest, "invalid application id")
		return
	}

	app, cidrs, err := h.applicationService.GetApplicationWithCIDRs(tenantID, appID)
	if err != nil {
		common.Error(w, http.StatusNotFound, "application not found")
		return
	}

	common.Success(w, http.StatusOK, map[string]interface{}{
		"application": app,
		"cidrs":       cidrs,
	})
}

func (h *Handler) CreateApplication(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var input struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		CIDRs       []string `json:"cidrs"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		common.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	app, err := h.applicationService.CreateApplication(tenantID, input.Name, input.Description, input.CIDRs)
	if err != nil {
		common.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	common.Success(w, http.StatusCreated, app)
}

func (h *Handler) UpdateApplication(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var input struct {
		ID          string   `json:"id"`
		Name        string   `json:"name"`
		Description string   `json:"description"`
		CIDRs       []string `json:"cidrs"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		common.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	appID, err := uuid.Parse(input.ID)
	if err != nil {
		common.Error(w, http.StatusBadRequest, "invalid application id")
		return
	}

	if err := h.applicationService.UpdateApplication(tenantID, appID, input.Name, input.Description, input.CIDRs); err != nil {
		common.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	common.Success(w, http.StatusOK, map[string]string{"message": "application updated"})
}

func (h *Handler) DeleteApplication(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var input struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		common.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	appID, err := uuid.Parse(input.ID)
	if err != nil {
		common.Error(w, http.StatusBadRequest, "invalid application id")
		return
	}

	if err := h.applicationService.DeleteApplication(tenantID, appID); err != nil {
		common.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	common.Success(w, http.StatusOK, map[string]string{"message": "application deleted"})
}
