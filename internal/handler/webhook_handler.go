package handler

import (
	"github.com/gnotnek/agent-allocation/internal/models"
	helper "github.com/gnotnek/agent-allocation/internal/utils"
	"github.com/gofiber/fiber/v2"
)

func HandleAllocateAgent(c *fiber.Ctx) error {
	payload := new(models.QiscusWebhookPayload)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	// Add room to queue if it doesn't exist
	err := helper.AddRoomToQueueIfNeeded(payload.RoomID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to add room to queue"})
	}

	// Check if the room is already assigned to an agent
	assigned, err := helper.IsRoomAssigned(payload.RoomID)
	if err != nil || assigned {
		return c.JSON(fiber.Map{"message": "Room already assigned to an agent"})
	}

	// Get available agents
	agents, err := helper.GetAvailableAgents(payload.RoomID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch available agents"})
	}

	// Select agent with least customers and assign
	err = helper.AssignAgentToRoom(payload.RoomID, agents)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to assign agent to room"})
	}

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

	// Update agent's active room count and mark the room as resolved
	err := helper.ResolveRoomAndUpdateAgent(payload.Service.RoomID, payload.ResolvedBy.ID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to resolve room and update agent"})
	}

	// Assign the next unassigned room in the queue
	err = helper.AssignNextRoomFromQueue()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to assign next room from queue"})
	}

	return c.JSON(fiber.Map{"message": "Room resolved and next room assigned"})
}
