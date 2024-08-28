package main

import (
	"github.com/gnotnek/agent-allocation/internal/database"
	"github.com/gnotnek/agent-allocation/internal/routes"
	"github.com/gofiber/fiber/v2"
)

func main() {
	database.ConnectDatabase()
	app := fiber.New()

	routes.SetupRoutes(app)

	app.Listen(":3000")
}
