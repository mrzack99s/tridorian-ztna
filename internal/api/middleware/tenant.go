package middleware

import (
	"context"
	"net/http"
	"strings"
	"tridorian-ztna/internal/api/common"
	"tridorian-ztna/internal/models"
	"tridorian-ztna/internal/services"
	"tridorian-ztna/pkg/utils"

	"github.com/google/uuid"
)

type contextKey string

const TenantIDKey contextKey = "tenant_id"

// TenantFromHeader extracts tenant ID from X-Tenant-ID header (Management API style)
func TenantFromHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idStr := r.Header.Get("X-Tenant-ID")
		if idStr == "" {
			common.Error(w, http.StatusBadRequest, "X-Tenant-ID header is missing")
			return
		}

		id, err := uuid.Parse(idStr)
		if err != nil {
			common.Error(w, http.StatusBadRequest, "invalid tenant ID format")
			return
		}

		ctx := context.WithValue(r.Context(), TenantIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

const TenantContextKey contextKey = "tenant_context"

// ResolveTenantByHost resolves tenant from the Request Host (Auth API style)
func ResolveTenantByHost(tenantService *services.TenantService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			host := r.Host

			if strings.Contains(host, ":") {
				host = strings.Split(host, ":")[0]
			}

			tenant, err := tenantService.FindByDomain(host)
			if err != nil {
				// Special case for local development
				if host == "localhost" || host == "127.0.0.1" {
					tenant, err = tenantService.FindBySlug("default")
				}

				if err != nil {
					// User is accessing via an unknown or deleted tenant domain
					ip := utils.GetClientIP(r)
					common.RenderErrorPage(w, http.StatusUnauthorized, "Unknown Tenant", "The domain you are accessing is not a valid Tridorian ZTNA tenant.", "Domain: "+host, ip, "")
					return
				}
			}

			ctx := context.WithValue(r.Context(), TenantIDKey, tenant.ID)
			ctx = context.WithValue(ctx, TenantContextKey, tenant)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetTenant retrieves the full tenant object from context
func GetTenant(ctx context.Context) *models.Tenant {
	if tenant, ok := ctx.Value(TenantContextKey).(*models.Tenant); ok {
		return tenant
	}
	return nil
}

// GetTenantID retrieves tenant ID from context
func GetTenantID(ctx context.Context) uuid.UUID {
	if id, ok := ctx.Value(TenantIDKey).(uuid.UUID); ok {
		return id
	}
	return uuid.Nil
}
