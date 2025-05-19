package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/thangit93/echo-base/config"
	"github.com/thangit93/echo-base/internal/handlers"
	"github.com/thangit93/echo-base/internal/repositories"
	"github.com/thangit93/echo-base/internal/services"
)

var (
	dsn         = config.MYSQL_DSN
	maxAttempts = 10
	retryDelay  = 5 * time.Second
)

// DBManager giữ instance *gorm.DB và mutex
type DBManager struct {
	mu sync.RWMutex
	db *gorm.DB
}

var manager = &DBManager{}

// connectToDB to connect to MySQL
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

// GetDB return pointer *gorm.DB, if lost connection, it will reconnect
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

	// need reconnect
	m.mu.Lock()
	defer m.mu.Unlock()
	newDB, err := connectToDB()
	if err != nil {
		log.Fatalf("Unable to reconnect to DB: %v", err)
	}
	m.db = newDB
	return m.db
}

// dbMiddleware ping and reconnect before each request
func dbMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Set("db", manager.GetDB())
		return next(c)
	}
}

func main() {
	// init DB when start
	initialDB, err := connectToDB()
	if err != nil {
		log.Fatalf("Initial DB connection failed: %v", err)
	}
	manager.db = initialDB

	e := echo.New()
	e.Use(middleware.Logger(), middleware.Recover(), dbMiddleware)

	// setting pool
	sqlDB, _ := manager.db.DB()
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	// Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.REDIS_ADDR,
		Password: "",
		DB:       0,
	})
	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		log.Fatalf("Redis connect failed: %v", err)
	}

	userRepo := repositories.NewUserRepository(manager.db)
	userService := services.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userService)

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Welcome to Echo API!")
	})
	e.GET("/users", userHandler.GetAllUsers)
	e.POST("/users", userHandler.CreateUser)

	// Start server
	go func() {
		if err := e.Start(":8080"); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
