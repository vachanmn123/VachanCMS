package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vachanmn123/vachancms/config"
	"github.com/vachanmn123/vachancms/services"
)

func AuthMiddleware(c *gin.Context) {
	token, err := c.Cookie("auth_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		c.Abort()
		return
	}

	claims, err := services.ValidateJWT(token, config.Cfg)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		c.Abort()
		return
	}

	c.Set("user_id", claims.UserID)
	c.Set("user_access_token", claims.UserAccessToken)
	c.Next()
}
