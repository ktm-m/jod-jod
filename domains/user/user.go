package user

import (
	"github.com/golang-jwt/jwt"
	"gorm.io/gorm"
)

type Users struct {
	gorm.Model
	Firstname string `gorm:"type:varchar(50); not null; column:firstname" validate:"required" json:"firstname"`
	Lastname  string `gorm:"type:varchar(50); not null; column:lastname" validate:"required" json:"lastname"`
	Email     string `gorm:"type:varchar(50); not null; unique; column:email" validate:"required; email" json:"email"`
	Username  string `gorm:"type:varchar(20); not null; unique; column:username" validate:"required" json:"username"`
	Password  string `gorm:"type:varchar(20); not null; column:password" validate:"required" json:"password"`
}

type Claims struct {
	jwt.StandardClaims
	UserID string
}

type Pagination struct {
	PageItem int `query:"page-item"`
	Page     int `query:"page"`
}

type UpdatePasswordRequest struct {
	ID       uint   `json:"user_id" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type GetUserResponse struct {
	ID        uint   `json:"user_id" gorm:"column:id"`
	Firstname string `json:"firstname" gorm:"column:firstname"`
	Lastname  string `json:"lastname" gorm:"column:lastname"`
	Email     string `json:"email" gorm:"column:email"`
}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type RegenTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type GetUserForLoginResponse struct {
	ID       uint   `gorm:"column:id"`
	Username string `gorm:"column:username"`
	Password string `gorm:"column:password"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
