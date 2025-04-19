package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/whaleship/med-token/internal/config"
	"github.com/whaleship/med-token/internal/database"
)

func main() {
	cfg := config.Load()

	app := fiber.New()
	app.Use(logger.New())

	_ = database.GetInitRedis(cfg)
	log.Fatal(app.Listen(":8080"))
}
