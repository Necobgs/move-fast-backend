package handler

import (
	"github.com/Necobgs/move-fast-backend/internal/auth"
	"github.com/Necobgs/move-fast-backend/internal/middleware"
	"github.com/Necobgs/move-fast-backend/internal/response"
	"github.com/Necobgs/move-fast-backend/internal/service"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	UserService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{UserService: userService}
}

func (h *UserHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rgUsers := rg.Group("/user/").
		Use(middleware.AuthMiddleware())
	{
		rgUsers.GET("/me", h.GetMe)
	}
}

func (h *UserHandler) GetMe(c *gin.Context) {
	value, exists := c.Get("user")
	if !exists {
		c.JSON(response.ErrUnauthorized.StatusCode, response.ErrUnauthorized)
		return
	}
	claims := value.(*auth.CustomClaims)
	user, err := h.UserService.GetMe(claims.Email)
	if err != nil {
		c.JSON(err.StatusCode, err)
		return
	}
	c.JSON(200, user)
}
