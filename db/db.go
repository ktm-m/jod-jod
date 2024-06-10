package db

import (
	"fmt"
	"github.com/Montheankul-K/jod-jod/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"sync"
	"time"
)

type DB interface {
	Connect() *gorm.DB
}

type db struct {
	conn *gorm.DB
}

var (
	dbInstance *db
	once       sync.Once
)

func InitDatabase(cfg *config.Database) DB {
	once.Do(func() {
		dsn := fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=%s",
			cfg.Host, cfg.Port, cfg.Username, cfg.Database, cfg.Password, cfg.SSLMode)
		dial := postgres.Open(dsn)

		logConfig := logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             time.Second,
				LogLevel:                  logger.Silent,
				IgnoreRecordNotFoundError: true,
				ParameterizedQueries:      true,
				Colorful:                  false,
			},
		)

		conn, err := gorm.Open(dial, &gorm.Config{
			Logger: logConfig,
		})
		if err != nil {
			panic("failed to connect database")
		}
		log.Printf("connected to database %s successfully", cfg.Database)

		dbInstance = &db{conn}
	})
	return dbInstance
}

func (db *db) Connect() *gorm.DB {
	return dbInstance.conn
}
