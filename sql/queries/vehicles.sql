-- name: CreateVehicle :one
insert into vehicles(id, license_plate, model, brand, color)
VALUES($1, $2, $3, $4, $5)
RETURNING id, license_plate, model, brand, color;

-- name: ExistsVehicle :one
select EXISTS(
    select 1 from vehicles
    where license_plate ilike $1
);