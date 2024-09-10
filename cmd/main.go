package main

import (
	"fmt"
	"time"

	"github.com/gnotnek/agent-allocation/internal/cron"
	"github.com/gnotnek/agent-allocation/internal/database"
	"github.com/gnotnek/agent-allocation/internal/routes"
	"github.com/go-co-op/gocron"
	"github.com/gofiber/fiber/v2"
)

func main() {
	database.ConnectDatabase()
	app := fiber.New()

	routes.SetupRoutes(app)
	runCron()

	app.Listen(":3000")
}

func runCron() {
	s := gocron.NewScheduler(time.UTC)
	s.Every(1).Minutes().Do(func() {
		fmt.Println("Running cron job")
		err := cron.CronAssignAgent()
		if err != nil {
			fmt.Printf("Error assigning agent: %v\n", err)
		}
	})

	s.StartAsync()
}
