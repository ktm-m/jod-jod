package mocks

import (
	"github.com/Montheankul-K/jod-jod/domains/entities"
	"github.com/stretchr/testify/mock"
)

type UserRepositoryMock struct {
	mock.Mock
}

func (m *UserRepositoryMock) GetUsers(pagination entities.Pagination) ([]entities.GetUserResponse, error) {
	args := m.Called(pagination)
	return args.Get(0).([]entities.GetUserResponse), args.Error(1)
}

func (m *UserRepositoryMock) GetUser(userId uint) (*entities.GetUserResponse, error) {
	args := m.Called(userId)
	return args.Get(0).(*entities.GetUserResponse), args.Error(1)
}

func (m *UserRepositoryMock) GetUserForLogin(username string) (*entities.GetUserForLoginResponse, error) {
	args := m.Called(username)
	return args.Get(0).(*entities.GetUserForLoginResponse), args.Error(1)
}

func (m *UserRepositoryMock) CreateUser(req entities.Users) (uint, error) {
	args := m.Called(req)
	return args.Get(0).(uint), args.Error(1)
}

func (m *UserRepositoryMock) UpdateUser(userId uint, req entities.Users) error {
	args := m.Called(userId, req)
	return args.Error(0)
}

func (m *UserRepositoryMock) UpdatePassword(userId uint, newPassword string) error {
	args := m.Called(userId, newPassword)
	return args.Error(0)
}

func (m *UserRepositoryMock) DeleteUser(userId uint) error {
	args := m.Called(userId)
	return args.Error(0)
}
