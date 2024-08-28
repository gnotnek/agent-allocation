package routes

import (
	"github.com/gnotnek/agent-allocation/internal/handler"
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	//default route
	app.Get("/api", func(c *fiber.Ctx) error {
		return c.SendString("Agent Allocation API")
	})

	//webhook route
	app.Post("/api/webhook", handler.HandleWebhook)
}
