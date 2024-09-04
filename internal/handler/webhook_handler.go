package handler

import (
	"fmt"

	"github.com/gnotnek/agent-allocation/internal/database"
	"github.com/gnotnek/agent-allocation/internal/models"
	helper "github.com/gnotnek/agent-allocation/internal/utils"
	"github.com/gofiber/fiber/v2"
)

func HandleAllocateAgent(c *fiber.Ctx) error {
	payload := new(models.QiscusWebhookPayload)
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	if payload.IsResolved {
		return c.JSON(fiber.Map{"message": "Conversation is already resolved."})
	}

	err := helper.AssignAgentToRoom(database.DB, payload.RoomID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to assign agent to room"})
	}

	fmt.Print("done allocate\n")
	return c.JSON(fiber.Map{"message": "Agent assigned successfully"})
}

func HandlerMarkAsResolved(c *fiber.Ctx) error {
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
