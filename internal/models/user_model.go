package models

import "github.com/Necobgs/move-fast-backend/internal/db/sqlc"

type UserResponse struct {
	sqlc.User
	Driver DriverResponse `json:"driver"`
}
