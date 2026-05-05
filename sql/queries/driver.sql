-- name: CreateDriver :one
insert into drivers(id, user_id,cnh_number, vehicle_id)
VALUES($1, $2, $3, $4)
RETURNING id, user_id,cnh_number, vehicle_id;

-- name: ExistsDriver :one
select EXISTS(
    SELECT 1 from drivers
    where user_id = $1 or cnh_number = $2
);