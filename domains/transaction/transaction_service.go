package transaction

import (
	"errors"
	"github.com/Montheankul-K/jod-jod/domains/entities"
	"github.com/Montheankul-K/jod-jod/repository/transaction_repository"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"strings"
	"time"
)

type ITransactionService interface {
	SaveByManual(req Transaction) (uint, error)
	GetDetails(req GetByTxnTypeRequest) ([]GetAllByTxnTypeResponse, error)
	GetSummary(req GetByTxnTypeRequest) (*GetSummaryResponse, error)
	GetBalance(spenderId uint) (*GetBalanceResponse, error)
	GetByCategory(req GetByCategoryRequest) ([]GetByCategoryResponse, error)
	GetByPeriod(req GetByTxnTypeRequest, filter PeriodFilter) ([]GetAllByTxnTypeResponse, error)
	Update(txnId uint, req Transaction) error
	Delete(spenderId, txnId uint) error
	GetAllTxn(filter GetAllTxnFilter, pagination Pagination) ([]GetAllResponse, error)
}

type transactionService struct {
	transactionRepository transaction_repository.ITransactionRepository
	logger                echo.Logger
}

func NewTransactionService(transactionRepository transaction_repository.ITransactionRepository, logger echo.Logger) ITransactionService {
	return &transactionService{
		transactionRepository: transactionRepository,
		logger:                logger,
	}
}

func (s *transactionService) SaveByManual(req Transaction) (uint, error) {
	txn := entities.Transaction{
		Date:            req.Date,
		Amount:          req.Amount,
		Category:        req.Category,
		TransactionType: req.TransactionType,
		Note:            req.Note,
		ImageUrl:        req.ImageUrl,
		SpenderId:       req.SpenderId,
	}
	result, err := s.transactionRepository.SaveTxn(txn)
	if err != nil {
		return 0, errors.New("failed to save transaction")
	}
	s.logger.Infof("saved transaction with ID: %d success", result)
	return result, nil
}

func (s *transactionService) GetDetails(req GetByTxnTypeRequest) ([]GetAllByTxnTypeResponse, error) {
	txn := entities.GetByTxnTypeRequest{
		SpenderId: req.SpenderId,
		TxnType:   req.TxnType,
	}
	results, err := s.transactionRepository.GetByTxnType(txn)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, errors.New("failed to get transaction")
	}
	s.logger.Infof("get transaction details of spender id: %d success", req.SpenderId)

	var newResults []GetAllByTxnTypeResponse
	for _, value := range results {
		result := &GetAllByTxnTypeResponse{
			ID:       value.ID,
			Date:     value.Date,
			Amount:   value.Amount,
			Category: value.Category,
			ImageUrl: value.ImageUrl,
		}
		newResults = append(newResults, *result)
	}
	return newResults, nil
}

func (s *transactionService) GetSummary(req GetByTxnTypeRequest) (*GetSummaryResponse, error) {
	txn := entities.GetByTxnTypeRequest{
		SpenderId: req.SpenderId,
		TxnType:   req.TxnType,
	}
	results, err := s.transactionRepository.GetByTxnType(txn)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, errors.New("failed to get transaction")
	}

	var newResults []GetAllByTxnTypeResponse
	for _, value := range results {
		result := &GetAllByTxnTypeResponse{
			ID:       value.ID,
			Date:     value.Date,
			Amount:   value.Amount,
			Category: value.Category,
			ImageUrl: value.ImageUrl,
		}
		newResults = append(newResults, *result)
	}

	result, err := calculateSummary(newResults)
	if err != nil {
		s.logger.Error(err)
		return nil, err
	}
	s.logger.Infof("get transaction detail summary of spender id: %d success", req.SpenderId)
	return result, nil
}

func calculateSummary(allTxn []GetAllByTxnTypeResponse) (*GetSummaryResponse, error) {
	var totalAmount float64
	var totalTxn int
	var minDate, maxDate *time.Time
	for _, txn := range allTxn {
		totalAmount += txn.Amount
		totalTxn += 1

		if minDate == nil || txn.Date.Before(*minDate) {
			minDate = txn.Date
		}

		if maxDate == nil || txn.Date.After(*maxDate) {
			maxDate = txn.Date
		}
	}

	avgAmountPerDay, err := calculateAvgAmountPerDay(totalAmount, minDate, maxDate)
	if err != nil {
		return nil, err
	}

	result := &GetSummaryResponse{
		TotalAmount:   totalAmount,
		AveragePerDay: avgAmountPerDay,
		TotalTxn:      totalTxn,
	}
	return result, nil
}

func calculateAvgAmountPerDay(totalAmount float64, minDate, maxDate *time.Time) (float64, error) {
	if minDate == nil || maxDate == nil {
		return 0, errors.New("minDate or maxDate is empty")
	}

	totalDays := int(maxDate.Sub(*minDate).Hours() / 24)
	if totalDays == 0 {
		return totalAmount, nil
	}

	result := totalAmount / float64(totalDays)
	return result, nil
}

func (s *transactionService) GetBalance(spenderId uint) (*GetBalanceResponse, error) {
	results, err := s.transactionRepository.GetAllBySpenderId(spenderId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, errors.New("failed to get transaction")
	}

	var newResults []GetAllResponse
	for _, value := range results {
		result := &GetAllResponse{
			ID:              value.ID,
			Date:            value.Date,
			Amount:          value.Amount,
			Category:        value.Category,
			ImageUrl:        value.ImageUrl,
			TransactionType: value.TransactionType,
		}
		newResults = append(newResults, *result)
	}

	result := calculateBalance(newResults)
	s.logger.Infof("get transaction balance of spender id: %d success", spenderId)
	return result, nil
}

func calculateBalance(allTxn []GetAllResponse) *GetBalanceResponse {
	var totalAmountEarned float64
	var totalAmountSpent float64
	for _, txn := range allTxn {
		if strings.ToLower(txn.TransactionType) == "income" {
			totalAmountEarned += txn.Amount
		} else if strings.ToLower(txn.TransactionType) == "expense" {
			totalAmountSpent += txn.Amount
		}
	}
	totalAmountSaved := totalAmountEarned - totalAmountSpent
	result := &GetBalanceResponse{
		TotalAmountEarned: totalAmountEarned,
		TotalAmountSpent:  totalAmountSpent,
		TotalAmountSaved:  totalAmountSaved,
	}
	return result
}

func (s *transactionService) GetByCategory(req GetByCategoryRequest) ([]GetByCategoryResponse, error) {
	txn := entities.GetByCategoryRequest{
		SpenderId: req.SpenderId,
		Category:  req.Category,
		TxnType:   req.TxnType,
	}
	results, err := s.transactionRepository.GetByCategory(txn)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, errors.New("failed to get transaction")
	}
	s.logger.Infof("get transaction by category of spender id: %d success", req.SpenderId)
	var newResults []GetByCategoryResponse
	for _, value := range results {
		result := &GetByCategoryResponse{
			ID:       value.ID,
			Date:     value.Date,
			Amount:   value.Amount,
			ImageUrl: value.ImageUrl,
		}
		newResults = append(newResults, *result)
	}
	return newResults, nil
}

func (s *transactionService) GetByPeriod(req GetByTxnTypeRequest, filter PeriodFilter) ([]GetAllByTxnTypeResponse, error) {
	txn := entities.GetByTxnTypeRequest{
		SpenderId: req.SpenderId,
		TxnType:   req.TxnType,
	}
	newFilter := entities.PeriodFilter{
		StartDate: filter.StartDate,
		EndDate:   filter.EndDate,
	}
	results, err := s.transactionRepository.GetByPeriod(txn, newFilter)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, errors.New("failed to get transaction")
	}
	s.logger.Infof("get transaction by period of spender id: %d success", req.SpenderId)
	var newResults []GetAllByTxnTypeResponse
	for _, value := range results {
		result := &GetAllByTxnTypeResponse{
			ID:       value.ID,
			Date:     value.Date,
			Amount:   value.Amount,
			Category: value.Category,
			ImageUrl: value.ImageUrl,
		}
		newResults = append(newResults, *result)
	}
	return newResults, nil
}

func (s *transactionService) Update(txnId uint, req Transaction) error {
	txn := entities.Transaction{
		Amount:          req.Amount,
		Category:        req.Category,
		TransactionType: req.TransactionType,
		Note:            req.Note,
	}
	err := s.transactionRepository.UpdateTxn(txnId, txn)
	if err != nil {
		return errors.New("failed to update transaction")
	}
	s.logger.Infof("update transaction with transaction id: %d success", req.ID)
	return nil
}

func (s *transactionService) Delete(spenderId, txnId uint) error {
	err := s.transactionRepository.DeleteTxn(spenderId, txnId)
	if err != nil {
		return errors.New("failed to delete transaction")
	}
	s.logger.Infof("delete transaction with transaction id: %d success", txnId)
	return nil
}

func (s *transactionService) GetAllTxn(filter GetAllTxnFilter, pagination Pagination) ([]GetAllResponse, error) {
	newFilter := entities.GetAllTxnFilter{
		Date:     filter.Date,
		Category: filter.Category,
		TxnType:  filter.TxnType,
	}
	newPagination := entities.Pagination{
		PageItem: pagination.PageItem,
		Page:     pagination.Page,
	}
	results, err := s.transactionRepository.GetAllTxn(newFilter, newPagination)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, errors.New("failed to get transaction")
	}
	s.logger.Infof("get all transaction success")
	var newResults []GetAllResponse
	for _, value := range results {
		result := &GetAllResponse{
			ID:              value.ID,
			Date:            value.Date,
			Amount:          value.Amount,
			Category:        value.Category,
			ImageUrl:        value.ImageUrl,
			TransactionType: value.TransactionType,
		}
		newResults = append(newResults, *result)
	}
	return newResults, nil
}
