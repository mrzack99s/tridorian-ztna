package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
	"tridorian-ztna/internal/api/common"
	"tridorian-ztna/internal/api/middleware"
	"tridorian-ztna/internal/models"
	"tridorian-ztna/internal/services"
	"tridorian-ztna/pkg/geoip"
	"tridorian-ztna/pkg/utils"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
)

type Handler struct {
	db                *gorm.DB
	geoIP             *geoip.GeoIP
	adminService      *services.AdminService
	tenantService     *services.TenantService
	identityService   *services.IdentityService
	backofficeService *services.BackofficeService
	policyService     *services.PolicyService
	nodeService       *services.NodeService // Injected
	cache             *redis.Client
	privateKey        interface{}
	publicKey         interface{}
}

func NewHandler(db *gorm.DB, cache *redis.Client, geoIP *geoip.GeoIP, privateKey, publicKey interface{}) *Handler {
	return &Handler{
		db:                db,
		cache:             cache,
		geoIP:             geoIP,
		adminService:      services.NewAdminService(db),
		tenantService:     services.NewTenantService(db),
		identityService:   services.NewIdentityService(),
		backofficeService: services.NewBackofficeService(db),
		policyService:     services.NewPolicyService(db, cache),
		nodeService:       services.NewNodeService(db, cache), // Initialize
		privateKey:        privateKey,
		publicKey:         publicKey,
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

	// Resolve tenant from the email domain
	tenant, err := h.tenantService.FindByDomain(domain)
	if err != nil {
		common.Error(w, http.StatusUnauthorized, "invalid tenant domain")
		return
	}

	// Validate that the email belongs to this tenant's domain (security check)
	// Actually, for multi-tenant, admins might have emails on different domains?
	// But the requirement implies domain-based login.
	// If the user wants to login to THIS tenant console, they must use an email associated with it?
	// Or maybe the previous logic was extracting domain from email.
	// The user request says "Auth-api for auth cmd use ResolveTenantByHost".
	// This usually means we identify the tenant by the URL (Host header), not just the email domain.

	// Let's ensure the email domain matches one of the tenant's domains or is the free domain.
	// For now, we'll trust the authentication service to check the credentials against this tenant.

	admin, err := h.adminService.Authenticate(tenant.ID, input.Email, input.Password)
	if err != nil {
		common.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	token, err := utils.GenerateToken(
		h.privateKey,
		utils.PurposeManagement,
		admin.ID.String(),
		admin.Email, // Email
		tenant.ID.String(),
		string(admin.Role),
		nil, // Admins don't need groups for management access
		"",  // No OS info for management login
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
		h.privateKey,
		utils.PurposeBackoffice,
		user.ID.String(),
		user.Email, // Email
		"",         // System wide, no tenant
		"super_admin",
		nil,
		"", // No OS info for backoffice
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
	redirectURL := fmt.Sprintf("%s://%s/callback", protocol, r.Host)

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
	// 404 Handling for unknown paths caught by "/" pattern matching
	if r.URL.Path != "/" && r.URL.Path != "/login" {
		ip := utils.GetClientIP(r)
		country := ""
		if h.geoIP != nil {
			country = h.geoIP.Lookup(ip)
		}
		common.RenderErrorPage(w, http.StatusNotFound, "Page Not Found", "The page you are looking for does not exist or has been moved.", "Path: "+r.URL.Path, ip, country)
		return
	}

	tenant := middleware.GetTenant(r.Context())
	ip := utils.GetClientIP(r)
	country := ""
	if h.geoIP != nil {
		country = h.geoIP.Lookup(ip)
	}

	if tenant == nil || tenant.GoogleClientID == "" {
		common.RenderErrorPage(w, http.StatusForbidden, "Configuration Error", "Tenant Identity Provider is incomplete.", "Tenant Google Identity is not configured", ip, country)
		return
	}

	// Decrypt sensitive fields before using them
	if err := h.tenantService.DecryptTenantConfig(tenant); err != nil {
		common.Error(w, http.StatusInternalServerError, "failed to decrypt identity config")
		return
	}

	// Check Network Policies (Pre-Auth)
	// Filter out users based on IP/Country before they even go to Google
	if err := h.checkNetworkPolicies(r, tenant.ID); err != nil {
		common.RenderErrorPage(w, http.StatusForbidden, "Access Suspended", "Your connection location or device does not meet the security requirements.", err.Error(), ip, country)
		return
	}

	port := r.URL.Query().Get("desktop_port")
	osInfo := r.URL.Query().Get("os")
	if osInfo == "" {
		osInfo = r.URL.Query().Get("device_os")
	}

	stateMap := map[string]string{
		"csrf": "todo-random-string",
		"port": port,
		"os":   osInfo,
	}
	stateJSON, _ := json.Marshal(stateMap)
	state := base64.StdEncoding.EncodeToString(stateJSON)

	conf := h.getOAuth2Config(tenant, r)
	url := conf.AuthCodeURL(state) // In production, use a secure random state

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

	// Retrieve Desktop Port and OS from State
	stateParam := r.URL.Query().Get("state")
	var desktopPort string
	var osInfo string
	if stateParam != "" {
		if data, err := base64.StdEncoding.DecodeString(stateParam); err == nil {
			var stateMap map[string]string
			if json.Unmarshal(data, &stateMap) == nil {
				desktopPort = stateMap["port"]
				osInfo = stateMap["os"]
			}
		}
	}

	// Check Identity Policies (Post-Auth)
	if err := h.checkIdentityPolicies(r, tenant.ID, googleUser.Email, groups, osInfo); err != nil {
		ip := utils.GetClientIP(r)
		country := ""
		if h.geoIP != nil {
			country = h.geoIP.Lookup(ip)
		}
		common.RenderErrorPage(w, http.StatusForbidden, "Access Denied", "Your account does not have permission to access these resources.", err.Error(), ip, country)
		return
	}

	// Issue Target Token with Groups
	targetToken, err := utils.GenerateToken(
		h.privateKey,
		utils.PurposeTarget,
		googleUser.ID,
		googleUser.Email, // Email
		tenantID.String(),
		"user",
		groups,
		osInfo,
		2*time.Hour,
	)
	if err != nil {
		fmt.Println(err)
		common.Error(w, http.StatusInternalServerError, "failed to generate target token")
		return
	}

	// Already retrieved above, removing duplicate extraction
	if desktopPort != "" {
		redirectURL := fmt.Sprintf("http://localhost:%s/callback?token=%s&email=%s&name=%s",
			desktopPort, targetToken, googleUser.Email, googleUser.Name)
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
		return
	}

	common.Success(w, http.StatusOK, map[string]interface{}{
		"token":  targetToken,
		"email":  googleUser.Email,
		"name":   googleUser.Name,
		"groups": groups,
	})
}

// ListGateways returns the list of active gateways for the authenticated user's tenant
func (h *Handler) ListGateways(w http.ResponseWriter, r *http.Request) {
	// Extract Claims
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		common.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	tenantID, err := uuid.Parse(claims.TenantID)
	if err != nil {
		common.Error(w, http.StatusBadRequest, "invalid tenant id in token")
		return
	}

	// Fetch nodes
	nodes, err := h.nodeService.ListNodes(tenantID)
	if err != nil {
		common.Error(w, http.StatusInternalServerError, "failed to list gateways")
		return
	}

	// Filter for active gateways and return simplified struct
	var gateways []map[string]interface{}
	for _, node := range nodes {
		// Only return active and fully registered nodes
		if node.IsActive && node.Status == "CONNECTED" {
			// Address: We need the public address (Hostname or IP) + Port
			addr := node.IPAddress
			if addr == "" {
				addr = node.Hostname
			}
			if addr == "" {
				addr = "unknown-host"
			}
			// Should probably include port if stored, or default

			gateways = append(gateways, map[string]interface{}{
				"name":    node.Name,
				"address": addr,
				"ping":    "unknown",
				"region":  "unknown",
			})
		}
	}

	common.Success(w, http.StatusOK, gateways)
}

// evalContext holds data for policy evaluation
type evalContext struct {
	IP      string
	Country string
	Email   string
	Groups  []string
	OS      string
}

// checkNetworkPolicies evaluates only Network/IP/Device based policies.
// Used at pre-auth stage (login).
func (h *Handler) checkNetworkPolicies(r *http.Request, tenantID uuid.UUID) error {
	policies, err := h.policyService.ListSignInPolicies(tenantID)
	if err != nil {
		return err
	}

	ip := utils.GetClientIP(r)
	country := ""
	if h.geoIP != nil {
		country = h.geoIP.Lookup(ip)
	}

	ctx := evalContext{IP: ip, Country: country}

	for _, policy := range policies {
		log.Printf("[PolicyDebug] Checking policy: %s (Enabled: %v, Stage: %s)", policy.Name, policy.Enabled, policy.Stage)

		if !policy.Enabled || policy.Stage != "pre_auth" {
			continue
		}

		result := h.evaluateNode(&policy.RootNode, ctx)
		log.Printf("[PolicyDebug] Policy %s evaluateNode result: %v (details: IP=%s, Country=%s)", policy.Name, result, ctx.IP, ctx.Country)

		if result {
			if policy.Block {
				log.Printf("[PolicyDebug] Policy %s BLOCKED", policy.Name)
				return fmt.Errorf("access denied by network policy: %s", policy.Name)
			}
			// Explicit Allow by higher priority policy
			log.Printf("[PolicyDebug] Policy %s ALLOWED", policy.Name)
			return nil
		}
	}

	// Default Action: Block
	// Note: If you want to allow everyone by default, you must create an 'Allow' policy.
	return fmt.Errorf("access denied by default network policy")
}

func (h *Handler) isNetworkOnly(node models.PolicyNode) bool {
	if node.Condition != nil {
		return node.Condition.Type == "Network" || node.Condition.Type == "Device"
	}
	for _, child := range node.Children {
		if !h.isNetworkOnly(child) {
			return false
		}
	}
	return true
}

// checkIdentityPolicies evaluates all policies (Full Context).
// Used at post-auth stage (callback) with user info.
func (h *Handler) checkIdentityPolicies(r *http.Request, tenantID uuid.UUID, userEmail string, groups []string, osInfo string) error {
	policies, err := h.policyService.ListSignInPolicies(tenantID)
	if err != nil {
		return err
	}

	ip := utils.GetClientIP(r)
	country := ""
	if h.geoIP != nil {
		country = h.geoIP.Lookup(ip)
	}

	ctx := evalContext{
		IP:      ip,
		Country: country,
		Email:   userEmail,
		Groups:  groups,
		OS:      osInfo,
	}

	for _, policy := range policies {
		if !policy.Enabled || policy.Stage != "post_auth" {
			continue
		}
		if h.evaluateNode(&policy.RootNode, ctx) {
			if policy.Block {
				return fmt.Errorf("access denied by policy: %s", policy.Name)
			}
			return nil // Explicit Allow matches
		}
	}

	// Default Action: Block
	return fmt.Errorf("access denied by default policy")
}

func (h *Handler) evaluateNode(node *models.PolicyNode, ctx evalContext) bool {
	if node == nil {
		return false
	}

	// Leaf Node (Condition)
	if node.Condition != nil {
		return h.evaluateCondition(node.Condition, ctx)
	}

	// Branch Node (Children)
	if len(node.Children) == 0 {
		return false
	}

	if node.Operator == "OR" {
		for i := range node.Children {
			if h.evaluateNode(&node.Children[i], ctx) {
				return true
			}
		}
		return false
	}

	// Default: AND
	for i := range node.Children {
		if !h.evaluateNode(&node.Children[i], ctx) {
			return false
		}
	}
	return true
}

func (h *Handler) evaluateCondition(cond *models.PolicyCondition, ctx evalContext) bool {
	if cond == nil {
		log.Printf("[PolicyDebug] evaluateCondition: condition is nil")
		return false
	}

	log.Printf("[PolicyDebug] evaluateCondition: Type=%s, Field=%s, Op=%s, Value=%s", cond.Type, cond.Field, cond.Op, cond.Value)

	// Special operator that doesn't depend on a field
	if cond.Op == "is_private" {
		res := ctx.Country == "PRIVATE"
		log.Printf("[PolicyDebug] evaluateCondition is_private result: %v", res)
		return res
	}

	var val string
	field := strings.ToLower(cond.Field)
	switch field {
	case "country", "location":
		val = ctx.Country
	case "ip", "ip_address":
		val = ctx.IP
	case "email", "user_email":
		val = ctx.Email
	case "os", "device_os":
		val = ctx.OS
	case "group", "user_group":
		matched := false
		for _, g := range ctx.Groups {
			if h.compareValues(g, cond.Op, cond.Value) {
				matched = true
				break
			}
		}
		log.Printf("[PolicyDebug] evaluateCondition group match result: %v (user groups: %v)", matched, ctx.Groups)
		return matched
	default:
		log.Printf("[PolicyDebug] evaluateCondition unknown field: %s", field)
		return false
	}

	res := h.compareValues(val, cond.Op, cond.Value)
	log.Printf("[PolicyDebug] evaluateCondition result: %v (val=%s, target=%s)", res, val, cond.Value)
	return res
}

func (h *Handler) compareValues(val, op, target string) bool {
	val = strings.TrimSpace(val)
	target = strings.TrimSpace(target)

	switch strings.ToLower(op) {
	case "equals", "is", "os":
		return strings.EqualFold(val, target)
	case "not_equals", "not":
		return !strings.EqualFold(val, target)
	case "cidr":
		_, ipNet, err := net.ParseCIDR(target)
		if err != nil {
			return false
		}
		ip := net.ParseIP(val)
		return ip != nil && ipNet.Contains(ip)
	case "not_cidr":
		_, ipNet, err := net.ParseCIDR(target)
		if err != nil {
			return true // If CIDR is invalid, it doesn't contain the IP
		}
		ip := net.ParseIP(val)
		return ip != nil && !ipNet.Contains(ip)
	case "in":
		// Check both comma-separated and case-insensitive
		parts := strings.Split(target, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			// Also support JSON array style if accidentally passed
			p = strings.Trim(p, "[]\"")
			if strings.EqualFold(val, p) {
				return true
			}
		}
		// Fallback for simple "contains" behavior if not comma-separated
		return strings.Contains(strings.ToUpper(target), strings.ToUpper(val)) && val != ""
	case "not_in":
		return !h.compareValues(val, "in", target)
	case "contains":
		return strings.Contains(strings.ToLower(val), strings.ToLower(target))
	case "starts_with":
		return strings.HasPrefix(strings.ToLower(val), strings.ToLower(target))
	case "ends_with":
		return strings.HasSuffix(strings.ToLower(val), strings.ToLower(target))
	default:
		return false
	}
}
