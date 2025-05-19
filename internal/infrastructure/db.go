package infrastructure

import (
	"fmt"
	"github.com/thangit93/echo-base/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"sync"
	"time"
)

var (
	dsn         = config.MYSQL_DSN
	maxAttempts = 10
	retryDelay  = 5 * time.Second
)

type DBManager struct {
	mu sync.RWMutex
	db *gorm.DB
}

var manager = &DBManager{}

func GetManager() *DBManager {
	return manager
}

// Connect initial connection
func connectToDB() (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	for i := 1; i <= maxAttempts; i++ {
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
		if err == nil {
			sqlDB, _ := db.DB()
			pingErr := sqlDB.Ping()
			if pingErr == nil {
				log.Println("✅ Connected to MySQL!")
				return db, nil
			}
			err = pingErr
		}
		log.Printf("❌ Attempt %d: %v, retrying in %s...", i, err, retryDelay)
		time.Sleep(retryDelay)
	}
	return nil, fmt.Errorf("failed to connect after %d attempts: %w", maxAttempts, err)
}

func (m *DBManager) Init() error {
	db, err := connectToDB()
	if err != nil {
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// ✅ Connection Pool Settings
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	m.db = db
	return nil
}

func (m *DBManager) GetDB() *gorm.DB {
	m.mu.RLock()
	db := m.db
	m.mu.RUnlock()

	if db != nil {
		if sqlDB, err := db.DB(); err == nil {
			if pingErr := sqlDB.Ping(); pingErr == nil {
				return db
			}
			log.Println("🔄 Lost DB ping, reconnecting...")
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	newDB, err := connectToDB()
	if err != nil {
		log.Fatalf("Unable to reconnect to DB: %v", err)
	}
	m.db = newDB
	return m.db
}
