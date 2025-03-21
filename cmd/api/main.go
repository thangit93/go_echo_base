package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"echo-base/config"
	"echo-base/internal/handlers"
	"echo-base/internal/repositories"
	"echo-base/internal/services"
)

const (
	dsn         = config.MYSQL_DSN
	maxAttempts = 10
	retryDelay  = 5 * time.Second
)

func connectToDB() (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	for i := 1; i <= maxAttempts; i++ {
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info), // Log thông tin truy vấn
		})
		if err == nil {
			sqlDB, _ := db.DB()
			if err = sqlDB.Ping(); err == nil {
				log.Println("Connected to MySQL successfully!")
				return db, nil
			}
		}
		log.Printf("Attempt %d: Failed to connect to MySQL: %v", i, err)
		log.Printf("Retrying in %v...", retryDelay)
		time.Sleep(retryDelay)
	}

	return nil, fmt.Errorf("failed to connect to MySQL after %d attempts", maxAttempts)
}

func keepDBAlive(db *gorm.DB) {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		sqlDB, err := db.DB()
		if err != nil {
			log.Println("Lost connection to database, reconnecting...")
			db, err = connectToDB()
			if err != nil {
				log.Fatalf("Reconnection failed: %v", err)
			}
		} else {
			err = sqlDB.Ping()
			if err != nil {
				log.Println("Ping failed, reconnecting...")
				db, err = connectToDB()
				if err != nil {
					log.Fatalf("Reconnection failed: %v", err)
				}
			}
		}
	}
}

func main() {
	// Init Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// MySQL connection
	db, err := connectToDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	go keepDBAlive(db)

	// get connection from GORM and check
	sqlDB, err := db.DB()
	defer sqlDB.Close()
	if err != nil {
		log.Fatalf("Failed to get database connection: %v", err)
	}

	// Config to keep alive connection
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	// Redis connection
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.REDIS_ADDR,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// check Redis connection
	ctx := context.Background()
	_, err = rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	// init repository, service, and handler with pointer
	userRepo := repositories.NewUserRepository(db)      // *UserRepository
	userService := services.NewUserService(userRepo)    // *UserService
	userHandler := handlers.NewUserHandler(userService) // *UserHandler

	// Routes
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

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
