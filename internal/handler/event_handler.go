package handler

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/Necobgs/move-fast-backend/internal/consts"
	"github.com/Necobgs/move-fast-backend/internal/db/sqlc"
	"github.com/Necobgs/move-fast-backend/internal/realtime"
	"github.com/Necobgs/move-fast-backend/internal/utils"
	"github.com/Necobgs/move-fast-backend/internal/ws/message"
	"github.com/Necobgs/move-fast-backend/pkg/logger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
)

type WsResponse struct {
	Event   string `json:"event"`
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

type EventHandler struct {
	repository *sqlc.Queries
	rdb        *redis.Client
	hub        realtime.Hub

	ridePassengers map[string]string
	mu             sync.RWMutex
}

func NewEventHandler(
	repository *sqlc.Queries,
	rdb *redis.Client,
	hub realtime.Hub,
) *EventHandler {
	handler := &EventHandler{
		repository: repository,
		rdb:        rdb,
		hub:        hub,

		ridePassengers: make(map[string]string),
	}

	go handler.startLocationCleanup()

	return handler
}

func (h *EventHandler) startLocationCleanup() {
	ticker := time.NewTicker(30 * time.Second)
	for range ticker.C {
		ctx := context.Background()

		drivers, err := h.rdb.ZRange(ctx, "driver_locations", 0, -1).Result()
		if err != nil {
			continue
		}

		for _, driverID := range drivers {
			exists, err := h.rdb.Exists(ctx, "driver:lastseen:"+driverID).Result()
			if err != nil {
				continue
			}

			if exists == 0 {
				h.rdb.ZRem(ctx, "driver_locations", driverID)
			}
		}
	}
}

func (h *EventHandler) setRidePassenger(
	rideID string,
	clientID string,
) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.ridePassengers[rideID] = clientID
}

func (h *EventHandler) getRidePassenger(
	rideID string,
) (string, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clientID, ok := h.ridePassengers[rideID]

	return clientID, ok
}

func (h *EventHandler) removeRidePassenger(
	rideID string,
) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.ridePassengers, rideID)
}

func (h *EventHandler) send(
	clientID string,
	response WsResponse,
) bool {

	payload, err := json.Marshal(response)
	if err != nil {
		logger.Log.Error(
			"falha ao serializar response",
			"event", response.Event,
			"error", err,
		)

		return false
	}

	return h.hub.SendToClient(
		clientID,
		payload,
	)
}

func (h *EventHandler) success(
	clientID string,
	event string,
	data any,
) {

	h.send(
		clientID,
		WsResponse{
			Event:   event,
			Success: true,
			Data:    data,
		},
	)
}

func (h *EventHandler) UpdateLocationDriver(
	c *realtime.Client,
	data message.BaseMessage,
) {

	var updateLocation message.UpdateLocMessage

	err := json.Unmarshal(data.Data, &updateLocation)
	if err != nil {

		return
	}

	ctx := context.Background()

	clientKey := utils.BuildClientKey(
		c.Claims.Id,
		c.Claims.DriverId,
	)

	_, err = h.rdb.GeoAdd(
		ctx,
		"driver_locations",
		&redis.GeoLocation{
			Name:      clientKey,
			Longitude: updateLocation.Lng,
			Latitude:  updateLocation.Lat,
		},
	).Result()

	if err != nil {
		return
	}

	h.rdb.Set(ctx, "driver:lastseen:"+c.Claims.DriverId, time.Now().Unix(), time.Minute*2)

	rideID, err := h.repository.GetRideFromDriver(
		ctx,
		sqlc.GetRideFromDriverParams{
			DriverID: pgtype.UUID{
				Bytes: uuid.MustParse(c.Claims.DriverId),
				Valid: true,
			},
			StatusID: consts.StatusWaitingDriverId,
		},
	)

	if err != nil {

		return
	}

	passengerClientID, ok := h.getRidePassenger(
		rideID.String(),
	)

	if !ok {

		logger.Log.Warn(
			"passenger não encontrado",
			"ride_id", rideID.String(),
		)

		return
	}

	sent := h.send(
		passengerClientID,
		WsResponse{
			Event:   "driver_location",
			Success: true,
			Data: map[string]any{
				"driver_id":           c.Claims.DriverId,
				"ride_id":             rideID.String(),
				"driver_location_lat": updateLocation.Lat,
				"driver_location_lng": updateLocation.Lng,
			},
		},
	)

	if !sent {

		logger.Log.Warn(
			"falha ao enviar localização para passenger",
			"ride_id", rideID.String(),
		)

		h.removeRidePassenger(
			rideID.String(),
		)
	}
}

func (h *EventHandler) getNearestDriver(
	lat float64,
	lng float64,
	radius float64,
	ctx context.Context,
) (*string, error) {

	nearestDrivers, err := h.rdb.GeoSearch(
		ctx,
		"driver_locations",
		&redis.GeoSearchQuery{
			Latitude:   lat,
			Longitude:  lng,
			Radius:     radius,
			RadiusUnit: "km",
			Sort:       "ASC",
			Count:      1,
		},
	).Result()

	if err != nil {
		return nil, err
	}

	if len(nearestDrivers) == 0 {
		return nil, errors.New("no driver found")
	}

	return &nearestDrivers[0], nil
}

func (h *EventHandler) getDriverLocation(
	driverID string,
	ctx context.Context,
) (float64, float64, error) {

	driverClientKey := utils.BuildClientKey(
		"",
		driverID,
	)

	driverLocation, err := h.rdb.GeoPos(
		ctx,
		"driver_locations",
		driverClientKey,
	).Result()

	if err != nil {
		return 0, 0, err
	}

	if len(driverLocation) == 0 {
		return 0, 0, errors.New("driver location not found")
	}

	if driverLocation[0] == nil {
		return 0, 0, errors.New("driver location is nil")
	}

	return driverLocation[0].Longitude,
		driverLocation[0].Latitude,
		nil
}

func (h *EventHandler) SyncPassengerConnection(
	client *realtime.Client,
) {

	ctx := context.Background()

	passengerID, err := uuid.Parse(
		client.Claims.Id,
	)

	if err != nil {

		logger.Log.Error(
			"passenger_id inválido",
			"error", err,
		)

		return
	}

	ride, err := h.repository.GetActiveRideFromPassenger(
		ctx,
		passengerID,
	)

	if err != nil {

		return
	}

	clientKey := utils.BuildClientKey(
		client.Claims.Id,
		"",
	)

	h.setRidePassenger(
		ride.ID.String(),
		clientKey,
	)
}

func (h *EventHandler) DispatchRideRequest(ctx context.Context, ride *sqlc.CreateRideRow) error {
	nearestDriverID, err := h.getNearestDriver(
		ride.OriginLat,
		ride.OriginLng,
		10,
		ctx,
	)

	if err != nil {
		return err
	}

	h.success(
		*nearestDriverID,
		"requested_ride",
		map[string]any{
			"ride_id": ride.ID.String(),

			"ride_origin_lat":     ride.OriginLat,
			"ride_origin_lng":     ride.OriginLng,
			"ride_origin_address": ride.OriginAddress,

			"ride_destination_lat":     ride.DestinationLat,
			"ride_destination_lng":     ride.DestinationLng,
			"ride_destination_address": ride.DestinationAddress,

			"ride_passenger_id": ride.PassengerID.String(),
		},
	)
	return nil
}

func (h *EventHandler) DispatchRideAccepted(ctx context.Context, ride *sqlc.UpdateRideRow) error {
	passengerClientKey := utils.BuildClientKey(
		ride.PassengerID.String(),
		"",
	)

	driverID := ride.DriverID.Bytes
	driverUUID := uuid.UUID(driverID)

	driverLng, driverLat, err := h.getDriverLocation(
		driverUUID.String(),
		ctx,
	)

	if err != nil {
		return err
	}

	h.send(
		passengerClientKey,
		WsResponse{
			Event:   "ride_accepted",
			Success: true,
			Data: map[string]any{
				"ride_id":             ride.ID.String(),
				"driver_id":           driverUUID.String(),
				"driver_location_lat": driverLat,
				"driver_location_lng": driverLng,
			},
		},
	)

	h.setRidePassenger(
		ride.ID.String(),
		passengerClientKey,
	)

	return nil
}

func (h *EventHandler) DispatchRideCanceled(ctx context.Context, rideID string, passengerID string) error {
	passengerClientKey := utils.BuildClientKey(
		passengerID,
		"",
	)

	h.send(
		passengerClientKey,
		WsResponse{
			Event:   "ride_canceled",
			Success: true,
			Data: map[string]any{
				"ride_id": rideID,
				"reason":  "Tempo limite de espera atingido",
			},
		},
	)

	return nil
}

func (h *EventHandler) DispatchRideStatusChanged(ctx context.Context, rideID string, recipientID string, status string) error {
	clientKey := utils.BuildClientKey(
		recipientID,
		"",
	)

	h.send(
		clientKey,
		WsResponse{
			Event:   "ride_status_changed",
			Success: true,
			Data: map[string]any{
				"ride_id": rideID,
				"status":  status,
			},
		},
	)

	return nil
}
