package user_repository

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

type IUserRepository interface {
	GetUsers(pagination entities.Pagination) ([]entities.GetUserResponse, error)
	GetUser(userId uint) (*entities.GetUserResponse, error)
	GetUserForLogin(username string) (*entities.GetUserForLoginResponse, error)
	CreateUser(req entities.Users) (uint, error)
	UpdateUser(userId uint, req entities.Users) error
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

func (r *userRepository) GetUsers(pagination entities.Pagination) ([]entities.GetUserResponse, error) {
	var res []entities.GetUserResponse
	key := fmt.Sprintf("get-all-users")
	userCache, err := r.redisClient.Get(context.Background(), key).Result()
	if err == nil {
		err = json.Unmarshal([]byte(userCache), &res)
		if err == nil {
			return res, nil
		}
	}

	//offset := (pagination.PageItem - 1) * pagination.Page
	//query := r.db.Model(&entities.Users{}).Offset(offset).Limit(pagination.PageItem)
	query := r.db.Model(&entities.Users{})
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
	fmt.Println(res)
	return res, nil
}

func (r *userRepository) GetUser(userId uint) (*entities.GetUserResponse, error) {
	var res entities.GetUserResponse
	key := fmt.Sprintf("get-user:%d", userId)
	userCache, err := r.redisClient.Get(context.Background(), key).Result()
	if err == nil {
		err = json.Unmarshal([]byte(userCache), &res)
		if err == nil {
			return &res, nil
		}
	}

	query := r.db.Model(&entities.Users{}).Where("id = ?", userId)
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

func (r *userRepository) GetUserForLogin(username string) (*entities.GetUserForLoginResponse, error) {
	var res entities.GetUserForLoginResponse
	query := r.db.Model(&entities.Users{}).Where("username = ?", username)
	err := query.First(&res).Error
	if err != nil {
		r.logger.Error(err)
		return nil, err
	}
	return &res, nil
}

func (r *userRepository) CreateUser(req entities.Users) (uint, error) {
	tx := r.db.Begin()
	err := tx.Create(&req).Error
	if err != nil {
		tx.Rollback()
		r.logger.Error(err)
		return 0, err
	}

	userId := req.ID
	key := "get-all-users"
	err = r.redisClient.Del(context.Background(), key).Err()
	if err != nil {
		tx.Rollback()
		r.logger.Error(err)
		return 0, err
	}
	return userId, tx.Commit().Error
}

func (r *userRepository) UpdateUser(userId uint, req entities.Users) error {
	tx := r.db.Begin()
	var existingUser entities.Users
	if err := r.db.Model(&entities.Users{}).Where("id = ?", userId).First(&existingUser).Error; err != nil {
		tx.Rollback()
		r.logger.Error(err)
		return err
	}

	if req.Firstname != "" {
		existingUser.Firstname = req.Firstname
	}

	if req.Lastname != "" {
		existingUser.Lastname = req.Lastname
	}

	if req.Email != "" {
		existingUser.Email = req.Email
	}

	if err := r.db.Model(&entities.Users{}).Where("id = ?", userId).Save(&existingUser).Error; err != nil {
		tx.Rollback()
		r.logger.Error(err)
		return err
	}

	keys := []string{"get-all-users", fmt.Sprintf("get-user:%d", userId)}
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
	var userInfo entities.Users
	tx := r.db.Begin()
	query := r.db.Model(&entities.Users{}).Where("id = ?", userId)
	err := query.First(&userInfo).Error
	if err != nil {
		r.logger.Error(err)
		return err
	}

	userInfo.Password = newPassword
	if err = r.db.Where("id = ?", userId).Save(&userInfo).Error; err != nil {
		tx.Rollback()
		r.logger.Error(err)
		return err
	}
	return tx.Commit().Error
}

func (r *userRepository) DeleteUser(userId uint) error {
	tx := r.db.Begin()
	if err := r.db.Model(&entities.Users{}).Where("id = ?", userId).Delete(&entities.Users{}).Error; err != nil {
		tx.Rollback()
		r.logger.Error(err)
		return err
	}

	keys := []string{"get-all-users", fmt.Sprintf("get-user:%d", userId)}
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
