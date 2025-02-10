package handler

import (
	"auth-go/internal/service"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) RegisterUser(c *gin.Context) {
	if c.Request.Method != http.MethodPost {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Method not allowed"})
		return
	}

	username := c.PostForm("username")
	password := c.PostForm("password")

	if len(username) < 8 || len(password) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username or password invalid"})
		return
	}

	err := h.userService.RegisterUser(username, password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

func (h *UserHandler) LoginUser(c *gin.Context) {
	if c.Request.Method != http.MethodPost {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Method not allowed"})
		return
	}

	username := c.PostForm("username")
	password := c.PostForm("password")

	if len(username) < 8 || len(password) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username or password invalid"})
		return
	}

	sessionToken, err := h.userService.LoginUser(username, password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.SetCookie(
		"session_token", // Cookie name
		sessionToken,    // Cookie value (the session token)
		3600,            // Max age in seconds (e.g., 1 hour)
		"/",             // Path
		"",              // Domain (leave empty for current domain)
		false,           // Secure (set to true if using HTTPS)
		true,            // HttpOnly (prevents client-side JavaScript from accessing the cookie)
	)

	c.JSON(http.StatusCreated, gin.H{"message": "User logged in successfully"})
}

func (h *UserHandler) Logout(c *gin.Context) {
	if c.Request.Method != http.MethodGet {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Method not allowed"})
		return
	}

	sessionToken, err := c.Cookie("session_token")
	if err != nil || sessionToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Session cookie is empty"})
		return
	}

	log.Printf("Session Token: %s", sessionToken)

	// err = h.userService.LogoutUser(sessionToken)
	_, err = h.userService.LogoutUser(sessionToken)

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
