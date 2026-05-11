package service

import (
	"context"

	"github.com/Necobgs/move-fast-backend/internal/db/sqlc"
	"github.com/Necobgs/move-fast-backend/internal/dto"
	"github.com/Necobgs/move-fast-backend/internal/models"
	"github.com/Necobgs/move-fast-backend/internal/response"
	"github.com/Necobgs/move-fast-backend/pkg/logger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DriverService struct {
	repository  *sqlc.Queries
	db          *pgxpool.Pool
	authService *AuthService
}

func NewDriverService(repository *sqlc.Queries, db *pgxpool.Pool) *DriverService {
	return &DriverService{repository: repository, db: db}
}

func (s *DriverService) CreateDriver(createDriverDto dto.CreateDriverDto, userId uuid.UUID) (*models.DriverResponse, *response.ErrorResponse) {
	ctx := context.Background()
	tx, err := s.db.Begin(ctx)
	if err != nil {
		logger.Log.Error("Erro na transaction do create driver", "error", err)
		return nil, &response.ErrInternalServer
	}
	defer tx.Rollback(ctx)

	qtx := s.repository.WithTx(tx)

	exists, _ := qtx.ExistsVehicle(ctx, createDriverDto.VehicleLicensePlate)
	if exists {
		return nil, &response.ErrVehicleAlreadyExists
	}

	exists, _ = qtx.ExistsDriver(ctx, sqlc.ExistsDriverParams{
		UserID:    userId,
		CnhNumber: createDriverDto.CnhNumber,
	})

	if exists {
		return nil, &response.ErrDriverAlreadyExists
	}

	vehicle, err := qtx.CreateVehicle(ctx, sqlc.CreateVehicleParams{
		ID:           uuid.New(),
		LicensePlate: createDriverDto.VehicleLicensePlate,
		Model:        createDriverDto.VehicleModel,
		Brand:        createDriverDto.VehicleBrand,
		Color:        createDriverDto.VehicleColor,
	})

	if err != nil {
		logger.Log.Error("Erro ao criar vehicle", "error", err)
		return nil, &response.ErrInternalServer
	}

	driver, err := qtx.CreateDriver(ctx, sqlc.CreateDriverParams{
		ID:        uuid.New(),
		CnhNumber: createDriverDto.CnhNumber,
		UserID:    userId,
		VehicleID: vehicle.ID,
	})
	if err != nil {
		logger.Log.Error("Erro ao criar driver no banco de dados", "error", err)
		return nil, &response.ErrInternalServer
	}

	user, err := qtx.FindUser(ctx, sqlc.FindUserParams{
		ID: userId,
	})
	if err != nil {
		logger.Log.Error("Erro ao buscar user para token", "error", err)
		return nil, &response.ErrInternalServer
	}

	token, err := s.authService.GenerateToken(user)
	if err != nil {
		logger.Log.Error("Erro ao gerar token para driver", "error", err)
		return nil, &response.ErrInternalServer
	}

	if err := tx.Commit(ctx); err != nil {
		logger.Log.Error("Erro no commit da transaction", "error", err)
		return nil, &response.ErrInternalServer
	}

	driverResponse := models.DriverResponse{
		Driver: sqlc.Driver{
			ID:        driver.ID,
			CnhNumber: driver.CnhNumber,
			UserID:    userId,
			VehicleID: vehicle.ID,
		},
		Vehicle: sqlc.Vehicle{
			ID:           vehicle.ID,
			LicensePlate: vehicle.LicensePlate,
			Model:        vehicle.Model,
			Brand:        vehicle.Brand,
			Color:        vehicle.Color,
		},
		Token: token,
	}

	return &driverResponse, nil
}
