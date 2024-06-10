package transaction

import (
	"errors"
	"github.com/Montheankul-K/jod-jod/domains/entities"
	"github.com/Montheankul-K/jod-jod/repository/mocks"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
	"testing"
	"time"
)

func TestTransactionService_SaveByManual_Success(t *testing.T) {
	mockRepo := new(mocks.TransactionRepositoryMock)
	logger := echo.New().Logger

	mockRepo.On("SaveTxn", mock.Anything).Return(uint(1), nil)
	service := NewTransactionService(mockRepo, logger)

	req := Transaction{
		Date:      time.Now(),
		Amount:    1000,
		Category:  "food",
		Note:      "note for expense",
		ImageUrl:  "https://image.jpg",
		SpenderId: 1,
	}
	result, err := service.SaveByManual(req)

	assert.NoError(t, err)
	assert.Equal(t, uint(1), result)
}

func TestTransactionService_SaveByManual_Error(t *testing.T) {
	mockRepo := new(mocks.TransactionRepositoryMock)
	logger := echo.New().Logger

	mockRepo.On("SaveTxn", mock.Anything).Return(uint(0), errors.New("some error"))
	service := NewTransactionService(mockRepo, logger)

	req := Transaction{
		Date:      time.Now(),
		Amount:    1000,
		Category:  "food",
		Note:      "note for expense",
		ImageUrl:  "https://image.jpg",
		SpenderId: 1,
	}
	_, err := service.SaveByManual(req)

	assert.EqualError(t, err, "failed to save transaction")
}

func TestTransactionService_GetDetails_Success(t *testing.T) {
	mockRepo := new(mocks.TransactionRepositoryMock)
	logger := echo.New().Logger

	date := time.Now()
	mockRepo.On("GetByTxnType", mock.Anything).Return([]entities.GetAllByTxnTypeResponse{
		{ID: uint(1), Date: &date, Amount: 1000, Category: "food", ImageUrl: ""},
	}, nil)
	service := NewTransactionService(mockRepo, logger)

	req := GetByTxnTypeRequest{
		SpenderId: uint(1),
		TxnType:   "food",
	}
	result, err := service.GetDetails(req)

	assert.NoError(t, err)
	assert.Equal(t, uint(1), result[0].ID)
	assert.Equal(t, float64(1000), result[0].Amount)
	assert.Equal(t, "food", result[0].Category)
}

func TestTransactionService_GetDetails_RecordNotFound(t *testing.T) {
	mockRepo := new(mocks.TransactionRepositoryMock)
	logger := echo.New().Logger

	mockRepo.On("GetByTxnType", mock.Anything).Return([]entities.GetAllByTxnTypeResponse{},
		gorm.ErrRecordNotFound)
	service := NewTransactionService(mockRepo, logger)

	req := GetByTxnTypeRequest{
		SpenderId: uint(1),
		TxnType:   "food",
	}
	_, err := service.GetDetails(req)

	assert.EqualError(t, err, gorm.ErrRecordNotFound.Error())
}

func TestTransactionService_GetDetails_OtherError(t *testing.T) {
	mockRepo := new(mocks.TransactionRepositoryMock)
	logger := echo.New().Logger

	mockRepo.On("GetByTxnType", mock.Anything).Return([]entities.GetAllByTxnTypeResponse{},
		errors.New("some error"))
	service := NewTransactionService(mockRepo, logger)

	req := GetByTxnTypeRequest{
		SpenderId: uint(1),
		TxnType:   "food",
	}
	_, err := service.GetDetails(req)

	assert.NotNil(t, err, "failed to get transaction")
}

func TestTransactionService_GetSummary_Success(t *testing.T) {
	mockRepo := new(mocks.TransactionRepositoryMock)
	logger := echo.New().Logger

	date1 := time.Now().AddDate(0, 0, -2)
	date2 := time.Now()
	mockRepo.On("GetByTxnType", mock.Anything).Return([]entities.GetAllByTxnTypeResponse{
		{ID: uint(1), Date: &date1, Amount: 1000, Category: "food", ImageUrl: ""},
		{ID: uint(2), Date: &date2, Amount: 2000, Category: "food", ImageUrl: ""},
	}, nil)
	service := NewTransactionService(mockRepo, logger)

	req := GetByTxnTypeRequest{
		SpenderId: uint(1),
		TxnType:   "food",
	}
	result, err := service.GetSummary(req)

	assert.NoError(t, err)
	assert.Equal(t, float64(3000), result.TotalAmount)
	assert.Equal(t, float64(1500), result.AveragePerDay)
	assert.Equal(t, 2, result.TotalTxn)
}

func TestTransactionService_GetSummary_RecordNotFound(t *testing.T) {
	mockRepo := new(mocks.TransactionRepositoryMock)
	logger := echo.New().Logger

	mockRepo.On("GetByTxnType", mock.Anything).Return([]entities.GetAllByTxnTypeResponse{}, gorm.ErrRecordNotFound)
	service := NewTransactionService(mockRepo, logger)

	req := GetByTxnTypeRequest{
		SpenderId: uint(1),
		TxnType:   "food",
	}
	_, err := service.GetSummary(req)

	assert.EqualError(t, err, gorm.ErrRecordNotFound.Error())
}

func TestTransactionService_GetSummary_OtherError(t *testing.T) {
	mockRepo := new(mocks.TransactionRepositoryMock)
	logger := echo.New().Logger

	mockRepo.On("GetByTxnType", mock.Anything).Return([]entities.GetAllByTxnTypeResponse{}, errors.New("some error"))
	service := NewTransactionService(mockRepo, logger)

	req := GetByTxnTypeRequest{
		SpenderId: uint(1),
		TxnType:   "food",
	}
	_, err := service.GetSummary(req)
	assert.NotNil(t, err)
}

func TestTransactionService_GetBalance_Success(t *testing.T) {
	mockRepo := new(mocks.TransactionRepositoryMock)
	logger := echo.New().Logger

	spenderId := uint(1)
	date1 := time.Now().AddDate(0, 0, -2)
	date2 := time.Now()
	mockRepo.On("GetAllBySpenderId", mock.Anything).Return([]entities.GetAllResponse{
		{ID: uint(1), Date: &date1, Amount: 1000, Category: "food", ImageUrl: "", TransactionType: "expense"},
		{ID: uint(2), Date: &date2, Amount: 2000, Category: "food", ImageUrl: "", TransactionType: "expense"},
	}, nil)
	service := NewTransactionService(mockRepo, logger)

	result, err := service.GetBalance(spenderId)

	assert.NoError(t, err)
	assert.Equal(t, float64(0), result.TotalAmountEarned)
	assert.Equal(t, float64(3000), result.TotalAmountSpent)
	assert.Equal(t, float64(-3000), result.TotalAmountSaved)
}

func TestTransactionService_GetBalance_RecordNotFound(t *testing.T) {
	mockRepo := new(mocks.TransactionRepositoryMock)
	logger := echo.New().Logger

	spenderId := uint(1)
	mockRepo.On("GetAllBySpenderId", mock.Anything).Return([]entities.GetAllResponse{}, gorm.ErrRecordNotFound)
	service := NewTransactionService(mockRepo, logger)

	_, err := service.GetBalance(spenderId)

	assert.EqualError(t, err, gorm.ErrRecordNotFound.Error())
}

func TestTransactionService_GetBalance_OtherError(t *testing.T) {
	mockRepo := new(mocks.TransactionRepositoryMock)
	logger := echo.New().Logger

	spenderId := uint(1)
	mockRepo.On("GetAllBySpenderId", mock.Anything).Return([]entities.GetAllResponse{}, errors.New("some error"))
	service := NewTransactionService(mockRepo, logger)

	_, err := service.GetBalance(spenderId)

	assert.NotNil(t, err)
}

func TestTransactionService_GetByCategory_Success(t *testing.T) {
	mockRepo := new(mocks.TransactionRepositoryMock)
	logger := echo.New().Logger

	date1 := time.Now().AddDate(0, 0, -2)
	date2 := time.Now()
	mockRepo.On("GetByCategory", mock.Anything).Return([]entities.GetByCategoryResponse{
		{ID: uint(1), Date: date1, Amount: 1000, ImageUrl: ""},
		{ID: uint(2), Date: date2, Amount: 2000, ImageUrl: ""},
	}, nil)
	service := NewTransactionService(mockRepo, logger)

	req := GetByCategoryRequest{
		SpenderId: uint(1),
		Category:  "food",
		TxnType:   "expense",
	}
	result, err := service.GetByCategory(req)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, uint(1), result[0].ID)
	assert.Equal(t, date1, result[0].Date)
	assert.Equal(t, float64(1000), result[0].Amount)
	assert.Equal(t, "", result[0].ImageUrl)
}

func TestTransactionService_GetByCategory_RecordNotFound(t *testing.T) {
	mockRepo := new(mocks.TransactionRepositoryMock)
	logger := echo.New().Logger

	mockRepo.On("GetByCategory", mock.Anything).Return([]entities.GetByCategoryResponse{}, gorm.ErrRecordNotFound)
	service := NewTransactionService(mockRepo, logger)

	req := GetByCategoryRequest{
		SpenderId: uint(1),
		Category:  "food",
		TxnType:   "expense",
	}
	_, err := service.GetByCategory(req)

	assert.EqualError(t, err, gorm.ErrRecordNotFound.Error())
}

func TestTransactionService_GetByCategory_OtherError(t *testing.T) {
	mockRepo := new(mocks.TransactionRepositoryMock)
	logger := echo.New().Logger

	mockRepo.On("GetByCategory", mock.Anything).Return([]entities.GetByCategoryResponse{}, errors.New("some error"))
	service := NewTransactionService(mockRepo, logger)

	req := GetByCategoryRequest{
		SpenderId: uint(1),
		Category:  "food",
		TxnType:   "expense",
	}
	_, err := service.GetByCategory(req)

	assert.NotNil(t, err)
}

func TestTransactionService_GetByPeriod_Success(t *testing.T) {
	mockRepo := new(mocks.TransactionRepositoryMock)
	logger := echo.New().Logger

	date1 := time.Now().AddDate(0, 0, -2)
	date2 := time.Now()
	mockRepo.On("GetByPeriod", mock.Anything, mock.Anything).Return([]entities.GetAllByTxnTypeResponse{
		{ID: uint(1), Date: &date1, Amount: 1000, Category: "food", ImageUrl: ""},
		{ID: uint(2), Date: &date2, Amount: 2000, Category: "food", ImageUrl: ""},
	}, nil)
	service := NewTransactionService(mockRepo, logger)

	req := GetByTxnTypeRequest{
		SpenderId: uint(1),
		TxnType:   "expense",
	}
	filter := PeriodFilter{
		StartDate: &date1,
		EndDate:   &date2,
	}
	result, err := service.GetByPeriod(req, filter)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, uint(1), result[0].ID)
	assert.Equal(t, &date1, result[0].Date)
	assert.Equal(t, float64(1000), result[0].Amount)
	assert.Equal(t, "", result[0].ImageUrl)
}

func TestTransactionService_GetByPeriod_RecordNotFound(t *testing.T) {
	mockRepo := new(mocks.TransactionRepositoryMock)
	logger := echo.New().Logger

	date1 := time.Now().AddDate(0, 0, -2)
	date2 := time.Now()
	mockRepo.On("GetByPeriod", mock.Anything, mock.Anything).Return([]entities.GetAllByTxnTypeResponse{},
		gorm.ErrRecordNotFound)
	service := NewTransactionService(mockRepo, logger)

	req := GetByTxnTypeRequest{
		SpenderId: uint(1),
		TxnType:   "expense",
	}
	filter := PeriodFilter{
		StartDate: &date1,
		EndDate:   &date2,
	}
	_, err := service.GetByPeriod(req, filter)

	assert.EqualError(t, err, gorm.ErrRecordNotFound.Error())
}

func TestTransactionService_GetByPeriod_OtherError(t *testing.T) {
	mockRepo := new(mocks.TransactionRepositoryMock)
	logger := echo.New().Logger

	date1 := time.Now().AddDate(0, 0, -2)
	date2 := time.Now()
	mockRepo.On("GetByPeriod", mock.Anything, mock.Anything).Return([]entities.GetAllByTxnTypeResponse{}, errors.New("some error"))
	service := NewTransactionService(mockRepo, logger)

	req := GetByTxnTypeRequest{
		SpenderId: uint(1),
		TxnType:   "expense",
	}
	filter := PeriodFilter{
		StartDate: &date1,
		EndDate:   &date2,
	}
	_, err := service.GetByPeriod(req, filter)

	assert.EqualError(t, err, "failed to get transaction")
}

func TestTransactionService_Update_Success(t *testing.T) {
	mockRepo := new(mocks.TransactionRepositoryMock)
	logger := echo.New().Logger

	txnId := uint(1)
	mockRepo.On("UpdateTxn", mock.Anything, mock.Anything).Return(nil)
	service := NewTransactionService(mockRepo, logger)

	req := Transaction{
		Date:      time.Now(),
		Amount:    1000,
		Category:  "food",
		Note:      "note for expense",
		ImageUrl:  "https://image.jpg",
		SpenderId: 1,
	}
	err := service.Update(txnId, req)

	assert.Nil(t, err)
}

func TestTransactionService_Update_Error(t *testing.T) {
	mockRepo := new(mocks.TransactionRepositoryMock)
	logger := echo.New().Logger

	txnId := uint(1)
	mockRepo.On("UpdateTxn", mock.Anything, mock.Anything).Return(errors.New("some error"))
	service := NewTransactionService(mockRepo, logger)

	req := Transaction{
		Date:      time.Now(),
		Amount:    1000,
		Category:  "food",
		Note:      "note for expense",
		ImageUrl:  "https://image.jpg",
		SpenderId: 1,
	}
	err := service.Update(txnId, req)

	assert.EqualError(t, err, "failed to update transaction")
}

func TestTransactionService_Delete_Success(t *testing.T) {
	mockRepo := new(mocks.TransactionRepositoryMock)
	logger := echo.New().Logger
	spenderId := uint(1)
	txnId := uint(1)

	mockRepo.On("DeleteTxn", mock.Anything, mock.Anything).Return(nil)
	service := NewTransactionService(mockRepo, logger)

	err := service.Delete(spenderId, txnId)

	assert.Nil(t, err)
}

func TestTransactionService_Delete_Error(t *testing.T) {
	mockRepo := new(mocks.TransactionRepositoryMock)
	logger := echo.New().Logger
	spenderId := uint(1)
	txnId := uint(1)

	mockRepo.On("DeleteTxn", mock.Anything, mock.Anything).Return(errors.New("some error"))
	service := NewTransactionService(mockRepo, logger)

	err := service.Delete(spenderId, txnId)

	assert.EqualError(t, err, "failed to delete transaction")
}

func TestTransactionService_GetAllTxn_Success(t *testing.T) {
	mockRepo := new(mocks.TransactionRepositoryMock)
	logger := echo.New().Logger

	date1 := time.Now().AddDate(0, 0, -2)
	mockRepo.On("GetAllTxn", mock.Anything, mock.Anything).Return([]entities.GetAllResponse{
		{ID: uint(1), Date: &date1, Amount: 1000, Category: "food", ImageUrl: "", TransactionType: "expense"},
		{ID: uint(2), Date: &date1, Amount: 2000, Category: "food", ImageUrl: "", TransactionType: "expense"},
	}, nil)
	service := NewTransactionService(mockRepo, logger)

	filter := GetAllTxnFilter{
		Date:     &date1,
		Category: "food",
		TxnType:  "expense",
	}
	pagination := Pagination{
		PageItem: 10,
		Page:     1,
	}
	result, err := service.GetAllTxn(filter, pagination)

	assert.Nil(t, err)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, uint(1), result[0].ID)
	assert.Equal(t, &date1, result[0].Date)
	assert.Equal(t, float64(1000), result[0].Amount)
	assert.Equal(t, "food", result[0].Category)
	assert.Equal(t, "", result[0].ImageUrl)
	assert.Equal(t, "expense", result[0].TransactionType)
}

func TestTransactionService_GetAllTxn_RecordNotFound(t *testing.T) {
	mockRepo := new(mocks.TransactionRepositoryMock)
	logger := echo.New().Logger

	date1 := time.Now().AddDate(0, 0, -2)
	mockRepo.On("GetAllTxn", mock.Anything, mock.Anything).Return([]entities.GetAllResponse{},
		gorm.ErrRecordNotFound)
	service := NewTransactionService(mockRepo, logger)

	filter := GetAllTxnFilter{
		Date:     &date1,
		Category: "food",
		TxnType:  "expense",
	}
	pagination := Pagination{
		PageItem: 10,
		Page:     1,
	}
	_, err := service.GetAllTxn(filter, pagination)

	assert.EqualError(t, err, gorm.ErrRecordNotFound.Error())
}

func TestTransactionService_GetAllTxn_OtherError(t *testing.T) {
	mockRepo := new(mocks.TransactionRepositoryMock)
	logger := echo.New().Logger

	date1 := time.Now().AddDate(0, 0, -2)
	mockRepo.On("GetAllTxn", mock.Anything, mock.Anything).Return([]entities.GetAllResponse{},
		errors.New("some error"))
	service := NewTransactionService(mockRepo, logger)

	filter := GetAllTxnFilter{
		Date:     &date1,
		Category: "food",
		TxnType:  "expense",
	}
	pagination := Pagination{
		PageItem: 10,
		Page:     1,
	}
	_, err := service.GetAllTxn(filter, pagination)

	assert.EqualError(t, err, "failed to get transaction")
}
