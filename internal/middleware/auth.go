package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SessionValidator interface {
	GetUserIDByToken(ctx context.Context, sessionToken string) (string, error)
}

func RequireAuth(sessions SessionValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionToken, err := c.Cookie("session_token")
		if err != nil || sessionToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
			c.Abort()
			return
		}

		userID, err := sessions.GetUserIDByToken(c.Request.Context(), sessionToken)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate session"})
			c.Abort()
			return
		}

		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "session expired or invalid"})
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}
