package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/whaleship/med-token/internal/config"
	"github.com/whaleship/med-token/internal/database"
	"github.com/whaleship/med-token/internal/handlers"
	"github.com/whaleship/med-token/internal/repository"
	"github.com/whaleship/med-token/internal/service"
)

func main() {
	cfg := config.Load()

	app := fiber.New()
	app.Use(logger.New())
	rConn := database.GetInitRedis(cfg)

	authRepo := repository.NewRefreshRepo(rConn)
	emailSvc := service.NewSMTPEmailService(
		cfg.SMTPHost,
		cfg.SMTPPort,
		cfg.SMTPUsername,
		cfg.SMTPPassword,
	)
	authSvc := service.NewAuthService(cfg.JWTSecret, authRepo, emailSvc)
	handlers := handlers.NewAuthHandler(authSvc)

	app.Get("/token", handlers.GetTokens)
	app.Post("/refresh", handlers.Refresh)
	log.Fatal(app.Listen(":8080"))
}
