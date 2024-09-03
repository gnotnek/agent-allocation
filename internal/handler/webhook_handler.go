package handler

import (
	"fmt"
	"os"
	"strconv"

	"github.com/gnotnek/agent-allocation/internal/database"
	"github.com/gnotnek/agent-allocation/internal/models"
	helper "github.com/gnotnek/agent-allocation/internal/utils"
	"github.com/gofiber/fiber/v2"
)

func HandleAllocateAgent(c *fiber.Ctx) error {
	payload := new(models.QiscusWebhookPayload)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	if payload.IsResolved {
		return c.JSON(fiber.Map{"message": "Conversation is already resolved."})
	}

	// Fetch available agents
	agents, err := helper.GetAvailableAgents(payload.RoomID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch available agents"})
	}

	maxCustomers, err := strconv.Atoi(os.Getenv("MAX_CUSTOMERS"))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to parse MAX_CUSTOMERS from environment"})
	}

	// Find an agent with less than maxCustomers
	var selectedAgent models.Agent
	customerCount := len(agents)
	for _, agent := range agents {
		if agent.CurrentCustomerCount >= maxCustomers {
			continue
		}
		if agent.CurrentCustomerCount < customerCount {
			selectedAgent = agent
			customerCount = agent.CurrentCustomerCount
		}
	}

	if selectedAgent.ID == 0 {
		// No available agent, add to queue
		err = addToQueue(payload.RoomID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to add room to queue"})
		}
		return c.JSON(fiber.Map{"message": "Room added to queue"})
	}

	// Assign the selected agent
	err = helper.AssignAgent(payload.RoomID, selectedAgent.ID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to assign agent"})
	}

	fmt.Print("done allocating agent\n")
	return c.JSON(fiber.Map{"message": "Agent assigned successfully", "agent_id": selectedAgent.ID})
}

func HandlerMarkAsResolvedWebhook(c *fiber.Ctx) error {
	payload := new(models.MarkAsResolvedPayload)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	if !payload.Service.IsResolved {
		return c.Status(400).JSON(fiber.Map{"error": "Service is not resolved"})
	}

	// Decrease the agent's active room count using the ResolvedBy ID
	err := helper.UpdateAgentRoomCount(payload.ResolvedBy.ID, -1)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update agent's room count"})
	}

	// Assign the next room in the queue
	err = helper.AssignNextRoomInQueue()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to assign the next room in queue"})
	}

	fmt.Print("done mark\n")
	return c.JSON(fiber.Map{"message": "Service marked as resolved successfully"})
}

func addToQueue(roomID string) error {
	var count int64
	database.DB.Model(&models.RoomQueue{}).Where("room_id = ?", roomID).Count(&count)
	if count > 0 {
		return nil
	}

	var max int
	database.DB.Model(&models.RoomQueue{}).Select("COALESCE(MAX(position), 0)").Scan(&max)

	roomQueue := models.RoomQueue{
		RoomID:   roomID,
		Position: max + 1,
	}

	return database.DB.Create(&roomQueue).Error
}
