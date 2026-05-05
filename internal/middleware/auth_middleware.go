package middleware

import (
	"strings"

	"github.com/Necobgs/move-fast-backend/configs"
	"github.com/Necobgs/move-fast-backend/internal/auth"
	"github.com/Necobgs/move-fast-backend/internal/response"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func AuthMiddleware() gin.HandlerFunc {

	secret := configs.GetConfig().JWTSecretKey

	return func(c *gin.Context) {

		authorization := c.GetHeader("Authorization")
		authSplited := strings.Split(authorization, " ")

		if len(authSplited) != 2 || authSplited[0] != "Bearer" {
			c.AbortWithStatus(response.ErrUnauthorized.StatusCode)
			return
		}

		tokenString := authSplited[1]
		token, err := jwt.ParseWithClaims(tokenString, &auth.CustomClaims{}, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); ok {
				return []byte(secret), nil
			}
			return nil, jwt.ErrSignatureInvalid
		})

		if err != nil || !token.Valid {
			c.AbortWithStatus(401)
		}

		if claims, ok := token.Claims.(*auth.CustomClaims); ok {
			c.Set("user", claims)
		} else {
			c.AbortWithStatus(401)
		}

	}
}
