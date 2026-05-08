package handler

import (
	"github.com/Necobgs/move-fast-backend/internal/auth"
	"github.com/Necobgs/move-fast-backend/internal/dto"
	"github.com/Necobgs/move-fast-backend/internal/middleware"
	"github.com/Necobgs/move-fast-backend/internal/response"
	"github.com/Necobgs/move-fast-backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type DriverHandler struct {
	service *service.DriverService
}

func NewDriverHandler(service *service.DriverService) *DriverHandler {
	return &DriverHandler{service: service}
}

func (h *DriverHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rgDriver := rg.Group("/driver").Use(middleware.AuthMiddleware())
	{
		rgDriver.POST("", h.CreateDriver)
	}
}

func (h *DriverHandler) CreateDriver(c *gin.Context) {
	var createDriver dto.CreateDriverDto
	err := c.BindJSON(&createDriver)
	if err != nil {
		c.JSON(response.ErrInternalServer.StatusCode, response.ErrInternalServer)
		return
	}

	value, exists := c.Get("user")
	if !exists {
		c.JSON(response.ErrUnauthorized.StatusCode, response.ErrUnauthorized)
		return
	}

	claims := value.(*auth.CustomClaims)
	userIdParsed, err := uuid.Parse(claims.Id)
	if err != nil {
		c.JSON(response.ErrInternalServer.StatusCode, response.ErrInternalServer)
		return
	}

	driver, errCreateDriver := h.service.CreateDriver(createDriver, userIdParsed)
	if errCreateDriver != nil {
		c.JSON(errCreateDriver.StatusCode, errCreateDriver)
		return
	}

	c.JSON(201, driver)
}
