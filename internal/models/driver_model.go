package models

import "github.com/Necobgs/move-fast-backend/internal/db/sqlc"

type DriverResponse struct {
	sqlc.Driver
	Vehicle sqlc.Vehicle `json:"vehicle"`
}
