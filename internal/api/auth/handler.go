package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
	"tridorian-ztna/internal/api/common"
	"tridorian-ztna/internal/api/middleware"
	"tridorian-ztna/internal/models"
	"tridorian-ztna/internal/services"
	"tridorian-ztna/pkg/utils"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
)

type Handler struct {
	db                *gorm.DB
	adminService      *services.AdminService
	tenantService     *services.TenantService
	identityService   *services.IdentityService
	backofficeService *services.BackofficeService
	jwtSecret         string
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{
		db:                db,
		adminService:      services.NewAdminService(db),
		tenantService:     services.NewTenantService(db),
		identityService:   services.NewIdentityService(),
		backofficeService: services.NewBackofficeService(db),
		jwtSecret:         utils.GetEnv("JWT_SECRET", "very-secret-key"),
	}
}

// LoginManagement handles local administrator login for Management API access.
func (h *Handler) LoginManagement(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		common.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Extract domain from email
	parts := strings.Split(input.Email, "@")
	if len(parts) != 2 {
		common.Error(w, http.StatusBadRequest, "invalid email format")
		return
	}
	domain := parts[1]

	// Find tenant by domain
	tenant, err := h.tenantService.FindByDomain(domain)
	if err != nil {
		common.Error(w, http.StatusUnauthorized, "tenant not found for this domain")
		return
	}

	admin, err := h.adminService.Authenticate(tenant.ID, input.Email, input.Password)
	if err != nil {
		common.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	token, err := utils.GenerateToken(
		h.jwtSecret,
		utils.PurposeManagement,
		admin.ID.String(),
		tenant.ID.String(),
		string(admin.Role),
		nil, // Admins don't need groups for management access
		24*time.Hour,
	)
	if err != nil {
		common.Error(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	// Set secure HTTP-Only cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "mgmt_token",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	})

	common.Success(w, http.StatusOK, map[string]string{
		"message": "login successful",
	})
}

// LoginBackoffice handles system administrator login for the Backoffice API.
func (h *Handler) LoginBackoffice(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		common.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.backofficeService.Authenticate(input.Email, input.Password)
	if err != nil {
		common.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	token, err := utils.GenerateToken(
		h.jwtSecret,
		utils.PurposeBackoffice,
		user.ID.String(),
		"", // System wide, no tenant
		"super_admin",
		nil,
		24*time.Hour,
	)
	if err != nil {
		common.Error(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	// Set secure HTTP-Only cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "backoffice_token",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	common.Success(w, http.StatusOK, map[string]string{
		"message": "backoffice login successful",
	})
}

// LogoutBackoffice clears the backoffice authentication cookie.
func (h *Handler) LogoutBackoffice(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "backoffice_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		MaxAge:   -1,
	})
	common.Success(w, http.StatusOK, map[string]string{"message": "logged out"})
}

// MeBackoffice returns the currently authenticated backoffice user's information.
func (h *Handler) MeBackoffice(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value(middleware.AdminIDKey).(string)
	id, _ := uuid.Parse(userIDStr)

	user, err := h.backofficeService.GetByID(id)
	if err != nil {
		common.Error(w, http.StatusNotFound, "user not found")
		return
	}

	common.Success(w, http.StatusOK, user)
}

// LogoutManagement clears the management authentication cookie.
func (h *Handler) LogoutManagement(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "mgmt_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		MaxAge:   -1,
	})
	common.Success(w, http.StatusOK, map[string]string{"message": "logged out"})
}

// MeManagement returns the currently authenticated administrator's information.
func (h *Handler) MeManagement(w http.ResponseWriter, r *http.Request) {
	adminIDStr := r.Context().Value(middleware.AdminIDKey).(string)
	id, _ := uuid.Parse(adminIDStr)

	admin, err := h.adminService.GetByID(id)
	if err != nil {
		common.Error(w, http.StatusNotFound, "admin not found")
		return
	}

	common.Success(w, http.StatusOK, admin)
}

// getOAuth2Config constructs the OAuth2 config using tenant-specific credentials
func (h *Handler) getOAuth2Config(tenant *models.Tenant, r *http.Request) *oauth2.Config {
	protocol := "http"
	if r.TLS != nil {
		protocol = "https"
	}
	redirectURL := fmt.Sprintf("%s://%s/auth/target/callback", protocol, r.Host)

	return &oauth2.Config{
		ClientID:     tenant.GoogleClientID,
		ClientSecret: tenant.GoogleClientSecret,
		RedirectURL:  redirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email",
		},
		Endpoint: google.Endpoint,
	}
}

// LoginTarget initiates Google OAuth2 for VPN users
func (h *Handler) LoginTarget(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.GetTenant(r.Context())
	if tenant == nil || tenant.GoogleClientID == "" {
		common.Error(w, http.StatusForbidden, "tenant Google Identity is not configured")
		return
	}

	// Decrypt sensitive fields before using them
	if err := h.tenantService.DecryptTenantConfig(tenant); err != nil {
		common.Error(w, http.StatusInternalServerError, "failed to decrypt identity config")
		return
	}

	conf := h.getOAuth2Config(tenant, r)
	url := conf.AuthCodeURL("state-todo") // In production, use a secure random state

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// CallbackTarget handles the callback from Google, verifies the user, and issues a Target Token
func (h *Handler) CallbackTarget(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.GetTenant(r.Context())
	if tenant == nil {
		common.Error(w, http.StatusForbidden, "tenant not found")
		return
	}
	tenantID := tenant.ID

	// Decrypt sensitive fields before using them
	if err := h.tenantService.DecryptTenantConfig(tenant); err != nil {
		common.Error(w, http.StatusInternalServerError, "failed to decrypt identity config")
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		common.Error(w, http.StatusBadRequest, "authorization code is missing")
		return
	}

	conf := h.getOAuth2Config(tenant, r)
	tok, err := conf.Exchange(r.Context(), code)
	if err != nil {
		common.Error(w, http.StatusInternalServerError, "failed to exchange token: "+err.Error())
		return
	}

	// Fetch user info from Google using the token
	client := conf.Client(r.Context(), tok)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		common.Error(w, http.StatusInternalServerError, "failed to get user info")
		return
	}
	defer resp.Body.Close()

	var googleUser struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		common.Error(w, http.StatusInternalServerError, "failed to decode user info")
		return
	}

	// Fetch groups using Service Account (Requires Admin SDK setup in Tenant)
	var groups []string
	if tenant.GoogleServiceAccountKey != "" {
		groups, _ = h.identityService.GetUserGroups(
			r.Context(),
			[]byte(tenant.GoogleServiceAccountKey),
			tenant.GoogleAdminEmail,
			googleUser.Email,
		)
	}

	// Issue Target Token with Groups
	targetToken, err := utils.GenerateToken(
		h.jwtSecret,
		utils.PurposeTarget,
		googleUser.ID,
		tenantID.String(),
		"user",
		groups,
		2*time.Hour,
	)
	if err != nil {
		common.Error(w, http.StatusInternalServerError, "failed to generate target token")
		return
	}

	common.Success(w, http.StatusOK, map[string]interface{}{
		"token":  targetToken,
		"email":  googleUser.Email,
		"name":   googleUser.Name,
		"groups": groups,
	})
}
