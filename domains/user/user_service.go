package user

import (
	"errors"
	"fmt"
	"github.com/Montheankul-K/jod-jod/config"
	"github.com/Montheankul-K/jod-jod/domains/entities"
	"github.com/Montheankul-K/jod-jod/repository/user_repository"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"strconv"
	"time"
)

type IUserService interface {
	GetUsers(pagination Pagination) ([]GetUserResponse, error)
	GetUser(userId uint) (*GetUserResponse, error)
	CreateUser(req Users) (uint, error)
	Login(req LoginRequest) (*LoginResponse, error)
	RegenToken(req RegenTokenRequest) (string, error)
	UpdateInfo(userId uint, req Users) error
	UpdatePassword(req UpdatePasswordRequest) error
	DeleteUser(userId uint) error
}

type userService struct {
	userRepository user_repository.IUserRepository
	logger         echo.Logger
}

func NewUserService(userRepository user_repository.IUserRepository, logger echo.Logger) IUserService {
	return &userService{
		userRepository: userRepository,
		logger:         logger,
	}
}

func (s *userService) GetUsers(pagination Pagination) ([]GetUserResponse, error) {
	newPagination := entities.Pagination{
		PageItem: pagination.PageItem,
		Page:     pagination.Page,
	}
	results, err := s.userRepository.GetUsers(newPagination)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, errors.New("failed to get users")
	}
	var newResult []GetUserResponse
	for _, value := range results {
		result := GetUserResponse{
			ID:        value.ID,
			Firstname: value.Firstname,
			Lastname:  value.Lastname,
			Email:     value.Email,
		}
		newResult = append(newResult, result)
	}
	return newResult, nil
}

func (s *userService) GetUser(userId uint) (*GetUserResponse, error) {
	result, err := s.userRepository.GetUser(userId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, errors.New("failed to get user")
	}
	newResult := GetUserResponse{
		ID:        result.ID,
		Firstname: result.Firstname,
		Lastname:  result.Lastname,
		Email:     result.Email,
	}
	return &newResult, nil
}

func (s *userService) CreateUser(req Users) (uint, error) {
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error(err)
		return 0, errors.New("failed to hash password")
	}

	user := entities.Users{
		Firstname: req.Firstname,
		Lastname:  req.Lastname,
		Email:     req.Email,
		Username:  req.Username,
		Password:  string(hashPassword),
	}
	result, err := s.userRepository.CreateUser(user)
	if err != nil {
		return 0, errors.New("failed to create user")
	}
	return result, nil
}

func (s *userService) Login(req LoginRequest) (*LoginResponse, error) {
	user, err := s.userRepository.GetUserForLogin(req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, errors.New("failed to get user")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		s.logger.Error(err)
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, err
		}
		return nil, errors.New("failed to compare password")
	}

	result, err := generateToken(user.ID)
	if err != nil {
		s.logger.Error(err)
		return nil, errors.New("failed to generate token")
	}
	s.logger.Infof("username: %s loggin success", req.Username)
	return result, nil
}

func generateToken(userId uint) (*LoginResponse, error) {
	cfg := config.GetConfig()
	if cfg == nil {
		return nil, errors.New("failed to get config")
	}

	issuer := fmt.Sprintf("%s v.%s", cfg.Server.Name, cfg.Server.Version)
	accessClaims := Claims{
		StandardClaims: jwt.StandardClaims{
			Issuer:    issuer,
			Subject:   "access token",
			ExpiresAt: time.Now().Add(time.Minute * 10).Unix(),
		},
		UserID: strconv.Itoa(int(userId)),
	}

	refreshClaims := Claims{
		StandardClaims: jwt.StandardClaims{
			Issuer:    issuer,
			Subject:   "refresh token",
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
		},
		UserID: strconv.Itoa(int(userId)),
	}

	access := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err := access.SignedString([]byte(cfg.Auth.Secret))
	if err != nil {
		return nil, errors.New("failed to generate access token")
	}

	refresh := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err := refresh.SignedString([]byte(cfg.Auth.Secret))
	if err != nil {
		return nil, errors.New("failed to generate refresh token")
	}

	res := &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	return res, nil
}

func (s *userService) RegenToken(req RegenTokenRequest) (string, error) {
	claims, err := validateToken(req.RefreshToken)
	if err != nil {
		s.logger.Error(err)
		return "", err
	}

	userId, err := strconv.ParseUint(claims.UserID, 10, 64)
	if err != nil {
		s.logger.Error(err)
		return "", errors.New("failed to parse user id")
	}

	token, err := generateToken(uint(userId))
	if err != nil {
		s.logger.Error(err)
		return "", errors.New("failed to generate token")
	}
	result := token.AccessToken
	return result, nil
}

func validateToken(tokenString string) (*Claims, error) {
	cfg := config.GetConfig()
	secret := cfg.Auth.Secret
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		var validationErr *jwt.ValidationError
		if errors.As(err, &validationErr) {
			switch {
			case validationErr.Errors&jwt.ValidationErrorMalformed != 0:
				return nil, errors.New("invalid token format")
			case validationErr.Errors&jwt.ValidationErrorExpired != 0:
				return nil, errors.New("token is expired")
			case validationErr.Errors&jwt.ValidationErrorNotValidYet != 0:
				return nil, errors.New("token is not valid")
			case validationErr.Errors&jwt.ValidationErrorSignatureInvalid != 0:
				return nil, errors.New("signature is invalid")
			default:
				return nil, errors.New("token is invalid")
			}
		}
		return nil, errors.New("internal server error")
	}

	if !token.Valid {
		return nil, errors.New("token is invalid")
	}

	if !claims.VerifyExpiresAt(time.Now().Unix(), true) {
		return nil, errors.New("token is expired")
	}
	return claims, nil
}

func (s *userService) UpdateInfo(userId uint, req Users) error {
	user := entities.Users{
		Firstname: req.Firstname,
		Lastname:  req.Lastname,
		Email:     req.Email,
	}
	err := s.userRepository.UpdateUser(userId, user)
	if err != nil {
		return errors.New("failed to update user")
	}
	return nil
}

func (s *userService) UpdatePassword(req UpdatePasswordRequest) error {
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error(err)
		return errors.New("failed to hash password")
	}

	err = s.userRepository.UpdatePassword(req.ID, string(hashPassword))
	if err != nil {
		return errors.New("failed to update user")
	}
	return nil
}

func (s *userService) DeleteUser(userId uint) error {
	err := s.userRepository.DeleteUser(userId)
	if err != nil {
		return errors.New("failed to delete user")
	}
	return nil
}
