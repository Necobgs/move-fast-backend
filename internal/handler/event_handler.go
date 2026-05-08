package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Necobgs/move-fast-backend/internal/consts"
	"github.com/Necobgs/move-fast-backend/internal/db/sqlc"
	"github.com/Necobgs/move-fast-backend/internal/utils"
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

	RidePassengers map[string]*ws.ClientWebSocket
}

func NewEventHandler(repository *sqlc.Queries, rdb *redis.Client, clientsWs map[string]*ws.ClientWebSocket) *EventHandler {
	return &EventHandler{
		repository:     repository,
		rdb:            rdb,
		ClientsWs:      clientsWs,
		RidePassengers: make(map[string]*ws.ClientWebSocket),
	}
}

func (h *EventHandler) UpdateLocationDriver(c *ws.ClientWebSocket, data message.BaseMessage) {

	var updateLocation message.UpdateLocMessage
	err := json.Unmarshal(data.Data, &updateLocation)
	if err != nil {
		return
	}
	ctx := context.Background()
	clientKey := utils.BuildClientKey(c.Claims.Id, c.Claims.DriverId)
	_, err = h.rdb.GeoAdd(ctx, "driver_locations", &redis.GeoLocation{
		Name:      clientKey,
		Longitude: updateLocation.Lng,
		Latitude:  updateLocation.Lat,
	}).Result()
	if err != nil {
		return
	}

	rideId, err := h.repository.GetRideFromDriver(ctx, sqlc.GetRideFromDriverParams{
		DriverID: pgtype.UUID{
			Bytes: uuid.MustParse(c.Claims.DriverId),
			Valid: true,
		},
		StatusID: uuid.MustParse(consts.StatusWaitingDriverId),
	})
	if err != nil {
		return
	}

	response, err := json.Marshal(map[string]any{
		"event": "driver_location",
		"data": map[string]any{
			"driver_id":           c.Claims.DriverId,
			"ride_id":             rideId.String(),
			"driver_location_lat": updateLocation.Lat,
			"driver_location_lng": updateLocation.Lng,
		},
	})
	if err != nil {
		return
	}

	h.RidePassengers[rideId.String()].Send <- response

}

func (h *EventHandler) getNearestDriver(lat float64, lng float64, radius float64, ctx context.Context) (*string, error) {
	nearestDrivers, err := h.rdb.GeoSearch(ctx, "driver_locations", &redis.GeoSearchQuery{
		Latitude:   lat,
		Longitude:  lng,
		Radius:     radius,
		RadiusUnit: "km",

		Sort:  "ASC",
		Count: 1,
	}).Result()

	if err != nil {
		return nil, err
	}

	if len(nearestDrivers) == 0 {
		return nil, errors.New("no driver found")
	}

	return &nearestDrivers[0], nil
}

func (h *EventHandler) getDriverLocation(driverId string, ctx context.Context) (float64, float64, error) {
	driverClientKey := utils.BuildClientKey("", driverId)
	driverLocation, err := h.rdb.GeoPos(ctx, "driver_locations", driverClientKey).Result()
	if err != nil || len(driverLocation) == 0 {
		return 0, 0, err
	}
	return driverLocation[0].Longitude, driverLocation[0].Latitude, nil
}

func (h *EventHandler) RequestRide(c *ws.ClientWebSocket, data message.BaseMessage) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fmt.Println("request_ride acionado!")
	var requestRide message.RequestRideMessage
	err := json.Unmarshal(data.Data, &requestRide)
	if err != nil {
		fmt.Println("Erro ao desearilar data")
		fmt.Println(err.Error())
		return
	}
	fmt.Println("Data deserealizado:")
	fmt.Println(requestRide)

	passengerId, err := uuid.Parse(c.Claims.Id)
	fmt.Println("PassengerId encontrado: ", passengerId)
	if err != nil {
		fmt.Println("PassengerId não encontrado no claims")
		return
	}

	ride, err := h.repository.CreateRide(ctx, sqlc.CreateRideParams{
		ID:          uuid.New(),
		PassengerID: passengerId,

		OriginLat:     requestRide.OriginLocationLat,
		OriginLng:     requestRide.OriginLocationLng,
		OriginAddress: requestRide.OriginLocationAddress,

		DestinationLat:     requestRide.DestinationLocationLat,
		DestinationLng:     requestRide.DestinationLocationLng,
		DestinationAddress: requestRide.DestinationLocationAddress,
	})

	passengerClientKey := utils.BuildClientKey(c.Claims.Id, c.Claims.DriverId)
	passengerChan, ok := h.ClientsWs[passengerClientKey]
	if !ok {
		fmt.Println("Passenger channel não encontrado")
		fmt.Println("PassengerChannels: ")
		fmt.Println(h.ClientsWs)
		return
	}

	if err != nil {
		fmt.Println("Erro ao criar corrida")
		json := []byte(`{"event":"request_ride","success": false,"message": "Erro ao criar solicitão de corrida"}`)
		passengerChan.Send <- json
		return
	}

	fmt.Println("Ride criado: ", *ride)

	nearestDriverId, err := h.getNearestDriver(requestRide.OriginLocationLat, requestRide.OriginLocationLng, 10, ctx)
	if err != nil {
		fmt.Println("Erro ao buscar motorista mais próximo")
		_, err := h.repository.UpdateRide(ctx, sqlc.UpdateRideParams{
			ID:       ride.ID,
			StatusID: uuid.MustParse(consts.StatusCanceledRideId),
		})
		if err != nil {
			fmt.Println("Erro ao atualizar corrida")
			fmt.Println(err.Error())
		}
		json := []byte(`{"event":"request_ride","success": false,"message": "Nenhum motorista disponível no momento"}`)
		passengerChan.Send <- json
		return
	}

	fmt.Println("Motorista mais próximo: ", *nearestDriverId)

	driverClient, ok := h.ClientsWs[*nearestDriverId]
	if !ok {
		return
	}

	response, err := json.Marshal(map[string]any{
		"event":   "requested_ride",
		"success": true,
		"data": map[string]any{
			"ride_id":             ride.ID.String(),
			"ride_origin_lat":     ride.OriginLat,
			"ride_origin_lng":     ride.OriginLng,
			"ride_origin_address": ride.OriginAddress,

			"ride_destination_lat":     ride.DestinationLat,
			"ride_destination_lng":     ride.DestinationLng,
			"ride_destination_address": ride.DestinationAddress,

			"ride_passenger_id": passengerId.String(),
		},
	})
	if err != nil {
		return
	}
	driverClient.Send <- response
}

func (h *EventHandler) RequestedRide(c *ws.ClientWebSocket, data message.BaseMessage) {
	ctx := context.Background()

	var requestedRide message.RequestedRideMessage
	err := json.Unmarshal(data.Data, &requestedRide)
	if err != nil {
		return
	}

	rideId, err := uuid.Parse(requestedRide.RideId)
	if err != nil {
		return
	}

	if !requestedRide.Accepted {
		// TODO: Adicionar ação de blacklist para não procura solicitar novamente a esse motorista a mesma corrida
		return
	}

	driverId, err := uuid.Parse(c.Claims.DriverId)
	if err != nil {
		return
	}

	waitingDriverId := uuid.MustParse(consts.StatusWaitingDriverId)
	rideBasicInfo, err := h.repository.UpdateRide(ctx, sqlc.UpdateRideParams{
		ID:       rideId,
		StatusID: waitingDriverId,
		DriverID: pgtype.UUID{
			Bytes: driverId,
			Valid: true,
		},
	})
	if err != nil {
		return
	}

	passengerClientKey := utils.BuildClientKey(rideBasicInfo.PassengerID.String(), "")
	value, ok := h.ClientsWs[passengerClientKey]
	if !ok {
		return
	}

	driverLng, driverLat, err := h.getDriverLocation(driverId.String(), ctx)
	if err != nil {
		return
	}

	rideAccepted := message.RideAcceptedMessage{
		RideId:            rideBasicInfo.ID.String(),
		DriverId:          driverId.String(),
		DriverLocationLat: driverLat,
		DriverLocationLng: driverLng,
	}

	rideAcceptedJson, err := json.Marshal(rideAccepted)
	if err != nil {
		return
	}
	value.Send <- rideAcceptedJson

	h.RidePassengers[rideBasicInfo.ID.String()] = value
}
