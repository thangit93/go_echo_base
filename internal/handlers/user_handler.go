package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/thangit93/echo-base/internal/services"
)

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) GetAllUsers(c echo.Context) error {
	users, err := h.userService.GetAllUsers()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, users)
}

func (h *UserHandler) CreateUser(c echo.Context) error {
	var req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}
	if err := h.userService.CreateUser(req.Name, req.Email); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, map[string]string{"message": "User created"})
}
