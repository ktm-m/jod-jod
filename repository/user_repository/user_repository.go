package user_repository

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Montheankul-K/jod-jod/domains/user"
	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"time"
)

type IUserRepository interface {
	GetUsers(pagination user.Pagination) ([]user.GetUserResponse, error)
	GetUser(userId uint) (*user.GetUserResponse, error)
	GetUserForLogin(username string) (*user.GetUserForLoginResponse, error)
	CreateUser(req user.Users) (uint, error)
	UpdateUser(req user.Users) error
	UpdatePassword(userId uint, newPassword string) error
	DeleteUser(userId uint) error
}

type userRepository struct {
	db          *gorm.DB
	logger      echo.Logger
	redisClient *redis.Client
}

func NewUserRepository(db *gorm.DB, logger echo.Logger, redisClient *redis.Client) IUserRepository {
	return &userRepository{
		db:          db,
		logger:      logger,
		redisClient: redisClient,
	}
}

func (r *userRepository) GetUsers(pagination user.Pagination) ([]user.GetUserResponse, error) {
	var res []user.GetUserResponse
	key := fmt.Sprintf("get-all-users")
	userCache, err := r.redisClient.Get(context.Background(), key).Result()
	if err == nil {
		err = json.Unmarshal([]byte(userCache), &res)
		if err == nil {
			return res, nil
		}
	}

	offset := (pagination.PageItem - 1) * pagination.Page
	query := r.db.Model(&user.Users{}).Offset(offset).Limit(pagination.Page)

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

	err = r.redisClient.Set(context.Background(), key, string(cache), time.Minute*1).Err()
	if err != nil {
		r.logger.Error(err)
		return nil, err
	}
	return res, nil
}

func (r *userRepository) GetUser(userId uint) (*user.GetUserResponse, error) {
	var res user.GetUserResponse
	key := fmt.Sprintf("get-user:%d", userId)
	userCache, err := r.redisClient.Get(context.Background(), key).Result()
	if err == nil {
		err = json.Unmarshal([]byte(userCache), &res)
		if err == nil {
			return &res, nil
		}
	}

	query := r.db.Model(&user.Users{}).Where("id = ?", userId)
	err = query.First(&res).Error
	if err != nil {
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
	return &res, nil
}

func (r *userRepository) GetUserForLogin(username string) (*user.GetUserForLoginResponse, error) {
	var res user.GetUserForLoginResponse
	query := r.db.Model(&user.Users{}).Where("username = ?", username)
	err := query.First(&res).Error
	if err != nil {
		r.logger.Error(err)
		return nil, err
	}
	return &res, nil
}

func (r *userRepository) CreateUser(req user.Users) (uint, error) {
	tx := r.db.Begin()
	err := tx.Create(req).Error
	if err != nil {
		tx.Rollback()
		r.logger.Error(err)
		return 0, err
	}

	var userId uint
	if err = r.db.Raw("SELECT LAST_INSERT_ID()").Scan(&userId).Error; err != nil {
		tx.Rollback()
		r.logger.Error(err)
		return 0, err
	}

	key := "get-all-users"
	err = r.redisClient.Del(context.Background(), key).Err()
	if err != nil {
		tx.Rollback()
		r.logger.Error(err)
		return 0, err
	}
	return userId, tx.Commit().Error
}

func (r *userRepository) UpdateUser(req user.Users) error {
	tx := r.db.Begin()
	if err := r.db.Model(&user.Users{}).Where("id = ?", req.ID).Updates(req).Error; err != nil {
		tx.Rollback()
		r.logger.Error(err)
		return err
	}

	keys := []string{"update-all-users", fmt.Sprintf("get-user:%d", req.ID)}
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

func (r *userRepository) UpdatePassword(userId uint, newPassword string) error {
	var userInfo user.Users
	tx := r.db.Begin()
	query := r.db.Model(&user.Users{}).Where("id = ?", userId)
	err := query.First(&userInfo).Error
	if err != nil {
		r.logger.Error(err)
		return err
	}

	userInfo.Password = newPassword
	if err = r.db.Save(&userInfo).Error; err != nil {
		tx.Rollback()
		r.logger.Error(err)
		return err
	}
	return tx.Commit().Error
}

func (r *userRepository) DeleteUser(userId uint) error {
	tx := r.db.Begin()
	if err := r.db.Model(&user.Users{}).Where("id = ?", userId).Delete(&user.Users{}).Error; err != nil {
		tx.Rollback()
		r.logger.Error(err)
		return err
	}

	keys := []string{"update-all-users", fmt.Sprintf("get-user:%d", userId)}
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
