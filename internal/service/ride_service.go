package service

import (
	"context"
	"time"

	"github.com/Necobgs/move-fast-backend/internal/auth"
	"github.com/Necobgs/move-fast-backend/internal/consts"
	"github.com/Necobgs/move-fast-backend/internal/db/sqlc"
	"github.com/Necobgs/move-fast-backend/internal/dto"
	"github.com/Necobgs/move-fast-backend/internal/response"
	"github.com/Necobgs/move-fast-backend/pkg/logger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type RideEventDispatcher interface {
	DispatchRideRequest(ctx context.Context, ride *sqlc.CreateRideRow) error
	DispatchRideAccepted(ctx context.Context, ride *sqlc.UpdateRideRow) error
	DispatchRideCanceled(ctx context.Context, rideID string, passengerID string) error
	DispatchRideStatusChanged(ctx context.Context, rideID string, recipientID string, status string) error
}

type RideService struct {
	repository      *sqlc.Queries
	eventDispatcher RideEventDispatcher
}

func NewRideService(repository *sqlc.Queries, eventDispatcher RideEventDispatcher) *RideService {
	return &RideService{
		repository:      repository,
		eventDispatcher: eventDispatcher,
	}
}

func (s *RideService) CreateRide(createRideDto dto.CreateRideDto, claims *auth.CustomClaims) (*sqlc.CreateRideRow, *response.ErrorResponse) {
	ctx := context.Background()

	passengerID := uuid.MustParse(claims.Id)
	exists, err := s.repository.ExistsActiveRidesForPassenger(ctx, passengerID)
	if err == nil && exists {
		return nil, &response.ErrorResponse{Message: "Você já possui uma corrida em andamento.", StatusCode: 400}
	}

	ride, err := s.repository.CreateRide(
		ctx,
		sqlc.CreateRideParams{
			ID:          uuid.New(),
			PassengerID: passengerID,

			OriginLat:     createRideDto.OriginLocationLat,
			OriginLng:     createRideDto.OriginLocationLng,
			OriginAddress: createRideDto.OriginLocationAddress,

			DestinationLat:     createRideDto.DestinationLocationLat,
			DestinationLng:     createRideDto.DestinationLocationLng,
			DestinationAddress: createRideDto.DestinationLocationAddress,
		},
	)
	if err != nil {
		return nil, &response.ErrorResponse{Message: "Não foi possível criar uma nova corrida", StatusCode: 500}
	}

	errEvent := s.eventDispatcher.DispatchRideRequest(ctx, ride)
	if errEvent != nil {
		// Cancela a corrida se não achar motorista
		s.repository.UpdateRide(ctx, sqlc.UpdateRideParams{
			ID:       ride.ID,
			StatusID: consts.StatusCanceledRideId,
		})
		return nil, &response.ErrorResponse{Message: "Nenhum motorista encontrado próximo a você", StatusCode: 404}
	}

	go func() {
		time.Sleep(30 * time.Second)
		bgCtx := context.Background()

		rideStatus, err := s.repository.GetRideById(bgCtx, ride.ID)
		if err != nil {
			return
		}

		if rideStatus.StatusID == consts.StatusRequestedRideId {
			s.repository.UpdateRide(bgCtx, sqlc.UpdateRideParams{
				ID:       ride.ID,
				StatusID: consts.StatusCanceledRideId,
			})
			s.eventDispatcher.DispatchRideCanceled(bgCtx, ride.ID.String(), ride.PassengerID.String())
		}
	}()

	return ride, nil
}

func (s *RideService) HandleRideRequest(handleRideRequestDto dto.HandleRideRequestDto, claims *auth.CustomClaims) (*sqlc.UpdateRideRow, *response.ErrorResponse) {
	ctx := context.Background()

	driverIdParsed, err := uuid.Parse(claims.DriverId)
	if err != nil {
		return nil, &response.ErrUnauthorized
	}

	var statusID string
	if handleRideRequestDto.Accepted {
		exists, err := s.repository.ExistsActiveRidesForDriver(ctx, pgtype.UUID{Bytes: driverIdParsed, Valid: true})
		if err == nil && exists {
			return nil, &response.ErrorResponse{Message: "Você já possui uma corrida em andamento.", StatusCode: 400}
		}
		statusID = consts.StatusWaitingDriverId
	} else {
		statusID = consts.StatusCanceledRideId
	}

	rideID, err := uuid.Parse(handleRideRequestDto.RideId)
	if err != nil {
		return nil, &response.ErrorResponse{Message: "Corrida não encontrada", StatusCode: 404}
	}

	ride, err := s.repository.UpdateRide(ctx, sqlc.UpdateRideParams{
		ID: rideID,
		DriverID: pgtype.UUID{
			Bytes: driverIdParsed,
			Valid: true,
		},
		StatusID: statusID,
	})

	if err != nil {
		return nil, &response.ErrorResponse{Message: "Não foi possível aceitar/recusar a corrida", StatusCode: 500}
	}

	if handleRideRequestDto.Accepted {
		s.eventDispatcher.DispatchRideAccepted(ctx, ride)
	}

	return ride, nil
}

func (s *RideService) StartRide(rideIDStr string, claims *auth.CustomClaims) (*sqlc.StartRideQueryRow, *response.ErrorResponse) {
	ctx := context.Background()

	rideID, err := uuid.Parse(rideIDStr)
	if err != nil {
		return nil, &response.ErrorResponse{Message: "ID de corrida inválido", StatusCode: 400}
	}

	driverIdParsed, err := uuid.Parse(claims.DriverId)
	if err != nil {
		return nil, &response.ErrUnauthorized
	}

	ride, err := s.repository.GetRideById(ctx, rideID)
	if err != nil {
		return nil, &response.ErrorResponse{Message: "Corrida não encontrada", StatusCode: 404}
	}

	if ride.DriverID.Bytes != driverIdParsed {
		return nil, &response.ErrorResponse{Message: "Não autorizado", StatusCode: 403}
	}

	if ride.StatusID != consts.StatusWaitingDriverId {
		return nil, &response.ErrorResponse{Message: "A corrida não pode ser iniciada a partir deste status", StatusCode: 400}
	}

	updatedRide, err := s.repository.StartRideQuery(ctx, sqlc.StartRideQueryParams{
		ID:       rideID,
		StatusID: consts.StatusStartedRideId,
	})
	if err != nil {
		return nil, &response.ErrorResponse{Message: "Erro ao iniciar corrida", StatusCode: 500}
	}

	s.eventDispatcher.DispatchRideStatusChanged(ctx, rideID.String(), ride.PassengerID.String(), consts.StatusStartedRideId)

	return updatedRide, nil
}

func (s *RideService) FinishRide(rideIDStr string, claims *auth.CustomClaims) (*sqlc.FinishRideQueryRow, *response.ErrorResponse) {
	ctx := context.Background()

	rideID, err := uuid.Parse(rideIDStr)
	if err != nil {
		return nil, &response.ErrorResponse{Message: "ID de corrida inválido", StatusCode: 400}
	}

	driverIdParsed, err := uuid.Parse(claims.DriverId)
	if err != nil {
		return nil, &response.ErrUnauthorized
	}

	ride, err := s.repository.GetRideById(ctx, rideID)
	if err != nil {
		return nil, &response.ErrorResponse{Message: "Corrida não encontrada", StatusCode: 404}
	}

	if ride.DriverID.Bytes != driverIdParsed {
		return nil, &response.ErrorResponse{Message: "Não autorizado", StatusCode: 403}
	}

	if ride.StatusID != consts.StatusStartedRideId {
		return nil, &response.ErrorResponse{Message: "A corrida não pode ser finalizada a partir deste status", StatusCode: 400}
	}

	updatedRide, err := s.repository.FinishRideQuery(ctx, sqlc.FinishRideQueryParams{
		ID:       rideID,
		StatusID: consts.StatusEndedRideId,
	})
	if err != nil {
		return nil, &response.ErrorResponse{Message: "Erro ao finalizar corrida", StatusCode: 500}
	}

	s.eventDispatcher.DispatchRideStatusChanged(ctx, rideID.String(), ride.PassengerID.String(), consts.StatusEndedRideId)

	return updatedRide, nil
}

func (s *RideService) CancelRide(rideIDStr string, claims *auth.CustomClaims) (*sqlc.UpdateRideRow, *response.ErrorResponse) {
	ctx := context.Background()

	rideID, err := uuid.Parse(rideIDStr)
	if err != nil {
		return nil, &response.ErrorResponse{Message: "ID de corrida inválido", StatusCode: 400}
	}

	ride, err := s.repository.GetRideById(ctx, rideID)
	if err != nil {
		return nil, &response.ErrorResponse{Message: "Corrida não encontrada", StatusCode: 404}
	}

	if ride.StatusID == consts.StatusEndedRideId || ride.StatusID == consts.StatusCanceledRideId {
		return nil, &response.ErrorResponse{Message: "A corrida não pode ser cancelada a partir deste status", StatusCode: 400}
	}

	isPassenger := claims.Id == ride.PassengerID.String()
	isDriver := claims.DriverId != "" && ride.DriverID.Valid && claims.DriverId == uuid.UUID(ride.DriverID.Bytes).String()

	if !isPassenger && !isDriver {
		return nil, &response.ErrorResponse{Message: "Não autorizado a cancelar esta corrida", StatusCode: 403}
	}

	updatedRide, err := s.repository.UpdateRide(ctx, sqlc.UpdateRideParams{
		ID:       rideID,
		StatusID: consts.StatusCanceledRideId,
	})
	if err != nil {
		return nil, &response.ErrorResponse{Message: "Erro ao cancelar corrida", StatusCode: 500}
	}

	if isPassenger && ride.DriverID.Valid {
		s.eventDispatcher.DispatchRideStatusChanged(ctx, rideID.String(), uuid.UUID(ride.DriverID.Bytes).String(), consts.StatusCanceledRideId)
	} else if isDriver {
		s.eventDispatcher.DispatchRideStatusChanged(ctx, rideID.String(), ride.PassengerID.String(), consts.StatusCanceledRideId)
	}

	return updatedRide, nil
}

func (s *RideService) GetRideHistory(claims *auth.CustomClaims, cursorStr string, limit int32) ([]*sqlc.GetRideHistoryRow, string, *response.ErrorResponse) {
	ctx := context.Background()

	passengerID, err := uuid.Parse(claims.Id)
	if err != nil {
		return nil, "", &response.ErrUnauthorized
	}

	var driverID pgtype.UUID
	if claims.DriverId != "" {
		driverUUID, err := uuid.Parse(claims.DriverId)
		if err == nil {
			driverID = pgtype.UUID{Bytes: driverUUID, Valid: true}
		}
	}

	var cursor time.Time
	if cursorStr != "" {
		parsedCursor, err := time.Parse(time.RFC3339Nano, cursorStr)
		if err != nil {
			return nil, "", &response.ErrorResponse{Message: "Cursor inválido", StatusCode: 400}
		}
		cursor = parsedCursor
	} else {
		cursor = time.Now()
	}

	if limit <= 0 || limit > 50 {
		limit = 10
	}

	rides, err := s.repository.GetRideHistory(ctx, sqlc.GetRideHistoryParams{
		PassengerID: passengerID,
		DriverID:    driverID,
		CreatedAt:   pgtype.Timestamptz{Time: cursor, Valid: true},
		Limit:       limit,
	})

	if err != nil {
		logger.Log.Error("Erro ao buscar histórico:", "error", err)
		return nil, "", &response.ErrorResponse{Message: "Erro ao buscar histórico de corridas", StatusCode: 500}
	}

	var nextCursor string
	if len(rides) > 0 {
		lastRide := rides[len(rides)-1]
		nextCursor = lastRide.CreatedAt.Time.Format(time.RFC3339Nano)
	}

	return rides, nextCursor, nil
}

func (s *RideService) GetActiveRide(claims *auth.CustomClaims) (*sqlc.Ride, *response.ErrorResponse) {
	ctx := context.Background()

	if claims.DriverId != "" {
		driverID, err := uuid.Parse(claims.DriverId)
		if err != nil {
			return nil, &response.ErrUnauthorized
		}

		ride, err := s.repository.GetRideInProgessFromDriver(ctx, pgtype.UUID{Bytes: driverID, Valid: true})
		if err != nil {
			return nil, &response.ErrorResponse{Message: "Corrida não encontrada", StatusCode: 404}
		}

		return ride, nil
	}

	passengerID := uuid.MustParse(claims.Id)
	ride, err := s.repository.GetRideInProgressFromPassenger(ctx, passengerID)
	if err != nil {
		return nil, &response.ErrorResponse{Message: "Corrida não encontrada", StatusCode: 404}
	}

	return ride, nil
}
