package handler

import (
	"auth-go/internal/service"
	"github.com/gin-gonic/gin"
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
