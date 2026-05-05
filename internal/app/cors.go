package app

import "github.com/gin-gonic/gin"

func SetCors(c *gin.Context) {
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*") // TODO: adicionar somente o dominio da aplicação
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	// Preflight request
	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(204)
		return
	}

	c.Next()
}
