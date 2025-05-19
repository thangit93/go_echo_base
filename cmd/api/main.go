package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/thangit93/echo-base/internal/infrastructure"
	"github.com/thangit93/echo-base/internal/migrations"
	"github.com/thangit93/echo-base/internal/routers"
)

// dbMiddleware ping and reconnect before each request
func dbMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Set("db", infrastructure.GetManager().GetDB())
		return next(c)
	}
}

func main() {
	// Auto run migrate when start app
	migrations.RunMigration()
	// init DB when start
	if err := infrastructure.GetManager().Init(); err != nil {
		log.Fatalf("❌ DB init failed: %v", err)
	}
	db := infrastructure.GetManager().GetDB()

	e := echo.New()
	e.Use(middleware.Logger(), middleware.Recover(), dbMiddleware)

	// Redis
	if err := infrastructure.InitRedis(); err != nil {
		log.Fatalf("❌ Redis init failed: %v", err)
	}

	routers.InitRouter(e, db)

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
