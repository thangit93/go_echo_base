package routers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/thangit93/echo-base/internal/handlers"
	"github.com/thangit93/echo-base/internal/repositories"
	"github.com/thangit93/echo-base/internal/services"
	"gorm.io/gorm"
)

// InitRouter initializes the echo routes
func InitRouter(e *echo.Echo, db *gorm.DB) {
	userRepo := repositories.NewUserRepository(db)
	userService := services.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userService)

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Welcome to Echo API!")
	})

	e.GET("/users", userHandler.GetAllUsers)
	e.POST("/users", userHandler.CreateUser)
}
