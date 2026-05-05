package handler

import (
	"context"
	"encoding/json"

	"github.com/Necobgs/move-fast-backend/internal/db/sqlc"
	"github.com/Necobgs/move-fast-backend/internal/ws"
	"github.com/Necobgs/move-fast-backend/internal/ws/message"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
)

type EventHandler struct {
	repository *sqlc.Queries
	rdb        *redis.Client
	ClientsWs  map[string]*ws.ClientWebSocket
}

func NewEventHandler(repository *sqlc.Queries, rdb *redis.Client, clientsWs map[string]*ws.ClientWebSocket) *EventHandler {
	return &EventHandler{
		repository: repository,
		rdb:        rdb,
		ClientsWs:  clientsWs,
	}
}

func (h *EventHandler) UpdateLocationDriver(c *ws.ClientWebSocket, data message.BaseMessage) {

	var updateLocation message.UpdateLocMessage
	err := json.Unmarshal(data.Data, &updateLocation)
	if err != nil {
		return
	}
	ctx := context.Background()
	clientKey := ws.BuildClientKey(c.Claims.Id, c.Claims.DriverId)
	h.rdb.GeoAdd(ctx, "driver_locations", &redis.GeoLocation{
		Name:      clientKey,
		Longitude: updateLocation.Lng,
		Latitude:  updateLocation.Lat,
	})
}

func (h *EventHandler) getNearestDriver(lat float64, lng float64, radius float64, ctx *context.Context) (*string, error) {
	nearestDrivers, err := h.rdb.GeoSearch(*ctx, "driver_locations", &redis.GeoSearchQuery{
		Latitude:   lat,
		Longitude:  lng,
		Radius:     radius,
		RadiusUnit: "km",

		Sort:  "ASC",
		Count: 1,
	}).Result()

	if err != nil || len(nearestDrivers) == 0 {
		return nil, err
	}

	return &nearestDrivers[0], nil
}

func (h *EventHandler) getDriverLocation(driverId string, ctx *context.Context) (*float64, *float64, error) {
	driverClientKey := ws.BuildClientKey("", driverId)
	driverLocation, err := h.rdb.GeoPos(*ctx, "driver_locations", driverClientKey).Result()
	if err != nil || len(driverLocation) == 0 {
		return nil, nil, err
	}
	return &driverLocation[0].Longitude, &driverLocation[0].Latitude, nil
}

func (h *EventHandler) GetMyDriverLocation(c *ws.ClientWebSocket, data message.BaseMessage) {
	ctx := context.Background()

	var driverLocation message.DriverLocationMessage
	err := json.Unmarshal(data.Data, &driverLocation)
	if err != nil {
		return
	}

	userId, err := uuid.Parse(c.Claims.Id)
	if err != nil {
		return
	}

	rideId, err := uuid.Parse(driverLocation.RideId)
	if err != nil {
		return
	}

	driverId, err := uuid.Parse(driverLocation.DriverId)
	if err != nil {
		return
	}

	isRideFromUser, err := h.repository.IsRideFromUser(ctx, sqlc.IsRideFromUserParams{
		ID:          rideId,
		PassengerID: userId,
		DriverID: pgtype.UUID{
			Bytes: driverId,
			Valid: true,
		},
	})

	if !isRideFromUser {
		return
	}

	lng, lat, err := h.getDriverLocation(driverLocation.DriverId, &ctx)
	if err != nil {
		return
	}

	passengerClientkey := ws.BuildClientKey(c.Claims.Id, c.Claims.DriverId)

	value, ok := h.ClientsWs[passengerClientkey]
	if !ok {
		return
	}
	value.Conn.WriteJSON(message.DriverLocation{
		Lat: *lat,
		Lng: *lng,
	})

	driverLocJson, err := json.Marshal(message.DriverLocation{
		Lat: *lat,
		Lng: *lng,
	})

	if err != nil {
		return
	}

	value.Send <- driverLocJson
}

func (h *EventHandler) RequestRide(c *ws.ClientWebSocket, data message.BaseMessage) {
	ctx := context.Background()

	var requestRide message.RequestRideMessage
	err := json.Unmarshal(data.Data, &requestRide)
	if err != nil {
		return
	}

	passengerId, err := uuid.Parse(c.Claims.Id)
	if err != nil {
		return
	}

	ride, err := h.repository.CreateRide(ctx, sqlc.CreateRideParams{
		PassengerID: passengerId,

		OriginLat:     requestRide.OriginLocationLat,
		OriginLng:     requestRide.OriginLocationLng,
		OriginAddress: requestRide.OriginLocationAddress,

		DestinationLat:     requestRide.DestinationLocationLat,
		DestinationLng:     requestRide.DestinationLocationLng,
		DestinationAddress: requestRide.DestinationLocationAddress,
	})
	if err != nil {
		return
	}

	nearestDriver, err := h.getNearestDriver(requestRide.OriginLocationLat, requestRide.OriginLocationLng, 10, &ctx)
	if err != nil {
		return
	}

	driverClient, ok := h.ClientsWs[*nearestDriver]
	if !ok {
		return
	}

	rideRequest := message.RideRequest(*ride)

	rideJson, err := json.Marshal(rideRequest)
	if err != nil {
		return
	}

	baseMessage := message.BaseMessage{
		Event: "ride_request",
		Data:  rideJson,
	}

	jsonMessage, err := json.Marshal(baseMessage)
	if err != nil {
		return
	}

	driverClient.Send <- jsonMessage
}

func (h *EventHandler) RideRequest(c *ws.ClientWebSocket, data message.BaseMessage) {
	ctx := context.Background()

	var rideRequest message.RideRequestMessage
	err := json.Unmarshal(data.Data, &rideRequest)
	if err != nil {
		return
	}

	rideId, err := uuid.Parse(rideRequest.RideId)
	if err != nil {
		return
	}

	if !rideRequest.Accepted {
		// TODO: Adicionar ação de blacklist para não procura solicitar novamente a esse motorista a mesma corrida
		return
	}

	// Segurança contra não motoristas
	driverId, err := uuid.Parse(c.Claims.DriverId)
	if err != nil {
		return
	}

	rideBasicInfo, err := h.repository.UpdateRide(ctx, sqlc.UpdateRideParams{
		ID: rideId,
		DriverID: pgtype.UUID{
			Bytes: driverId,
			Valid: true,
		},
	})
	if err != nil {
		return
	}

	passengerClientKey := ws.BuildClientKey(rideBasicInfo.PassengerID.String(), "")
	value, ok := h.ClientsWs[passengerClientKey]
	if !ok {
		return
	}

	driverLng, driverLat, err := h.getDriverLocation(driverId.String(), &ctx)
	if err != nil {

	}

	rideAccepted := message.RideAcceptedMessage{
		RideId:            rideBasicInfo.ID.String(),
		DriverId:          driverId.String(),
		DriverLocationLat: *driverLat,
		DriverLocationLng: *driverLng,
	}

	rideAcceptedJson, err := json.Marshal(rideAccepted)
	if err != nil {
		return
	}
	value.Send <- rideAcceptedJson
}
