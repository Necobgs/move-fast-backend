package handler

import (
	"github.com/Necobgs/move-fast-backend/internal/auth"
	"github.com/Necobgs/move-fast-backend/internal/db/sqlc"
	"github.com/Necobgs/move-fast-backend/internal/dto"
	"github.com/Necobgs/move-fast-backend/internal/middleware"
	"github.com/Necobgs/move-fast-backend/internal/response"
	"github.com/Necobgs/move-fast-backend/internal/service"
	"github.com/gin-gonic/gin"
)

type RideHandler struct {
	repository *sqlc.Queries
	service    *service.RideService
}

func NewRideHandler(repository *sqlc.Queries, service *service.RideService) *RideHandler {
	return &RideHandler{
		repository: repository,
		service:    service,
	}
}

func (h *RideHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rgRide := rg.Group("/ride").Use(middleware.AuthMiddleware())
	{
		rgRide.POST("", h.CreateRide)
		rgRide.POST("/handle-ride-request", h.HandleRideRequest)
		rgRide.POST("/:id/start", h.StartRide)
		rgRide.POST("/:id/finish", h.FinishRide)
		rgRide.POST("/:id/cancel", h.CancelRide)
		rgRide.GET("/history", h.GetRideHistory)
		rgRide.GET("/active", h.GetActiveRide)
	}
}

func (h *RideHandler) CreateRide(c *gin.Context) {

	var createRide dto.CreateRideDto
	if err := c.ShouldBindJSON(&createRide); err != nil {
		c.JSON(response.ErrInvalidBody.StatusCode, response.ErrInvalidBody)
		return
	}

	value, _ := c.Get("user")
	claims := value.(*auth.CustomClaims)

	ride, errCreateRide := h.service.CreateRide(createRide, claims)
	if errCreateRide != nil {
		c.JSON(errCreateRide.StatusCode, errCreateRide)
		return
	}

	c.JSON(201, ride)

}

func (h *RideHandler) HandleRideRequest(c *gin.Context) {
	var rideRequestDto dto.HandleRideRequestDto
	if err := c.ShouldBindJSON(&rideRequestDto); err != nil {
		c.JSON(response.ErrInvalidBody.StatusCode, response.ErrInvalidBody)
		return
	}

	value, _ := c.Get("user")
	claims := value.(*auth.CustomClaims)

	ride, errHandleRideRequest := h.service.HandleRideRequest(rideRequestDto, claims)
	if errHandleRideRequest != nil {
		c.JSON(errHandleRideRequest.StatusCode, errHandleRideRequest)
		return
	}

	c.JSON(200, ride)

}

func (h *RideHandler) StartRide(c *gin.Context) {
	rideID := c.Param("id")
	value, _ := c.Get("user")
	claims := value.(*auth.CustomClaims)

	ride, errResp := h.service.StartRide(rideID, claims)
	if errResp != nil {
		c.JSON(errResp.StatusCode, errResp)
		return
	}
	c.JSON(200, ride)
}

func (h *RideHandler) FinishRide(c *gin.Context) {
	rideID := c.Param("id")
	value, _ := c.Get("user")
	claims := value.(*auth.CustomClaims)

	ride, errResp := h.service.FinishRide(rideID, claims)
	if errResp != nil {
		c.JSON(errResp.StatusCode, errResp)
		return
	}
	c.JSON(200, ride)
}

func (h *RideHandler) CancelRide(c *gin.Context) {
	rideID := c.Param("id")
	value, _ := c.Get("user")
	claims := value.(*auth.CustomClaims)

	ride, errResp := h.service.CancelRide(rideID, claims)
	if errResp != nil {
		c.JSON(errResp.StatusCode, errResp)
		return
	}
	c.JSON(200, ride)
}

func (h *RideHandler) GetRideHistory(c *gin.Context) {
	cursorStr := c.Query("cursor")

	// Optional: dá pra extrair da queryParam o limit
	// limitStr := c.Query("limit")
	var limit int32 = 10

	value, _ := c.Get("user")
	claims := value.(*auth.CustomClaims)

	rides, nextCursor, errResp := h.service.GetRideHistory(claims, cursorStr, limit)
	if errResp != nil {
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	c.JSON(200, gin.H{
		"data":        rides,
		"next_cursor": nextCursor,
	})
}

func (h *RideHandler) GetActiveRide(c *gin.Context) {
	value, _ := c.Get("user")
	claims := value.(*auth.CustomClaims)

	ride, errResp := h.service.GetActiveRide(claims)
	if errResp != nil {
		c.JSON(errResp.StatusCode, errResp)
		return
	}
	c.JSON(200, ride)
}
