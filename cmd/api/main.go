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

func main() {
	// Khởi tạo Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Kết nối MySQL
	db, err := gorm.Open(mysql.Open(config.MySQLDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Kết nối Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// Kiểm tra kết nối Redis
	ctx := context.Background()
	_, err = rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	// Khởi tạo repository, service, và handler với con trỏ
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
