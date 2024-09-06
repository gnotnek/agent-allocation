package handler

import (
	"fmt"

	"github.com/gnotnek/agent-allocation/internal/models"
	helper "github.com/gnotnek/agent-allocation/internal/utils"
	"github.com/gofiber/fiber/v2"
)

func HandleAllocateAgent(c *fiber.Ctx) error {
	payload := new(models.QiscusWebhookPayload)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "cannot parse JSON"})
	}

	// Check if the room exists
	exists, err := helper.IsRoomAlreadyExists(payload.RoomID)
	if err != nil || !exists {
		// pass to next step
	} else {
		return c.JSON(fiber.Map{"message": "Room already exists"})
	}

	// Add room to queue if it doesn't exist
	err = helper.AddRoomToQueue(payload.RoomID)
	if err != nil {
		return c.JSON(fiber.Map{"error": "failed to add room to queue"})
	}

	// Get available agents
	agents, err := helper.GetAvailableAgents(payload.RoomID)
	if err != nil {
		return c.JSON(fiber.Map{"error": "failed to get available agents"})
	}

	// Select agent with least customers and assign
	err = helper.AssignAgentToRoom(payload.RoomID, agents)
	if err != nil {
		return c.JSON(fiber.Map{"error": "failed to assign agent"})
	}

	return c.JSON(fiber.Map{"message": "Agent assigned successfully"})
}

func HandlerMarkAsResolved(c *fiber.Ctx) error {
	payload := new(models.MarkAsResolvedPayload)
	if err := c.BodyParser(payload); err != nil {
		return c.JSON(fiber.Map{"error": "cannot parse JSON"})
	}
	fmt.Printf("Payload: %+v\n", payload)

	// Mark room as resolved
	err := helper.ResolveRoom(payload.Service.RoomID)
	if err != nil {
		return c.JSON(fiber.Map{"error": "failed to resolve room"})
	}

	// Assign the next unassigned room in the queue
	err = helper.AssignNextRoomFromQueue()
	if err != nil {
		return c.JSON(fiber.Map{"error": "failed to assign next room"})
	}

	return c.JSON(fiber.Map{"message": "Room resolved and next room assigned"})
}
