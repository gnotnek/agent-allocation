package handler

import (
	"github.com/gnotnek/agent-allocation/internal/database"
	"github.com/gnotnek/agent-allocation/internal/models"
	"github.com/gofiber/fiber/v2"
)

type WebhookPayload struct {
	GroupID    string `json:"group_id"`
	CustomerID string `json:"customer_id"`
}

func HandleWebhook(c *fiber.Ctx) error {
	payload := new(WebhookPayload)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(400).SendString(err.Error())
	}

	var agents []models.Agent
	database.DB.Where("status = ?", "online").Find(&agents)
	if len(agents) == 0 {
		return c.Status(404).SendString("No agents available")
	}

	var selectedAgent *models.Agent
	for i := range agents {
		if agents[i].CurrentCustomer < agents[i].MaxCustomer {
			if selectedAgent == nil || agents[i].CurrentCustomer < selectedAgent.CurrentCustomer {
				selectedAgent = &agents[i]
			}
		}
	}

	if selectedAgent == nil {
		return c.Status(500).JSON(fiber.Map{"error": "No suitable agent found"})
	}

	newAssigment := models.Assignment{
		AgentID:    selectedAgent.ID,
		CustomerID: payload.CustomerID,
	}

	database.DB.Create(&newAssigment)

	selectedAgent.CurrentCustomer++
	database.DB.Save(&selectedAgent)

	return c.JSON(fiber.Map{"agent_id": selectedAgent.ID, "message": "Agent assigned successfully"})
}
