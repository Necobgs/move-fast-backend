package service

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/Necobgs/move-fast-backend/configs"
	"github.com/Necobgs/move-fast-backend/internal/db/sqlc"
	"github.com/Necobgs/move-fast-backend/internal/dto"
	"github.com/Necobgs/move-fast-backend/internal/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repository *sqlc.Queries
}

func NewUserService(repository *sqlc.Queries) *UserService {
	return &UserService{repository: repository}
}

func (s *UserService) GetMe(email string) (*sqlc.FindSafeUserByEmailRow, *response.ErrorResponse) {
	ctx := context.Background()
	user, err := s.repository.FindSafeUserByEmail(ctx, email)
	if err != nil {
		return nil, &response.ErrNotFound
	}
	return user, nil
}

func (s *UserService) CreateUser(createUserDto *dto.CreateUserDTO, c *gin.Context) (*sqlc.CreateUserRow, *response.ErrorResponse) {
	ctx := context.Background()

	exists, err := s.repository.ExistsUserByEmail(ctx, createUserDto.Email)
	if err != nil {
		return nil, &response.ErrInternalServer
	}

	if exists {
		return nil, &response.ErrUserAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(createUserDto.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, &response.ErrInternalServer
	}

	var photoUrl string
	if createUserDto.Photo != nil {
		filename := uuid.New().String() + filepath.Ext(createUserDto.Photo.Filename)

		path := fmt.Sprintf("%s/%s", configs.Cfg.PathUploads, filename)
		photoUrl = fmt.Sprintf("%s/%s", configs.Cfg.UrlUploads, filename)

		c.SaveUploadedFile(createUserDto.Photo, path)
	} else {
		photoUrl = fmt.Sprintf("%s/%s", configs.Cfg.UrlUploads, "default_photo_profile.webp")
	}

	createdUser, err := s.repository.CreateUser(ctx, sqlc.CreateUserParams{
		ID:       uuid.New(),
		Name:     createUserDto.Name,
		Email:    createUserDto.Email,
		PhotoUrl: photoUrl,
		Password: string(hashedPassword),
	})
	if err != nil {
		return nil, &response.ErrInternalServer
	}
	return createdUser, nil
}
