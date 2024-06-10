package mocks

import (
	"github.com/Montheankul-K/jod-jod/domains/entities"
	"github.com/stretchr/testify/mock"
)

type TransactionRepositoryMock struct {
	mock.Mock
}

func (m *TransactionRepositoryMock) GetAllTxn(filter entities.GetAllTxnFilter, pagination entities.Pagination) ([]entities.GetAllResponse, error) {
	args := m.Called(filter, pagination)
	return args.Get(0).([]entities.GetAllResponse), args.Error(1)
}

func (m *TransactionRepositoryMock) GetAllBySpenderId(spenderId uint) ([]entities.GetAllResponse, error) {
	args := m.Called(spenderId)
	return args.Get(0).([]entities.GetAllResponse), args.Error(1)
}

func (m *TransactionRepositoryMock) GetByTxnType(req entities.GetByTxnTypeRequest) ([]entities.GetAllByTxnTypeResponse, error) {
	args := m.Called(req)
	return args.Get(0).([]entities.GetAllByTxnTypeResponse), args.Error(1)
}

func (m *TransactionRepositoryMock) GetByCategory(req entities.GetByCategoryRequest) ([]entities.GetByCategoryResponse, error) {
	args := m.Called(req)
	return args.Get(0).([]entities.GetByCategoryResponse), args.Error(1)
}

func (m *TransactionRepositoryMock) GetByPeriod(req entities.GetByTxnTypeRequest, filter entities.PeriodFilter) ([]entities.GetAllByTxnTypeResponse, error) {
	args := m.Called(req, filter)
	return args.Get(0).([]entities.GetAllByTxnTypeResponse), args.Error(1)
}

func (m *TransactionRepositoryMock) SaveTxn(req entities.Transaction) (uint, error) {
	args := m.Called(req)
	return args.Get(0).(uint), args.Error(1)
}

func (m *TransactionRepositoryMock) UpdateTxn(txnId uint, req entities.Transaction) error {
	args := m.Called(txnId, req)
	return args.Error(0)
}

func (m *TransactionRepositoryMock) DeleteTxn(spenderId uint, txnId uint) error {
	args := m.Called(spenderId, txnId)
	return args.Error(0)
}
