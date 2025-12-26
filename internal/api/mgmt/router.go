package mgmt

import (
	"net/http"
	"strings"
	"tridorian-ztna/internal/api/middleware"
	"tridorian-ztna/internal/services"
	"tridorian-ztna/pkg/utils"

	"gorm.io/gorm"
)

type Router struct {
	handler   *Handler
	jwtSecret string
}

func NewRouter(db *gorm.DB, jwtSecret string) *Router {
	adminService := services.NewAdminService(db)
	tenantService := services.NewTenantService(db)
	policyService := services.NewPolicyService(db)
	nodeService := services.NewNodeService(db)
	identityService := services.NewIdentityService()

	return &Router{
		handler:   NewHandler(adminService, tenantService, policyService, nodeService, identityService, jwtSecret),
		jwtSecret: jwtSecret,
	}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path

	// Public Health Check
	if path == "/health" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
		return
	}

	// 1. Backoffice Routes (System Admin)
	if path == "/api/v1/tenants" {
		middleware.JWTAuth(r.jwtSecret, utils.PurposeBackoffice)(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if req.Method == "GET" {
				r.handler.ListTenants(w, req)
			} else if req.Method == "POST" {
				r.handler.CreateTenant(w, req)
			} else if req.Method == "DELETE" {
				r.handler.DeleteTenant(w, req)
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		})).ServeHTTP(w, req)
		return
	}

	// 2. Tenant Management Routes (Tenant Admin)
	// These routes require PurposeManagement token which contains the TenantID
	tenantRoutes := []string{
		"/api/v1/tenants/activate",
		"/api/v1/tenants/identity",
		"/api/v1/admins",
		"/api/v1/tenant/me",
		"/api/v1/tenants/domains", // Includes /verify
		"/api/v1/profile/change-password",
		"/api/v1/policies/access",
		"/api/v1/policies/sign-in",
		"/api/v1/nodes",
		"/api/v1/identity/search",
	}

	isTenantRoute := false
	for _, route := range tenantRoutes {
		if strings.HasPrefix(path, route) {
			isTenantRoute = true
			break
		}
	}

	if isTenantRoute {
		middleware.JWTAuth(r.jwtSecret, utils.PurposeManagement)(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			switch {
			case path == "/api/v1/tenants/activate" && req.Method == "POST":
				r.handler.ActivateDomain(w, req)

			case path == "/api/v1/tenants/identity" && req.Method == "POST":
				r.handler.UpdateIdentity(w, req)

			case path == "/api/v1/tenant/me" && req.Method == "GET":
				r.handler.GetMyTenant(w, req)

			case path == "/api/v1/tenant/me" && req.Method == "PATCH":
				r.handler.UpdateTenant(w, req)

			// Custom Domain Management
			case path == "/api/v1/tenants/domains" && req.Method == "POST":
				r.handler.RegisterCustomDomain(w, req)

			case path == "/api/v1/tenants/domains/verify" && req.Method == "POST":
				r.handler.VerifyCustomDomain(w, req)

			case path == "/api/v1/profile/change-password" && req.Method == "POST":
				r.handler.ChangePassword(w, req)

			case strings.HasPrefix(path, "/api/v1/admins"):
				if req.Method == "GET" {
					r.handler.ListAdmins(w, req)
				} else if req.Method == "POST" {
					r.handler.CreateAdmin(w, req)
				} else if req.Method == "PATCH" {
					r.handler.UpdateAdmin(w, req)
				} else if req.Method == "DELETE" {
					r.handler.DeleteAdmin(w, req)
				} else {
					w.WriteHeader(http.StatusMethodNotAllowed)
				}

			// Policy Management
			case path == "/api/v1/policies/access":
				if req.Method == "GET" {
					r.handler.ListAccessPolicies(w, req)
				} else if req.Method == "POST" {
					r.handler.CreateAccessPolicy(w, req)
				} else if req.Method == "PATCH" {
					r.handler.UpdateAccessPolicy(w, req)
				} else if req.Method == "DELETE" {
					r.handler.DeleteAccessPolicy(w, req)
				}

			case path == "/api/v1/policies/sign-in":
				if req.Method == "GET" {
					r.handler.ListSignInPolicies(w, req)
				} else if req.Method == "POST" {
					r.handler.CreateSignInPolicy(w, req)
				} else if req.Method == "PATCH" {
					r.handler.UpdateSignInPolicy(w, req)
				} else if req.Method == "DELETE" {
					r.handler.DeleteSignInPolicy(w, req)
				}

			// Node Management
			case path == "/api/v1/nodes":
				if req.Method == "GET" {
					r.handler.ListNodes(w, req)
				} else if req.Method == "POST" {
					r.handler.CreateNode(w, req)
				} else if req.Method == "DELETE" {
					r.handler.DeleteNode(w, req)
				}

			// Identity Management
			case path == "/api/v1/identity/search" && req.Method == "GET":
				r.handler.SearchIdentity(w, req)

			default:
				w.WriteHeader(http.StatusNotFound)
			}
		})).ServeHTTP(w, req)
		return
	}

	w.WriteHeader(http.StatusNotFound)
}
