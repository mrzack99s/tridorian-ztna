package mgmt

import (
	"net/http"
	"strings"
	"tridorian-ztna/internal/api/middleware"
	"tridorian-ztna/internal/models"
	"tridorian-ztna/internal/services"
	"tridorian-ztna/pkg/utils"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Router struct {
	handler   *Handler
	publicKey interface{}
}

func NewRouter(db *gorm.DB, cache *redis.Client, privateKey, publicKey interface{}) *Router {
	adminService := services.NewAdminService(db)
	tenantService := services.NewTenantService(db)
	policyService := services.NewPolicyService(db, cache)
	nodeService := services.NewNodeService(db, cache)
	identityService := services.NewIdentityService()
	applicationService := services.NewApplicationService(db)

	return &Router{
		handler:   NewHandler(adminService, tenantService, policyService, nodeService, identityService, applicationService),
		publicKey: publicKey,
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
		middleware.JWTAuth(r.publicKey, utils.PurposeBackoffice)(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
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
		"/api/v1/applications",
		"/api/v1/nodes",
		"/api/v1/nodes/skus",
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
		middleware.JWTAuth(r.publicKey, utils.PurposeManagement)(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			switch {
			case path == "/api/v1/tenants/activate" && req.Method == "POST":
				middleware.RequireRole(models.RoleSuperAdmin, models.RoleAdmin)(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					r.handler.ActivateDomain(w, req)
				})).ServeHTTP(w, req)

			case path == "/api/v1/tenants/identity" && req.Method == "POST":
				middleware.RequireRole(models.RoleSuperAdmin)(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					r.handler.UpdateIdentity(w, req)
				})).ServeHTTP(w, req)

			case path == "/api/v1/tenant/me" && req.Method == "GET":
				r.handler.GetMyTenant(w, req)

			case path == "/api/v1/tenant/me" && (req.Method == "PATCH" || req.Method == "PUT"):
				middleware.RequireRole(models.RoleSuperAdmin, models.RoleAdmin)(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					r.handler.UpdateTenant(w, req)
				})).ServeHTTP(w, req)

			// Custom Domain Management
			case path == "/api/v1/tenants/domains":
				switch req.Method {
				case "GET":
					r.handler.ListDomains(w, req)
				case "POST":
					middleware.RequireRole(models.RoleSuperAdmin, models.RoleAdmin)(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
						r.handler.RegisterCustomDomain(w, req)
					})).ServeHTTP(w, req)
				}

			case path == "/api/v1/tenants/domains/verify" && req.Method == "POST":
				middleware.RequireRole(models.RoleSuperAdmin, models.RoleAdmin)(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					r.handler.VerifyCustomDomain(w, req)
				})).ServeHTTP(w, req)

			case path == "/api/v1/profile/change-password" && req.Method == "POST":
				r.handler.ChangePassword(w, req)

			case strings.HasPrefix(path, "/api/v1/admins"):
				middleware.RequireRole(models.RoleSuperAdmin)(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					switch req.Method {
					case "GET":
						r.handler.ListAdmins(w, req)
					case "POST":
						r.handler.CreateAdmin(w, req)
					case "PATCH", "PUT":
						r.handler.UpdateAdmin(w, req)
					case "DELETE":
						r.handler.DeleteAdmin(w, req)
					default:
						w.WriteHeader(http.StatusMethodNotAllowed)
					}
				})).ServeHTTP(w, req)

			// Policy Management
			case path == "/api/v1/policies/access":
				middleware.RequireRole(models.RoleSuperAdmin, models.RoleAdmin, models.RolePolicyAdmin)(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					switch req.Method {
					case "GET":
						r.handler.ListAccessPolicies(w, req)
					case "POST":
						r.handler.CreateAccessPolicy(w, req)
					case "PATCH", "PUT":
						r.handler.UpdateAccessPolicy(w, req)
					case "DELETE":
						r.handler.DeleteAccessPolicy(w, req)
					default:
						w.WriteHeader(http.StatusMethodNotAllowed)
					}
				})).ServeHTTP(w, req)

			case path == "/api/v1/policies/sign-in":
				middleware.RequireRole(models.RoleSuperAdmin, models.RoleAdmin, models.RolePolicyAdmin)(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					switch req.Method {
					case "GET":
						r.handler.ListSignInPolicies(w, req)
					case "POST":
						r.handler.CreateSignInPolicy(w, req)
					case "PATCH", "PUT":
						r.handler.UpdateSignInPolicy(w, req)
					case "DELETE":
						r.handler.DeleteSignInPolicy(w, req)
					default:
						w.WriteHeader(http.StatusMethodNotAllowed)
					}
				})).ServeHTTP(w, req)

			// Application Management
			case strings.HasPrefix(path, "/api/v1/applications"):
				middleware.RequireRole(models.RoleSuperAdmin, models.RoleAdmin)(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					switch req.Method {
					case "GET":
						if strings.Contains(req.URL.RawQuery, "id=") {
							r.handler.GetApplication(w, req)
						} else {
							r.handler.ListApplications(w, req)
						}
					case "POST":
						r.handler.CreateApplication(w, req)
					case "PATCH", "PUT":
						r.handler.UpdateApplication(w, req)
					case "DELETE":
						r.handler.DeleteApplication(w, req)
					default:
						w.WriteHeader(http.StatusMethodNotAllowed)
					}
				})).ServeHTTP(w, req)

			// Node Management
			case path == "/api/v1/nodes/skus" && req.Method == "GET":
				r.handler.ListNodeSkus(w, req)

			case path == "/api/v1/nodes":
				middleware.RequireRole(models.RoleSuperAdmin, models.RoleAdmin)(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					switch req.Method {
					case "GET":
						if strings.Contains(path, "/sessions") {
							r.handler.ListNodeSessions(w, req)
						} else {
							r.handler.ListNodes(w, req)
						}
					case "POST":
						r.handler.CreateNode(w, req)
					case "DELETE":
						r.handler.DeleteNode(w, req)
					}
				})).ServeHTTP(w, req)

			case path == "/api/v1/nodes/sessions" && req.Method == "GET":
				middleware.RequireRole(models.RoleSuperAdmin, models.RoleAdmin)(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					r.handler.ListNodeSessions(w, req)
				})).ServeHTTP(w, req)

			// Identity Management
			case path == "/api/v1/identity/search" && req.Method == "GET":
				middleware.RequireRole(models.RoleSuperAdmin, models.RoleAdmin, models.RolePolicyAdmin)(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					r.handler.SearchIdentity(w, req)
				})).ServeHTTP(w, req)

			default:
				w.WriteHeader(http.StatusNotFound)
			}
		})).ServeHTTP(w, req)
		return
	}

	w.WriteHeader(http.StatusNotFound)
}
