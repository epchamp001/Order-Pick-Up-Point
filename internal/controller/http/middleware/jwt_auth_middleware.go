package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"order-pick-up-point/internal/models/dto"
	"order-pick-up-point/pkg/jwt"
	"strings"
)

func JWTAuthMiddleware(tokenSvc jwt.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, dto.Error{
				Message: "missing token",
			})
			c.Abort()
			return
		}

		token = strings.TrimPrefix(token, "Bearer ")

		claims, err := tokenSvc.ParseJWTToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, dto.Error{
				Message: "invalid token",
			})
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("role", claims.Role)

		c.Next()
	}
}
