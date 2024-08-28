package main

import (
	"fmt"

	"github.com/gnotnek/agent-allocation/internal/database"
	"github.com/gnotnek/agent-allocation/internal/handler"
	"github.com/gnotnek/agent-allocation/internal/routes"
	"github.com/gofiber/fiber/v2"
)

func main() {
	database.InitDB()
	app := fiber.New()

	err := handler.SetMarkAsResolvedWebhook("webhookurl", "app_id", "admin_token", true)
	if err != nil {
		fmt.Printf("error setting mark as resolve webhook: %v", err)
		return
	}
	routes.SetupRoutes(app)

	app.Listen(":3000")
}
