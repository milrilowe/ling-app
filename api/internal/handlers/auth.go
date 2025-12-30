package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"ling-app/api/internal/config"
	"ling-app/api/internal/middleware"
	"ling-app/api/internal/services"
	"ling-app/api/internal/services/auth"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	AuthService    *auth.AuthService
	OAuthService   *services.OAuthService
	CreditsService *services.CreditsService
	Config         *config.Config
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *auth.AuthService, oauthService *services.OAuthService, creditsService *services.CreditsService, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		AuthService:    authService,
		OAuthService:   oauthService,
		CreditsService: creditsService,
		Config:         cfg,
	}
}

// Request/Response types

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UserResponse struct {
	ID            string  `json:"id"`
	Email         string  `json:"email"`
	Name          string  `json:"name"`
	AvatarURL     *string `json:"avatarUrl,omitempty"`
	EmailVerified bool    `json:"emailVerified"`
}

// Helper to determine cookie settings based on environment
func (h *AuthHandler) getCookieSettings() (secure bool, sameSite http.SameSite, domain string) {
	if h.Config.Environment == "production" {
		return true, http.SameSiteLaxMode, ""
	}
	// Development: don't require HTTPS
	return false, http.SameSiteLaxMode, ""
}

// setSessionCookie sets the session cookie on the response
func (h *AuthHandler) setSessionCookie(c *gin.Context, token string) {
	secure, sameSite, domain := h.getCookieSettings()

	c.SetSameSite(sameSite)
	c.SetCookie(
		"session_token",           // name
		token,                     // value
		h.Config.SessionMaxAge,    // maxAge in seconds
		"/",                       // path
		domain,                    // domain (empty = current domain)
		secure,                    // secure (HTTPS only)
		true,                      // httpOnly (not accessible via JS)
	)
}

// clearSessionCookie removes the session cookie
func (h *AuthHandler) clearSessionCookie(c *gin.Context) {
	secure, sameSite, domain := h.getCookieSettings()

	c.SetSameSite(sameSite)
	c.SetCookie(
		"session_token",
		"",
		-1,     // negative maxAge = delete cookie
		"/",
		domain,
		secure,
		true,
	)
}

// Register creates a new user account
// POST /api/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Normalize email to lowercase
	email := strings.ToLower(strings.TrimSpace(req.Email))
	name := strings.TrimSpace(req.Name)

	// Create user with credits (atomic transaction)
	user, err := h.AuthService.CreateUser(email, req.Password, name, h.CreditsService)
	if err != nil {
		if err == auth.ErrEmailTaken {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create account"})
		return
	}

	// Create session
	token, err := h.AuthService.CreateSession(user.ID, c.Request.UserAgent(), c.ClientIP())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	// Set cookie
	h.setSessionCookie(c, token)

	// Return user (without sensitive fields)
	c.JSON(http.StatusCreated, UserResponse{
		ID:            user.ID.String(),
		Email:         user.Email,
		Name:          user.Name,
		AvatarURL:     user.AvatarURL,
		EmailVerified: user.EmailVerified,
	})
}

// Login authenticates a user and creates a session
// POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Normalize email
	email := strings.ToLower(strings.TrimSpace(req.Email))

	// Authenticate
	user, err := h.AuthService.AuthenticateUser(email, req.Password)
	if err != nil {
		if err == auth.ErrInvalidCredentials {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication failed"})
		return
	}

	// Create session
	token, err := h.AuthService.CreateSession(user.ID, c.Request.UserAgent(), c.ClientIP())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	// Set cookie
	h.setSessionCookie(c, token)

	// Return user
	c.JSON(http.StatusOK, UserResponse{
		ID:            user.ID.String(),
		Email:         user.Email,
		Name:          user.Name,
		AvatarURL:     user.AvatarURL,
		EmailVerified: user.EmailVerified,
	})
}

// Logout ends the current session
// POST /api/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// Get session token from cookie
	token, err := c.Cookie("session_token")
	if err == nil && token != "" {
		// Delete session from database (ignore errors - we're logging out anyway)
		_ = h.AuthService.DeleteSession(token)
	}

	// Clear cookie
	h.clearSessionCookie(c)

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// GetMe returns the current authenticated user
// GET /api/auth/me
// Requires: RequireAuth middleware
func (h *AuthHandler) GetMe(c *gin.Context) {
	user, ok := middleware.GetUserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	c.JSON(http.StatusOK, UserResponse{
		ID:            user.ID.String(),
		Email:         user.Email,
		Name:          user.Name,
		AvatarURL:     user.AvatarURL,
		EmailVerified: user.EmailVerified,
	})
}

// ============================================
// OAuth Handlers
// ============================================

// generateOAuthState creates a random state token for CSRF protection
func generateOAuthState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// setOAuthStateCookie stores the OAuth state in a short-lived cookie
func (h *AuthHandler) setOAuthStateCookie(c *gin.Context, state string) {
	secure, sameSite, domain := h.getCookieSettings()
	c.SetSameSite(sameSite)
	c.SetCookie(
		"oauth_state",
		state,
		300, // 5 minutes
		"/",
		domain,
		secure,
		true,
	)
}

// clearOAuthStateCookie removes the OAuth state cookie
func (h *AuthHandler) clearOAuthStateCookie(c *gin.Context) {
	secure, sameSite, domain := h.getCookieSettings()
	c.SetSameSite(sameSite)
	c.SetCookie("oauth_state", "", -1, "/", domain, secure, true)
}

// GoogleLogin initiates Google OAuth flow
// GET /api/auth/google
func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	if !h.OAuthService.IsGoogleEnabled() {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Google OAuth not configured"})
		return
	}

	state, err := generateOAuthState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate state"})
		return
	}

	h.setOAuthStateCookie(c, state)

	url, err := h.OAuthService.GetGoogleAuthURL(state)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate auth URL"})
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GoogleCallback handles the Google OAuth callback
// GET /api/auth/google/callback
func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	// Verify state parameter
	state := c.Query("state")
	storedState, err := c.Cookie("oauth_state")
	if err != nil || state != storedState {
		c.Redirect(http.StatusTemporaryRedirect, h.Config.FrontendURL+"/login?error=invalid_state")
		return
	}
	h.clearOAuthStateCookie(c)

	// Check for error from OAuth provider
	if errParam := c.Query("error"); errParam != "" {
		c.Redirect(http.StatusTemporaryRedirect, h.Config.FrontendURL+"/login?error="+errParam)
		return
	}

	// Exchange code for user info
	code := c.Query("code")
	googleUser, err := h.OAuthService.ExchangeGoogleCode(c.Request.Context(), code)
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, h.Config.FrontendURL+"/login?error=oauth_failed")
		return
	}

	// Find or create user (credits initialized atomically for new users)
	user, _, err := h.AuthService.FindOrCreateOAuthUser(
		"google",
		googleUser.ID,
		googleUser.Email,
		googleUser.Name,
		googleUser.Picture,
		h.CreditsService,
	)
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, h.Config.FrontendURL+"/login?error=account_error")
		return
	}

	// Create session
	token, err := h.AuthService.CreateSession(user.ID, c.Request.UserAgent(), c.ClientIP())
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, h.Config.FrontendURL+"/login?error=session_error")
		return
	}

	// Set session cookie
	h.setSessionCookie(c, token)

	// Redirect to frontend
	c.Redirect(http.StatusTemporaryRedirect, h.Config.FrontendURL+"/auth/callback")
}

// GitHubLogin initiates GitHub OAuth flow
// GET /api/auth/github
func (h *AuthHandler) GitHubLogin(c *gin.Context) {
	if !h.OAuthService.IsGitHubEnabled() {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "GitHub OAuth not configured"})
		return
	}

	state, err := generateOAuthState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate state"})
		return
	}

	h.setOAuthStateCookie(c, state)

	url, err := h.OAuthService.GetGitHubAuthURL(state)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate auth URL"})
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GitHubCallback handles the GitHub OAuth callback
// GET /api/auth/github/callback
func (h *AuthHandler) GitHubCallback(c *gin.Context) {
	// Verify state parameter
	state := c.Query("state")
	storedState, err := c.Cookie("oauth_state")
	if err != nil || state != storedState {
		c.Redirect(http.StatusTemporaryRedirect, h.Config.FrontendURL+"/login?error=invalid_state")
		return
	}
	h.clearOAuthStateCookie(c)

	// Check for error from OAuth provider
	if errParam := c.Query("error"); errParam != "" {
		c.Redirect(http.StatusTemporaryRedirect, h.Config.FrontendURL+"/login?error="+errParam)
		return
	}

	// Exchange code for user info
	code := c.Query("code")
	githubUser, err := h.OAuthService.ExchangeGitHubCode(c.Request.Context(), code)
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, h.Config.FrontendURL+"/login?error=oauth_failed")
		return
	}

	// GitHub ID is an int64, convert to string
	githubID := fmt.Sprintf("%d", githubUser.ID)

	// Use login as name if name is empty
	name := githubUser.Name
	if name == "" {
		name = githubUser.Login
	}

	// Find or create user (credits initialized atomically for new users)
	user, _, err := h.AuthService.FindOrCreateOAuthUser(
		"github",
		githubID,
		githubUser.Email,
		name,
		githubUser.AvatarURL,
		h.CreditsService,
	)
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, h.Config.FrontendURL+"/login?error=account_error")
		return
	}

	// Create session
	token, err := h.AuthService.CreateSession(user.ID, c.Request.UserAgent(), c.ClientIP())
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, h.Config.FrontendURL+"/login?error=session_error")
		return
	}

	// Set session cookie
	h.setSessionCookie(c, token)

	// Redirect to frontend
	c.Redirect(http.StatusTemporaryRedirect, h.Config.FrontendURL+"/auth/callback")
}
