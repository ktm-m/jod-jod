package transaction_repository

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Montheankul-K/jod-jod/domains/transaction"
	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"time"
)

type ITransactionRepository interface {
	GetAllTxn(filter transaction.GetAllTxnFilter, pagination transaction.Pagination) ([]transaction.GetAllResponse, error)
	GetAllBySpenderId(spenderId uint) ([]transaction.GetAllResponse, error)
	GetByTxnType(req transaction.GetByTxnTypeRequest) ([]transaction.GetAllByTxnTypeResponse, error)
	GetByCategory(req transaction.GetByCategoryRequest) ([]transaction.GetByCategoryResponse, error)
	GetByPeriod(req transaction.GetByTxnTypeRequest, filter transaction.PeriodFilter) ([]transaction.GetAllByTxnTypeResponse, error)
	SaveTxn(req transaction.Transaction) (uint, error)
	UpdateTxn(req transaction.Transaction) error
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

func (r *transactionRepository) GetAllTxn(filter transaction.GetAllTxnFilter, pagination transaction.Pagination) ([]transaction.GetAllResponse, error) {
	var res []transaction.GetAllResponse
	key := fmt.Sprintf("get-all-expenses:%v", filter)
	txnCache, err := r.redisClient.Get(context.Background(), key).Result()
	if err == nil && txnCache != "" {
		err = json.Unmarshal([]byte(txnCache), &res)
		if err == nil {
			return res, nil
		}
	}

	query := r.db.Model(&transaction.Transaction{})
	if filter.Date != nil {
		query = query.Where("date = ?", filter.Date)
	}

	if filter.Category != "" {
		query = query.Where("category = ?", filter.Category)
	}

	if filter.TxnType != "" {
		query = query.Where("transaction_type = ?", filter.TxnType)
	}

	offset := (pagination.PageItem - 1) * pagination.Page
	query = query.Offset(offset).Limit(pagination.Page)

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

	err = r.redisClient.Set(context.Background(), key, string(cache), time.Second*10).Err()
	if err != nil {
		r.logger.Error(err)
		return nil, err
	}
	return res, nil
}

func (r *transactionRepository) GetAllBySpenderId(spenderId uint) ([]transaction.GetAllResponse, error) {
	var res []transaction.GetAllResponse
	key := fmt.Sprintf("get-all-spender:%v", spenderId)
	txnCache, err := r.redisClient.Get(context.Background(), key).Result()
	if err == nil && txnCache != "" {
		err = json.Unmarshal([]byte(txnCache), &res)
		if err == nil {
			return res, nil
		}
	}

	query := r.db.Model(&transaction.Transaction{}).Where("spender_id = ?", spenderId)
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

func (r *transactionRepository) GetByTxnType(req transaction.GetByTxnTypeRequest) ([]transaction.GetAllByTxnTypeResponse, error) {
	var res []transaction.GetAllByTxnTypeResponse
	key := fmt.Sprintf("get-by-txn-type:%v", req.SpenderId)
	txnCache, err := r.redisClient.Get(context.Background(), key).Result()
	if err == nil && txnCache != "" {
		err = json.Unmarshal([]byte(txnCache), &res)
		if err == nil {
			return res, nil
		}
	}

	query := r.db.Model(&transaction.Transaction{}).Where("spender_id = ? AND transaction_type = ?", req.SpenderId, req.TxnType)
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

func (r *transactionRepository) GetByCategory(req transaction.GetByCategoryRequest) ([]transaction.GetByCategoryResponse, error) {
	var res []transaction.GetByCategoryResponse
	query := r.db.Model(&transaction.Transaction{}).Where("spender_id = ? AND transaction_type = ? AND category = ?", req.SpenderId, req.TxnType, req.Category)

	err := query.Find(&res).Error
	if err != nil {
		r.logger.Error(err)
		return nil, err
	}
	return res, nil
}

func (r *transactionRepository) GetByPeriod(req transaction.GetByTxnTypeRequest, filter transaction.PeriodFilter) ([]transaction.GetAllByTxnTypeResponse, error) {
	var res []transaction.GetAllByTxnTypeResponse
	query := r.db.Model(&transaction.Transaction{}).Where("spender_id = ? AND transaction_type = ?", req.SpenderId, req.TxnType)
	if filter.StartDate != nil {
		query = query.Where("date >= ?", filter.StartDate)
	}

	if filter.EndDate != nil {
		query = query.Where("date <= ?", filter.EndDate)
	}

	err := query.Find(&res).Error
	if err != nil {
		r.logger.Error(err)
		return nil, err
	}
	return res, nil
}

func (r *transactionRepository) SaveTxn(req transaction.Transaction) (uint, error) {
	tx := r.db.Begin()
	result := r.db.Create(&req)
	if err := result.Error; err != nil {
		tx.Rollback()
		r.logger.Error(err)
		return 0, err
	}

	var txnId uint
	if err := r.db.Raw("SELECT LAST_INSERT_ID()").Scan(&txnId).Error; err != nil {
		tx.Rollback()
		r.logger.Error(err)
		return 0, err
	}

	var keys []string
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

func (r *transactionRepository) UpdateTxn(req transaction.Transaction) error {
	tx := r.db.Begin()
	if err := r.db.Model(&transaction.Transaction{}).Where("id = ?", req.ID).Updates(req).Error; err != nil {
		tx.Rollback()
		r.logger.Error(err)
		return err
	}

	var keys []string
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
	if err := r.db.Model(&transaction.Transaction{}).Where("id = ?", txnId).Delete(&transaction.Transaction{}).Error; err != nil {
		tx.Rollback()
		r.logger.Error(err)
		return err
	}

	var keys []string
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
