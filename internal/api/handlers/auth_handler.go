package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/bernardofernandezz/scheduling-api/internal/models"
	"github.com/bernardofernandezz/scheduling-api/internal/service"
	"github.com/bernardofernandezz/scheduling-api/pkg/auth"
)

// AuthHandler handles authentication related requests
type AuthHandler struct {
	userService service.UserService
	jwtManager  *auth.JWTManager
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(userService service.UserService, jwtManager *auth.JWTManager) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		jwtManager:  jwtManager,
	}
}

// RegisterRequest is the request body for user registration
type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Role     string `json:"role" binding:"required,oneof=admin employee supplier"`
	Phone    string `json:"phone" binding:"required"`
}

// LoginRequest is the request body for user login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RefreshTokenRequest is the request body for token refresh
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// PasswordResetRequest is the request body for password reset
type PasswordResetRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// PasswordChangeRequest is the request body for changing password
type PasswordChangeRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// AuthResponse is the response body for authentication
type AuthResponse struct {
	User         *models.User `json:"user"`
	Token        string       `json:"token"`
	RefreshToken string       `json:"refresh_token,omitempty"`
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Create user model from request
	user := &models.User{
		Name:         req.Name,
		Email:        strings.ToLower(req.Email),
		PasswordHash: req.Password, // This will be hashed in the service
		Role:         req.Role,
		Phone:        req.Phone,
		Active:       true,
	}

	// Register user
	if err := h.userService.Register(user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate tokens
	token, err := h.jwtManager.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	refreshToken, err := h.jwtManager.GenerateRefreshToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	// Set password to empty for response
	user.PasswordHash = ""

	c.JSON(http.StatusCreated, AuthResponse{
		User:         user,
		Token:        token,
		RefreshToken: refreshToken,
	})
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(

