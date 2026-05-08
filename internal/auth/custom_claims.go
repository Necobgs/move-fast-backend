package auth

import (
	"github.com/golang-jwt/jwt"
)

type CustomClaims struct {
	Id         string `json:"id"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	DriverId   string `json:"driver_id"`
	Identifier string `json:"identifier"`
	jwt.StandardClaims
}
