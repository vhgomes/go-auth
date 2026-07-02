package handler

import (
	"auth-go/internal/config"
	"auth-go/internal/service"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *service.UserService
	cfg         *config.Config
}

func NewUserHandler(userService *service.UserService, cfg *config.Config) *UserHandler {
	return &UserHandler{userService: userService, cfg: cfg}
}

func (h *UserHandler) RegisterUser(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	if len(username) < 8 || len(password) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username or password invalid"})
		return
	}

	err := h.userService.RegisterUser(c.Request.Context(), username, password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

func (h *UserHandler) LoginUser(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	if len(username) < 8 || len(password) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username or password invalid"})
		return
	}

	sessionToken, err := h.userService.LoginUser(c.Request.Context(), username, password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.SetSameSite(http.SameSiteLaxMode) // precisa ser chamado antes do SetCookie
	c.SetCookie(
		"session_token",
		sessionToken,
		3600,
		"/",
		"",
		h.cfg.CookieSecure, // vem de env: true em produção, false em dev local
		true,               // httpOnly continua true
	)

	c.JSON(http.StatusOK, gin.H{"message": "User logged in successfully"})
}

func (h *UserHandler) Logout(c *gin.Context) {
	sessionToken, err := c.Cookie("session_token")
	if err != nil || sessionToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Session cookie is empty"})
		return
	}

	log.Printf("Session Token: %s", sessionToken)

	_, err = h.userService.LogoutUser(c.Request.Context(), sessionToken)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.SetCookie(
		"session_token", // Cookie name
		"",              // Empty value
		-1,              // Max age: -1 (expire immediately)
		"/",             // Path
		"",              // Domain
		false,           // Secure
		true,            // HttpOnly
	)

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}
