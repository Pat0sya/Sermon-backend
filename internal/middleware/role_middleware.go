package middleware

import (
	"net/http"

	"sermon-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		currentRole := c.GetString("role")

		for _, role := range roles {
			if currentRole == role {
				c.Next()
				return
			}
		}

		utils.Error(c, http.StatusForbidden, "forbidden")
		c.Abort()
	}
}
