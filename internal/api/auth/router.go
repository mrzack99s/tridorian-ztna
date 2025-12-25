package auth

import (
	"net/http"
	"tridorian-ztna/internal/api/middleware"
	"tridorian-ztna/internal/services"
	"tridorian-ztna/pkg/utils"

	"gorm.io/gorm"
)

type Router struct {
	handler          *Handler
	tenantMiddleware func(http.Handler) http.Handler
}

func NewRouter(db *gorm.DB) *Router {
	tenantService := services.NewTenantService(db)
	return &Router{
		handler:          NewHandler(db),
		tenantMiddleware: middleware.ResolveTenantByHost(tenantService),
	}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	method := req.Method

	// 0. Backoffice Authentication (System Wide)
	if path == "/auth/backoffice/login" {
		if method == http.MethodPost {
			r.handler.LoginBackoffice(w, req)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}

	if path == "/auth/backoffice/logout" {
		r.handler.LogoutBackoffice(w, req)
		return
	}

	if path == "/auth/backoffice/me" {
		middleware.JWTAuth(r.handler.jwtSecret, utils.PurposeBackoffice)(http.HandlerFunc(r.handler.MeBackoffice)).ServeHTTP(w, req)
		return
	}

	// 1. Management Authentication (For local admins)
	if path == "/auth/mgmt/login" {
		if method == http.MethodPost {
			r.handler.LoginManagement(w, req)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}

	if path == "/auth/mgmt/logout" {
		r.handler.LogoutManagement(w, req)
		return
	}

	if path == "/auth/mgmt/me" {
		middleware.JWTAuth(r.handler.jwtSecret, utils.PurposeManagement)(http.HandlerFunc(r.handler.MeManagement)).ServeHTTP(w, req)
		return
	}

	// Auth routes always require tenant context from host
	tenantBoundHandlers := http.NewServeMux()

	// 2. Target Authentication (For VPN users via Google Identity)
	tenantBoundHandlers.HandleFunc("/auth/target/login", func(w http.ResponseWriter, req *http.Request) {
		r.handler.LoginTarget(w, req)
	})

	tenantBoundHandlers.HandleFunc("/auth/target/callback", func(w http.ResponseWriter, req *http.Request) {
		r.handler.CallbackTarget(w, req)
	})

	// Public Health Check (Optional, but good to have inside the mux or outside)
	if path == "/health" && method == http.MethodGet {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
		return
	}

	// Wrap specific paths with tenant middleware
	r.tenantMiddleware(tenantBoundHandlers).ServeHTTP(w, req)
}
