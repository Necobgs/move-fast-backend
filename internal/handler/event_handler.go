package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/Necobgs/move-fast-backend/internal/consts"
	"github.com/Necobgs/move-fast-backend/internal/db/sqlc"
	"github.com/Necobgs/move-fast-backend/internal/realtime"
	"github.com/Necobgs/move-fast-backend/internal/utils"
	"github.com/Necobgs/move-fast-backend/internal/ws/message"
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
	return &EventHandler{
		repository: repository,
		rdb:        rdb,
		hub:        hub,

		ridePassengers: make(map[string]string),
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
		log.Printf(
			"[ERROR] falha ao serializar response event=%s err=%v",
			response.Event,
			err,
		)

		return false
	}

	return h.hub.SendToClient(
		clientID,
		payload,
	)
}

func (h *EventHandler) fail(
	c *realtime.Client,
	event string,
	message string,
	err error,
) {

	log.Printf(
		"[ERROR] event=%s user_id=%s message=%s err=%v",
		event,
		c.Claims.Id,
		message,
		err,
	)

	h.send(
		c.Claims.Id,
		WsResponse{
			Event:   event,
			Success: false,
			Message: message,
		},
	)
}

func (h *EventHandler) success(
	clientID string,
	event string,
	data any,
) {

	log.Printf(
		"[INFO] event=%s client_id=%s success=true",
		event,
		clientID,
	)

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

		h.fail(
			c,
			"update_location_driver",
			"payload inválido",
			err,
		)

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

		h.fail(
			c,
			"update_location_driver",
			"falha ao atualizar localização",
			err,
		)

		return
	}

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
		log.Printf(
			"[INFO] driver sem corrida ativa driver_id=%s",
			c.Claims.DriverId,
		)

		return
	}

	passengerClientID, ok := h.getRidePassenger(
		rideID.String(),
	)

	if !ok {

		log.Printf(
			"[WARN] passenger não encontrado ride_id=%s",
			rideID.String(),
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

		log.Printf(
			"[WARN] falha ao enviar localização para passenger ride_id=%s",
			rideID.String(),
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

func (h *EventHandler) RequestRide(
	c *realtime.Client,
	data message.BaseMessage,
) {

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	var requestRide message.RequestRideMessage

	err := json.Unmarshal(data.Data, &requestRide)
	if err != nil {

		h.fail(
			c,
			"requested_ride",
			"payload inválido",
			err,
		)

		return
	}

	passengerID := uuid.MustParse(
		c.Claims.Id,
	)

	log.Printf(
		"[INFO] passenger solicitando corrida passenger_id=%s",
		passengerID.String(),
	)

	ride, err := h.repository.CreateRide(
		ctx,
		sqlc.CreateRideParams{
			ID:          uuid.New(),
			PassengerID: passengerID,

			OriginLat:     requestRide.OriginLocationLat,
			OriginLng:     requestRide.OriginLocationLng,
			OriginAddress: requestRide.OriginLocationAddress,

			DestinationLat:     requestRide.DestinationLocationLat,
			DestinationLng:     requestRide.DestinationLocationLng,
			DestinationAddress: requestRide.DestinationLocationAddress,
		},
	)

	if err != nil {

		h.fail(
			c,
			"requested_ride",
			"falha ao criar corrida",
			err,
		)

		return
	}

	nearestDriverID, err := h.getNearestDriver(
		requestRide.OriginLocationLat,
		requestRide.OriginLocationLng,
		10,
		ctx,
	)

	if err != nil {

		log.Printf(
			"[WARN] nenhum motorista encontrado ride_id=%s",
			ride.ID.String(),
		)

		_, _ = h.repository.UpdateRide(
			ctx,
			sqlc.UpdateRideParams{
				ID:       ride.ID,
				StatusID: consts.StatusCanceledRideId,
			},
		)

		h.fail(
			c,
			"requested_ride",
			"nenhum motorista encontrado",
			err,
		)

		return
	}

	log.Printf(
		"[INFO] motorista encontrado ride_id=%s driver_id=%s",
		ride.ID.String(),
		*nearestDriverID,
	)

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

			"ride_passenger_id": passengerID.String(),
		},
	)
}

func (h *EventHandler) RequestedRide(
	c *realtime.Client,
	data message.BaseMessage,
) {

	ctx := context.Background()

	var requestedRide message.RequestedRideMessage

	err := json.Unmarshal(
		data.Data,
		&requestedRide,
	)

	if err != nil {

		h.fail(
			c,
			"ride_accepted",
			"payload inválido",
			err,
		)

		return
	}

	rideID, err := uuid.Parse(
		requestedRide.RideId,
	)

	if err != nil {

		h.fail(
			c,
			"ride_accepted",
			"ride_id inválido",
			err,
		)

		return
	}

	if !requestedRide.Accepted {

		log.Printf(
			"[INFO] corrida recusada ride_id=%s driver_id=%s",
			rideID.String(),
			c.Claims.DriverId,
		)

		return
	}

	driverID, err := uuid.Parse(
		c.Claims.DriverId,
	)

	if err != nil {

		h.fail(
			c,
			"ride_accepted",
			"driver_id inválido",
			err,
		)

		return
	}

	rideBasicInfo, err := h.repository.UpdateRide(
		ctx,
		sqlc.UpdateRideParams{
			ID:       rideID,
			StatusID: consts.StatusWaitingDriverId,

			DriverID: pgtype.UUID{
				Bytes: driverID,
				Valid: true,
			},
		},
	)

	if err != nil {

		h.fail(
			c,
			"ride_accepted",
			"falha ao atualizar corrida",
			err,
		)

		return
	}

	passengerClientKey := utils.BuildClientKey(
		rideBasicInfo.PassengerID.String(),
		"",
	)

	driverLng, driverLat, err := h.getDriverLocation(
		driverID.String(),
		ctx,
	)

	if err != nil {

		h.fail(
			c,
			"ride_accepted",
			"falha ao buscar localização do motorista",
			err,
		)

		return
	}

	sent := h.send(
		passengerClientKey,
		WsResponse{
			Event:   "ride_accepted",
			Success: true,
			Data: map[string]any{
				"ride_id": rideBasicInfo.ID.String(),

				"driver_id": driverID.String(),

				"driver_location_lat": driverLat,
				"driver_location_lng": driverLng,
			},
		},
	)

	if !sent {

		log.Printf(
			"[WARN] passenger offline ride_id=%s",
			rideBasicInfo.ID.String(),
		)

		return
	}

	h.setRidePassenger(
		rideBasicInfo.ID.String(),
		passengerClientKey,
	)

	log.Printf(
		"[INFO] corrida aceita ride_id=%s driver_id=%s",
		rideBasicInfo.ID.String(),
		driverID.String(),
	)
}

func (h *EventHandler) SyncPassengerConnection(
	client *realtime.Client,
) {

	ctx := context.Background()

	passengerID, err := uuid.Parse(
		client.Claims.Id,
	)

	if err != nil {

		log.Printf(
			"[ERROR] passenger_id inválido err=%v",
			err,
		)

		return
	}

	ride, err := h.repository.GetActiveRideFromPassenger(
		ctx,
		passengerID,
	)

	if err != nil {

		log.Printf(
			"[INFO] passenger sem corrida ativa passenger_id=%s",
			passengerID.String(),
		)

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

	log.Printf(
		"[INFO] passenger reconectado ride_id=%s passenger_id=%s",
		ride.ID.String(),
		passengerID.String(),
	)
}
