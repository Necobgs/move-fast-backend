-- name: CreateRide :one
insert into
    rides (
        id,
        destination_address,
        destination_lat,
        destination_lng,
        origin_address,
        origin_lat,
        origin_lng,
        passenger_id
    )
values (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8
    )
RETURNING
    id,
    destination_address,
    destination_lat,
    destination_lng,
    origin_address,
    origin_lat,
    origin_lng,
    passenger_id,
    status_id;

-- name: UpdateRide :one
UPDATE rides
SET
    status_id = COALESCE($1, status_id),
    driver_id = COALESCE($2, driver_id)
WHERE
    id = $3
    and status_id != 'CANCELLED_RIDE'
RETURNING
    id,
    passenger_id,
    driver_id,
    status_id;

-- name: GetRideFromDriver :one
select id from rides where driver_id = $1 and status_id = $2;

-- name: GetRideFromPassenger :one
select id
from rides
where
    passenger_id = $1
    and status_id = $2;

-- name: GetActiveRideFromPassenger :one
SELECT *
FROM rides
WHERE
    passenger_id = $1
    AND status_id IN (
        'WAITING_DRIVER',
        'STARTED_RIDE'
    )
LIMIT 1;

-- name: GetRideById :one
SELECT * FROM rides WHERE id = $1;

-- name: ExistsActiveRidesForPassenger :one
SELECT EXISTS (
        SELECT 1
        FROM rides
        WHERE
            passenger_id = $1
            AND status_id NOT IN (
                'ENDED_RIDE', 'CANCELLED_RIDE'
            )
    );

-- name: ExistsActiveRidesForDriver :one
SELECT EXISTS (
        SELECT 1
        FROM rides
        WHERE
            driver_id = $1
            AND status_id NOT IN (
                'ENDED_RIDE', 'CANCELLED_RIDE'
            )
    );

-- name: GetRideHistory :many
SELECT
    r.id,
    r.destination_address,
    r.destination_lat,
    r.destination_lng,
    r.origin_address,
    r.origin_lat,
    r.origin_lng,
    r.created_at,
    r.started_at,
    r.ended_at,
    up.name AS passenger_name,
    ud.name AS driver_name
FROM
    rides r
    LEFT JOIN users up ON r.passenger_id = up.id
    LEFT JOIN drivers dr ON r.driver_id = dr.id
    LEFT JOIN users ud ON dr.user_id = ud.id
    LEFT JOIN vehicles vd ON dr.vehicle_id = vd.id
WHERE (
        r.passenger_id = $1
        OR r.driver_id = $2
    )
    AND r.created_at < $3
ORDER BY r.created_at DESC
LIMIT $4;

-- name: StartRideQuery :one
UPDATE rides
SET
    status_id = $1,
    started_at = NOW()
WHERE
    id = $2
    and status_id != 'CANCELLED_RIDE'
RETURNING
    id,
    passenger_id,
    driver_id,
    status_id;

-- name: FinishRideQuery :one
UPDATE rides
SET
    status_id = $1,
    ended_at = NOW()
WHERE
    id = $2
    and status_id != 'CANCELLED_RIDE'
RETURNING
    id,
    passenger_id,
    driver_id,
    status_id;

-- name: GetRideInProgessFromDriver :one
SELECT *
FROM rides
WHERE
    driver_id = $1
    AND status_id IN (
        'STARTED_RIDE',
        'WAITING_DRIVER'
    )
LIMIT 1;

-- name: GetRideInProgressFromPassenger :one
SELECT *
FROM rides
WHERE
    passenger_id = $1
    AND status_id IN (
        'STARTED_RIDE',
        'WAITING_DRIVER'
    )
LIMIT 1;