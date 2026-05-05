package dto

import "mime/multipart"

type CreateUserDTO struct {
	Name     string                `form:"name" binding:"required"`
	Email    string                `form:"email" binding:"required,email"`
	Password string                `form:"password" binding:"required"`
	Photo    *multipart.FileHeader `form:"photo"`
}
