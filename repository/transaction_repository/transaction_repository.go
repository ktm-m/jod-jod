package transaction_repository

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Montheankul-K/jod-jod/domains/entities"
	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"time"
)

type ITransactionRepository interface {
	GetAllTxn(filter entities.GetAllTxnFilter, pagination entities.Pagination) ([]entities.GetAllResponse, error)
	GetAllBySpenderId(spenderId uint) ([]entities.GetAllResponse, error)
	GetByTxnType(req entities.GetByTxnTypeRequest) ([]entities.GetAllByTxnTypeResponse, error)
	GetByCategory(req entities.GetByCategoryRequest) ([]entities.GetByCategoryResponse, error)
	GetByPeriod(req entities.GetByTxnTypeRequest, filter entities.PeriodFilter) ([]entities.GetAllByTxnTypeResponse, error)
	SaveTxn(req entities.Transaction) (uint, error)
	UpdateTxn(txnId uint, req entities.Transaction) error
	DeleteTxn(spenderId uint, txnId uint) error
}

type transactionRepository struct {
	db          *gorm.DB
	logger      echo.Logger
	redisClient *redis.Client
}

func NewTransactionRepository(db *gorm.DB, logger echo.Logger, redisClient *redis.Client) ITransactionRepository {
	return &transactionRepository{
		db:          db,
		logger:      logger,
		redisClient: redisClient,
	}
}

func (r *transactionRepository) GetAllTxn(filter entities.GetAllTxnFilter, pagination entities.Pagination) ([]entities.GetAllResponse, error) {
	var res []entities.GetAllResponse
	var err error
	key := fmt.Sprintf("get-all-txn")
	txnCache, err := r.redisClient.Get(context.Background(), key).Result()
	if err == nil && txnCache != "" {
		err = json.Unmarshal([]byte(txnCache), &res)
		if err == nil {
			return res, nil
		}
	}

	query := r.db.Model(&entities.Transaction{})
	//if filter.Date != nil {
	//	query = query.Where("date = ?", filter.Date)
	//}
	//
	//if filter.Category != "" {
	//	query = query.Where("category = ?", filter.Category)
	//}
	//
	//if filter.TxnType != "" {
	//	query = query.Where("transaction_type = ?", filter.TxnType)
	//}

	//offset := (pagination.PageItem - 1) * pagination.Page
	//query = query.Offset(offset).Limit(pagination.Page)

	err = query.Find(&res).Error
	if err != nil {
		r.logger.Error(err)
		return nil, err
	}

	cache, err := json.Marshal(res)
	if err != nil {
		r.logger.Error(err)
		return nil, err
	}

	err = r.redisClient.Set(context.Background(), key, string(cache), time.Hour*1).Err()
	if err != nil {
		r.logger.Error(err)
		return nil, err
	}
	return res, nil
}

func (r *transactionRepository) GetAllBySpenderId(spenderId uint) ([]entities.GetAllResponse, error) {
	var res []entities.GetAllResponse
	var err error
	key := fmt.Sprintf("get-all-spender:%v", spenderId)
	txnCache, err := r.redisClient.Get(context.Background(), key).Result()
	if err == nil && txnCache != "" {
		err = json.Unmarshal([]byte(txnCache), &res)
		if err == nil {
			return res, nil
		}
	}

	query := r.db.Model(&entities.Transaction{}).Where("spender_id = ?", spenderId)
	err = query.Find(&res).Error
	if err != nil {
		r.logger.Error(err)
		return nil, err
	}

	cache, err := json.Marshal(res)
	if err != nil {
		r.logger.Error(err)
		return nil, err
	}

	err = r.redisClient.Set(context.Background(), key, string(cache), time.Hour*1).Err()
	if err != nil {
		r.logger.Error(err)
		return nil, err
	}
	return res, nil
}

func (r *transactionRepository) GetByTxnType(req entities.GetByTxnTypeRequest) ([]entities.GetAllByTxnTypeResponse, error) {
	var res []entities.GetAllByTxnTypeResponse
	var err error
	key := fmt.Sprintf("get-by-txn-type:%v", req.SpenderId)
	txnCache, err := r.redisClient.Get(context.Background(), key).Result()
	if err == nil && txnCache != "" {
		err = json.Unmarshal([]byte(txnCache), &res)
		if err == nil {
			return res, nil
		}
	}

	query := r.db.Model(&entities.Transaction{}).Where("spender_id = ? AND transaction_type = ?", req.SpenderId, req.TxnType)
	err = query.Find(&res).Error
	if err != nil {
		r.logger.Error(err)
		return nil, err
	}

	cache, err := json.Marshal(res)
	if err != nil {
		r.logger.Error(err)
		return nil, err
	}

	err = r.redisClient.Set(context.Background(), key, string(cache), time.Hour*1).Err()
	if err != nil {
		r.logger.Error(err)
		return nil, err
	}
	return res, nil
}

func (r *transactionRepository) GetByCategory(req entities.GetByCategoryRequest) ([]entities.GetByCategoryResponse, error) {
	var results []entities.Transaction
	var res []entities.GetByCategoryResponse
	query := r.db.Model(&entities.Transaction{}).Where("spender_id = ? AND transaction_type = ? AND category = ?", req.SpenderId, req.TxnType, req.Category)

	err := query.Find(&results).Error
	if err != nil {
		r.logger.Error(err)
		return nil, err
	}
	for _, value := range results {
		result := entities.GetByCategoryResponse{
			ID:       value.ID,
			Date:     value.Date,
			Amount:   value.Amount,
			ImageUrl: value.ImageUrl,
		}
		res = append(res, result)
	}

	return res, nil
}

func (r *transactionRepository) GetByPeriod(req entities.GetByTxnTypeRequest, filter entities.PeriodFilter) ([]entities.GetAllByTxnTypeResponse, error) {
	var res []entities.GetAllByTxnTypeResponse
	if filter.EndDate.IsZero() {
		defaultEndDate := time.Now()
		filter.EndDate = &defaultEndDate
	}

	query := r.db.Model(&entities.Transaction{}).Where("spender_id = ? AND transaction_type =? AND date >= ? AND date <= ?", req.SpenderId, req.TxnType, filter.StartDate, filter.EndDate)
	err := query.Find(&res).Error
	if err != nil {
		fmt.Println(err)
		r.logger.Error(err)
		return nil, err
	}
	return res, nil
}

func (r *transactionRepository) SaveTxn(req entities.Transaction) (uint, error) {
	tx := r.db.Begin()
	result := r.db.Create(&req)
	if err := result.Error; err != nil {
		tx.Rollback()
		r.logger.Error(err)
		return 0, err
	}

	txnId := req.ID
	var keys []string
	keys = append(keys, "get-all-txn")
	keys = append(keys, fmt.Sprintf("get-all-spender:%v", req.SpenderId))
	keys = append(keys, fmt.Sprintf("get-by-txn-type:%v", req.SpenderId))

	for _, key := range keys {
		err := r.redisClient.Del(context.Background(), key).Err()
		if err != nil {
			tx.Rollback()
			r.logger.Error(err)
			return 0, err
		}
	}
	return txnId, tx.Commit().Error
}

func (r *transactionRepository) UpdateTxn(txnId uint, req entities.Transaction) error {
	var existingTxn entities.Transaction
	tx := r.db.Begin()
	query := r.db.Model(&entities.Transaction{}).Where("id = ?", txnId)
	err := query.First(&existingTxn).Error
	if err != nil {
		r.logger.Error(err)
		return err
	}

	if req.Amount != 0 {
		existingTxn.Amount = req.Amount
	}

	if req.Category != "" {
		existingTxn.Category = req.Category
	}

	if req.TransactionType != "" {
		existingTxn.TransactionType = req.TransactionType
	}

	if req.Note != "" {
		existingTxn.Note = req.Note
	}

	if err = r.db.Model(&entities.Transaction{}).Where("id = ?", txnId).Save(&existingTxn).Error; err != nil {
		tx.Rollback()
		r.logger.Error(err)
		return err
	}

	var keys []string
	keys = append(keys, "get-all-txn")
	keys = append(keys, fmt.Sprintf("get-all-spender:%v", req.SpenderId))
	keys = append(keys, fmt.Sprintf("get-by-txn-type:%v", req.SpenderId))

	for _, key := range keys {
		err := r.redisClient.Del(context.Background(), key).Err()
		if err != nil {
			tx.Rollback()
			r.logger.Error(err)
			return err
		}
	}
	return tx.Commit().Error
}

func (r *transactionRepository) DeleteTxn(spenderId uint, txnId uint) error {
	tx := r.db.Begin()
	if err := r.db.Model(&entities.Transaction{}).Where("id = ?", txnId).Delete(&entities.Transaction{}).Error; err != nil {
		tx.Rollback()
		r.logger.Error(err)
		return err
	}

	var keys []string
	keys = append(keys, "get-all-txn")
	keys = append(keys, fmt.Sprintf("get-all-spender:%v", spenderId))
	keys = append(keys, fmt.Sprintf("get-by-txn-type:%v", spenderId))

	for _, key := range keys {
		err := r.redisClient.Del(context.Background(), key).Err()
		if err != nil {
			tx.Rollback()
			r.logger.Error(err)
			return err
		}
	}

	return tx.Commit().Error
}
