package handler

import (
	"fmt"
	"time"

	"github.com/gnotnek/agent-allocation/internal/models"
	helper "github.com/gnotnek/agent-allocation/internal/utils"
	"github.com/gofiber/fiber/v2"
)

func HandleAllocateAgent(c *fiber.Ctx) error {
	payload := new(models.QiscusWebhookPayload)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "cannot parse JSON"})
	}

	// Add room to queue if it doesn't exist
	err := helper.AddRoomToQueue(payload.RoomID)
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
		fmt.Printf("Error parsing JSON: %v\n", err)
		return c.JSON(fiber.Map{"error": "cannot parse JSON"})
	}

	// Mark room as resolved
	err := helper.ResolveRoom(payload.Service.RoomID)
	if err != nil {
		fmt.Printf("Error resolving room: %v\n", err)
		return c.JSON(fiber.Map{"error": "failed to resolve room"})
	}

	time.Sleep(2 * time.Second)
	// Assign the next unassigned room in the queue
	err = helper.AssignNextRoomFromQueue()
	if err != nil {
		fmt.Printf("Error assigning next room: %v\n", err)
		return c.JSON(fiber.Map{"error": "failed to assign next room"})
	}
	fmt.Println("Room resolved and next room assigned")

	return c.JSON(fiber.Map{"message": "Room resolved and next room assigned"})
}
