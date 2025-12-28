package middleware

import (
	"context"
	"net/http"
	"strings"
	"tridorian-ztna/internal/api/common"
	"tridorian-ztna/internal/models"
	"tridorian-ztna/pkg/utils"

	"github.com/google/uuid"
)

const (
	AdminIDKey contextKey = "admin_id"
	RoleKey    contextKey = "role"
	ClaimsKey  contextKey = "claims"
)

// GetClaims retrieves the full token claims from context
func GetClaims(ctx context.Context) *utils.Claims {
	if claims, ok := ctx.Value(ClaimsKey).(*utils.Claims); ok {
		return claims
	}
	return nil
}

func JWTAuth(key interface{}, purpose utils.TokenPurpose) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := ""
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && parts[0] == "Bearer" {
					token = parts[1]
				}
			}

			// If no header, check cookie
			if token == "" {
				cookieName := "mgmt_token"
				if purpose == utils.PurposeBackoffice {
					cookieName = "backoffice_token"
				}

				cookie, err := r.Cookie(cookieName)
				if err == nil {
					token = cookie.Value
				}
			}

			if token == "" {
				common.Error(w, http.StatusUnauthorized, "authentication required")
				return
			}

			claims, err := utils.ParseToken(key, token)
			if err != nil {
				common.Error(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			// Validate Purpose
			if claims.Purpose != purpose {
				common.Error(w, http.StatusForbidden, "token not authorized for this action")
				return
			}

			// Add to context
			ctx := context.WithValue(r.Context(), AdminIDKey, claims.Subject)
			ctx = context.WithValue(ctx, RoleKey, claims.Role)
			ctx = context.WithValue(ctx, ClaimsKey, claims)

			// Ensure TenantID is also in context from the token
			if claims.TenantID != "" {
				tenantUUID, _ := uuid.Parse(claims.TenantID)
				ctx = context.WithValue(ctx, TenantIDKey, tenantUUID)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireRole(roles ...models.AdminRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := r.Context().Value(RoleKey).(string)
			if !ok {
				common.Error(w, http.StatusUnauthorized, "role not found in token")
				return
			}

			// Super Admin can do anything
			if role == string(models.RoleSuperAdmin) {
				next.ServeHTTP(w, r)
				return
			}

			for _, allowedRole := range roles {
				if role == string(allowedRole) {
					next.ServeHTTP(w, r)
					return
				}
			}

			common.Error(w, http.StatusForbidden, "you do not have permission to perform this action")
		})
	}
}
