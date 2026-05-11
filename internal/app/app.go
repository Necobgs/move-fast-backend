package app

import (
	"context"

	"github.com/Necobgs/move-fast-backend/configs"
	"github.com/Necobgs/move-fast-backend/internal/db/sqlc"
	"github.com/Necobgs/move-fast-backend/internal/handler"
	"github.com/Necobgs/move-fast-backend/internal/service"
	"github.com/Necobgs/move-fast-backend/internal/ws"
	"github.com/Necobgs/move-fast-backend/pkg/logger"
	"github.com/gin-gonic/gin"
)

func Bootstrap() {
	logger.InitLogger()

	cfg := configs.LoadConfig("./")
	ctx := context.Background()

	dbConn := ConnectDatabase(ctx, cfg.DBUrl)
	defer dbConn.Close()

	queries := sqlc.New(dbConn)

	rdbc := ConnectRedis(&ctx)

	// Dependence Injection
	userService := service.NewUserService(queries)
	authService := service.NewAuthService(queries)
	driverService := service.NewDriverService(queries, dbConn)

	wsHub := ws.NewHub()
	eventRegistry := ws.NewEventRegistry()
	go wsHub.Run()

	userHandler := handler.NewUserHandler(userService)
	authHandler := handler.NewAuthHandler(authService, userService)
	driverHandler := handler.NewDriverHandler(driverService)
	eventHandler := handler.NewEventHandler(queries, rdbc, wsHub)
	wsHandler := ws.NewWsHandler(rdbc, authService, wsHub, eventRegistry, eventHandler)

	rideService := service.NewRideService(queries, eventHandler)
	rideHandler := handler.NewRideHandler(queries, rideService)

	// Events WebSocket
	eventRegistry.Register("update_location_driver", eventHandler.UpdateLocationDriver) // Atualizar localização do motorista

	// HTTP
	gin.SetMode(cfg.GinMode)
	route := gin.Default()
	route.Use(SetCors)
	route.SetTrustedProxies(nil)

	route.GET("/ws", wsHandler.HandleConnection)

	api := route.Group("/api")

	route.Static(cfg.UrlUploads, cfg.PathUploads)

	authHandler.RegisterRoutes(api)
	userHandler.RegisterRoutes(api)
	driverHandler.RegisterRoutes(api)
	rideHandler.RegisterRoutes(api)

	logger.Log.Info("Server is running", "url", cfg.WebServerUrl)
	route.Run(cfg.WebServerUrl)
}
