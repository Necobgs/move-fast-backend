package handler

import (
	"github.com/Necobgs/move-fast-backend/internal/dto"
	"github.com/Necobgs/move-fast-backend/internal/response"
	"github.com/Necobgs/move-fast-backend/internal/service"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service     *service.AuthService
	userService *service.UserService
}

func NewAuthHandler(service *service.AuthService, userService *service.UserService) *AuthHandler {
	return &AuthHandler{service: service, userService: userService}
}

func (h *AuthHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rgAuth := rg.Group("/auth")
	{
		rgAuth.POST("/signin", h.Signin)
		rgAuth.POST("/signup", h.Signup)
	}
}

func (h *AuthHandler) Signup(c *gin.Context) {
	var createUserDto dto.CreateUserDTO
	errBinding := c.ShouldBind(&createUserDto)
	if errBinding != nil {
		c.JSON(response.ErrInternalServer.StatusCode, response.ErrInternalServer)
		return
	}

	user, err := h.userService.CreateUser(&createUserDto, c)
	if err != nil {
		c.JSON(err.StatusCode, err)
		return
	}
	c.JSON(201, user)
}

func (h *AuthHandler) Signin(c *gin.Context) {
	var signinDto dto.SigninDto
	bindError := c.ShouldBindJSON(&signinDto)
	if bindError != nil {
		c.JSON(response.ErrInternalServer.StatusCode, response.ErrInternalServer)
		return
	}

	signinResponseDto, err := h.service.Signin(signinDto)
	if err != nil {
		c.JSON(err.StatusCode, err)
		return
	}

	c.JSON(200, signinResponseDto)
}
