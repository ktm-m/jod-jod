package user

import (
	"errors"
	"github.com/Montheankul-K/jod-jod/domains/entities"
	"github.com/Montheankul-K/jod-jod/repository/mocks"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
	"testing"
)

func TestUserService_GetUsers_Success(t *testing.T) {
	mockRepo := new(mocks.UserRepositoryMock)
	var logger echo.Logger
	repoPagination := entities.Pagination{
		PageItem: 10,
		Page:     1,
	}

	mockRepo.On("GetUsers", repoPagination).Return([]entities.GetUserResponse{
		{ID: 1, Firstname: "John", Lastname: "Doe", Email: "john.d@gmail.com"},
	}, nil)
	service := NewUserService(mockRepo, logger)
	pagination := Pagination{
		PageItem: 10,
		Page:     1,
	}
	result, err := service.GetUsers(pagination)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, uint(1), result[0].ID)
	assert.Equal(t, "John", result[0].Firstname)
	assert.Equal(t, "Doe", result[0].Lastname)
	assert.Equal(t, "john.d@gmail.com", result[0].Email)
}

func TestUserService_GetUsers_RecordNotFound(t *testing.T) {
	mockRepo := new(mocks.UserRepositoryMock)
	var logger echo.Logger
	repoPagination := entities.Pagination{
		PageItem: 10,
		Page:     1,
	}

	mockRepo.On("GetUsers", repoPagination).Return([]entities.GetUserResponse{}, gorm.ErrRecordNotFound)
	service := NewUserService(mockRepo, logger)
	pagination := Pagination{
		PageItem: 10,
		Page:     1,
	}
	_, err := service.GetUsers(pagination)

	assert.EqualError(t, err, gorm.ErrRecordNotFound.Error())
}

func TestUserService_GetUsers_OtherError(t *testing.T) {
	mockRepo := new(mocks.UserRepositoryMock)
	var logger echo.Logger
	repoPagination := entities.Pagination{
		PageItem: 10,
		Page:     1,
	}

	mockRepo.On("GetUsers", repoPagination).Return([]entities.GetUserResponse{}, errors.New("some error"))
	service := NewUserService(mockRepo, logger)
	pagination := Pagination{
		PageItem: 10,
		Page:     1,
	}
	_, err := service.GetUsers(pagination)

	assert.EqualError(t, err, "failed to get users")
}

func TestUserService_GetUser_Success(t *testing.T) {
	mockRepo := new(mocks.UserRepositoryMock)
	var logger echo.Logger

	mockRepo.On("GetUser", uint(1)).Return(&entities.GetUserResponse{
		ID: 1, Firstname: "John", Lastname: "Doe", Email: "john.d@gmail.com",
	}, nil)
	service := NewUserService(mockRepo, logger)
	result, err := service.GetUser(uint(1))

	assert.Nil(t, err)
	assert.Equal(t, uint(1), result.ID)
	assert.Equal(t, "John", result.Firstname)
	assert.Equal(t, "Doe", result.Lastname)
	assert.Equal(t, "john.d@gmail.com", result.Email)
}

func TestUserService_GetUser_RecordNotFound(t *testing.T) {
	mockRepo := new(mocks.UserRepositoryMock)
	var logger echo.Logger

	mockRepo.On("GetUser", uint(1)).Return(&entities.GetUserResponse{}, gorm.ErrRecordNotFound)
	service := NewUserService(mockRepo, logger)
	_, err := service.GetUser(uint(1))

	assert.EqualError(t, err, gorm.ErrRecordNotFound.Error())
}

func TestUserService_GetUser_OtherError(t *testing.T) {
	mockRepo := new(mocks.UserRepositoryMock)
	var logger echo.Logger

	mockRepo.On("GetUser", uint(1)).Return(&entities.GetUserResponse{}, errors.New("some error"))
	service := NewUserService(mockRepo, logger)
	_, err := service.GetUser(uint(1))

	assert.EqualError(t, err, "failed to get user")
}

func TestUserService_CreateUser_Success(t *testing.T) {
	mockRepo := new(mocks.UserRepositoryMock)
	var logger echo.Logger

	mockRepo.On("CreateUser", mock.Anything).Return(uint(1), nil)
	service := NewUserService(mockRepo, logger)

	req := Users{
		Firstname: "John",
		Lastname:  "Doe",
		Email:     "john.d@gmail.com",
		Username:  "john.d",
		Password:  "password",
	}
	result, err := service.CreateUser(req)

	assert.Nil(t, err)
	assert.Equal(t, uint(1), result)
}

func TestUserService_CreateUser_RecordNotFound(t *testing.T) {
	mockRepo := new(mocks.UserRepositoryMock)
	var logger echo.Logger

	mockRepo.On("CreateUser", mock.Anything).Return(uint(0), gorm.ErrRecordNotFound)
	service := NewUserService(mockRepo, logger)

	req := Users{
		Firstname: "John",
		Lastname:  "Doe",
		Email:     "john.d@gmail.com",
		Username:  "john.d",
		Password:  "password",
	}
	_, err := service.CreateUser(req)

	assert.EqualError(t, err, gorm.ErrRecordNotFound.Error())
}

func TestUserService_CreateUser_OtherError(t *testing.T) {
	mockRepo := new(mocks.UserRepositoryMock)
	var logger echo.Logger

	mockRepo.On("CreateUser", mock.Anything).Return(uint(0), errors.New("some error"))
	service := NewUserService(mockRepo, logger)

	req := Users{
		Firstname: "John",
		Lastname:  "Doe",
		Email:     "john.d@gmail.com",
		Username:  "john.d",
		Password:  "password",
	}
	_, err := service.CreateUser(req)

	assert.EqualError(t, err, "failed to create user")
}

func TestUserService_UpdateInfo_Success(t *testing.T) {
	mockRepo := new(mocks.UserRepositoryMock)
	var logger echo.Logger
	userId := uint(1)
	req := Users{
		Firstname: "John",
		Lastname:  "Doe",
		Email:     "john.d@gmail.com",
	}

	mockRepo.On("UpdateUser", userId, mock.Anything).Return(nil)
	service := NewUserService(mockRepo, logger)
	err := service.UpdateInfo(userId, req)

	assert.Nil(t, err)
}

func TestUserService_UpdateInfo_Error(t *testing.T) {
	mockRepo := new(mocks.UserRepositoryMock)
	var logger echo.Logger
	userId := uint(1)
	req := Users{
		Firstname: "John",
		Lastname:  "Doe",
		Email:     "john.d@gmail.com",
	}

	mockRepo.On("UpdateUser", userId, mock.Anything).Return(errors.New("some error"))
	service := NewUserService(mockRepo, logger)
	err := service.UpdateInfo(userId, req)

	assert.EqualError(t, err, "failed to update user")
}

func TestUserService_UpdatePassword_Success(t *testing.T) {
	mockRepo := new(mocks.UserRepositoryMock)
	var logger echo.Logger
	userId := uint(1)
	newPassword := "newPassword"

	mockRepo.On("UpdatePassword", userId, mock.AnythingOfType("string")).Return(nil)
	service := NewUserService(mockRepo, logger)

	req := UpdatePasswordRequest{
		ID:       userId,
		Password: newPassword,
	}
	err := service.UpdatePassword(req)

	assert.Nil(t, err)
}

func TestUserService_UpdatePassword_Error(t *testing.T) {
	mockRepo := new(mocks.UserRepositoryMock)
	var logger echo.Logger
	userId := uint(1)
	newPassword := "newPassword"

	mockRepo.On("UpdatePassword", userId, mock.AnythingOfType("string")).Return(errors.New("some error"))
	service := NewUserService(mockRepo, logger)

	req := UpdatePasswordRequest{
		ID:       userId,
		Password: newPassword,
	}
	err := service.UpdatePassword(req)

	assert.EqualError(t, err, "failed to update user")
}

func TestUserService_DeleteUser_Success(t *testing.T) {
	mockRepo := new(mocks.UserRepositoryMock)
	var logger echo.Logger
	userId := uint(1)

	mockRepo.On("DeleteUser", userId).Return(nil)
	service := NewUserService(mockRepo, logger)
	err := service.DeleteUser(userId)

	assert.Nil(t, err)
}

func TestUserService_DeleteUser_Error(t *testing.T) {
	mockRepo := new(mocks.UserRepositoryMock)
	var logger echo.Logger
	userId := uint(1)

	mockRepo.On("DeleteUser", userId).Return(errors.New("some error"))
	service := NewUserService(mockRepo, logger)
	err := service.DeleteUser(userId)

	assert.EqualError(t, err, "failed to delete user")
}
