-- name: CreateRide :one
insert into rides(
    id,
    destination_address, 
    destination_lat, 
    destination_lng,
 
    origin_address, 
    origin_lat, 
    origin_lng, 

    passenger_id
    )
values ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id,destination_address,destination_lat,destination_lng,origin_address,origin_lat,origin_lng,passenger_id,status_id;

-- name: UpdateRide :one
UPDATE rides
SET 
    status_id = COALESCE($1, status_id),
    driver_id = COALESCE($2, driver_id)
WHERE 
    id = $3
RETURNING id, passenger_id, driver_id, status_id;

-- name: GetRideFromDriver :one
select
    id
from 
    rides
where 
    driver_id = $1 and
    status_id = $2;