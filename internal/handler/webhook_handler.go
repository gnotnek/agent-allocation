package handler

import (
	"github.com/gnotnek/agent-allocation/internal/database"
	"github.com/gnotnek/agent-allocation/internal/models"
	"github.com/gofiber/fiber/v2"
)

type WebhookPayload struct {
	AppID     string `json:"app_id"`
	Source    string `json:"source"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
	Extras    string `json:"extras"`
	RoomID    string `json:"room_id"`
}

func HandleWebhook(c *fiber.Ctx) error {
	payload := new(WebhookPayload)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(400).SendString(err.Error())
	}

	var customer models.Customer
	database.DB.Where("email = ?", payload.Email).FirstOrCreate(&customer, models.Customer{
		Email:  payload.Email,
		Avatar: payload.AvatarURL,
		RoomID: payload.RoomID,
		AppID:  payload.AppID,
		Status: "waiting",
		Extras: payload.Extras,
	})

	var agents []models.Agent
	database.DB.Where("status = ?", "online").Find(&agents)

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
		CustomerID: string(rune(customer.ID)),
	}

	database.DB.Create(&newAssigment)

	selectedAgent.CurrentCustomer++
	database.DB.Save(&selectedAgent)

	return c.JSON(fiber.Map{"agent_id": selectedAgent.ID, "message": "Agent assigned successfully"})
}
