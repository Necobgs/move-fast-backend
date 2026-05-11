package service

import (
	"context"
	"time"

	"github.com/Necobgs/move-fast-backend/configs"
	"github.com/Necobgs/move-fast-backend/internal/auth"
	"github.com/Necobgs/move-fast-backend/internal/db/sqlc"
	"github.com/Necobgs/move-fast-backend/internal/dto"
	"github.com/Necobgs/move-fast-backend/internal/response"
	"github.com/Necobgs/move-fast-backend/internal/utils"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repository *sqlc.Queries
}

func NewAuthService(repository *sqlc.Queries) *AuthService {
	return &AuthService{repository: repository}
}

func (s *AuthService) Signin(signinDto dto.SigninDto) (*dto.SigninResponseDto, *response.ErrorResponse) {
	userFounded, err := s.repository.FindUser(context.Background(), sqlc.FindUserParams{
		Email: signinDto.Email,
	})

	if err != nil {
		return nil, &response.ErrUnauthorized
	}

	err = bcrypt.CompareHashAndPassword([]byte(userFounded.Password), []byte(signinDto.Password))
	if err != nil {
		return nil, &response.ErrUnauthorized
	}

	tokenString, err := s.GenerateToken(userFounded)
	if err != nil {
		return nil, &response.ErrInternalServer
	}

	signInResponseDto := dto.SigninResponseDto{Token: tokenString, PhotoUrl: userFounded.PhotoUrl}
	return &signInResponseDto, nil
}

func (s *AuthService) GenerateToken(user *sqlc.FindUserRow) (string, error) {
	claims := s.GenerateClaims(user)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, *claims)
	tokenString, err := token.SignedString([]byte(configs.Cfg.JWTSecretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (s *AuthService) GenerateClaims(user *sqlc.FindUserRow) *jwt.MapClaims {
	return &jwt.MapClaims{
		"id":         user.ID,
		"email":      user.Email,
		"name":       user.Name,
		"driver_id":  user.DriverID,
		"identifier": utils.BuildClientKey(user.ID.String(), user.DriverID.String()),
		"exp":        time.Now().Add(time.Hour).Unix(),
	}
}

func (s *AuthService) ValidateToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(tokenString, &auth.CustomClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); ok {
			return []byte(configs.Cfg.JWTSecretKey), nil
		}
		return nil, jwt.ErrSignatureInvalid
	})
	return token, err
}
