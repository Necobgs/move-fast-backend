package service

import (
	"context"

	"github.com/Necobgs/move-fast-backend/internal/db/sqlc"
	"github.com/Necobgs/move-fast-backend/internal/response"
)

type RideService struct {
	repository *sqlc.Queries
}

func (s *RideService) NewRideService(repository *sqlc.Queries) *RideService {
	return &RideService{repository: repository}
}

func (s *RideService) CreateRide(params sqlc.CreateRideParams) (*sqlc.CreateRideRow, *response.ErrorResponse) {
	ctx := context.Background()
	ride, err := s.repository.CreateRide(ctx, params)
	if err != nil {
		return nil, &response.ErrInternalServer
	}

	return ride, nil
}
