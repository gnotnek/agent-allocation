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
		return fmt.Errorf("cannot parse JSON: %w", err)
	}

	// Check if the room exists
	exists, err := helper.IsRoomAlreadyExists(payload.RoomID)
	if err != nil || !exists {
		// pass to next step
	} else {
		return fmt.Errorf("room already exists: %w", err)
	}

	// Add room to queue if it doesn't exist
	err = helper.AddRoomToQueue(payload.RoomID)
	if err != nil {
		return fmt.Errorf("failed to add room to queue: %w", err)
	}

	// Get available agents
	agents, err := helper.GetAvailableAgents(payload.RoomID)
	if err != nil {
		return fmt.Errorf("failed to get available agents: %w", err)
	}

	// Select agent with least customers and assign
	err = helper.AssignAgentToRoom(payload.RoomID, agents)
	if err != nil {
		return fmt.Errorf("failed to assign agent to room: %w", err)
	}

	return c.JSON(fiber.Map{"message": "Agent assigned successfully"})
}

func HandlerMarkAsResolved(c *fiber.Ctx) error {
	payload := new(models.MarkAsResolvedPayload)
	if err := c.BodyParser(payload); err != nil {
		return fmt.Errorf("cannot parse JSON: %w", err)
	}
	fmt.Printf("Payload: %+v\n", payload)

	// Mark room as resolved
	err := helper.ResolveRoom(payload.Service.RoomID)
	if err != nil {
		return fmt.Errorf("failed to resolve room and update agent: %w", err)
	}

	// Assign the next unassigned room in the queue
	err = helper.AssignNextRoomFromQueue()
	if err != nil {
		return fmt.Errorf("failed to assign next room from queue: %w", err)
	}

	return c.JSON(fiber.Map{"message": "Room resolved and next room assigned"})
}
