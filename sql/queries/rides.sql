-- name: CreateRide :one
insert into rides(
    destination_address, 
    destination_lat, 
    destination_lng,
 
    origin_address, 
    origin_lat, 
    origin_lng, 

    passenger_id
    )
values ($1, $2, $3, $4, $5, $6, $7)
RETURNING id,destination_address,destination_lat,destination_lng,origin_address,origin_lat,origin_lng,passenger_id,status_id;

-- name: UpdateRide :one
UPDATE rides
SET status_id = '730e3d5c-5abf-44e1-86d9-697014eee856', driver_id = $2
WHERE id = $1
RETURNING id, passenger_id, driver_id;

-- name: IsRideFromUser :one
select EXISTS(
    select 1 from rides
    where id = $1 and passenger_id = $2 and driver_id = $3
);