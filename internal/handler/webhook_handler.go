package handler

import (
	"fmt"

	"github.com/gnotnek/agent-allocation/internal/helper"
	"github.com/gnotnek/agent-allocation/internal/models"
	"github.com/gofiber/fiber/v2"
)

func HandleAllocateAgent(c *fiber.Ctx) error {
	payload := new(models.QiscusWebhookPayload)
	if err := c.BodyParser(payload); err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		return c.Status(400).JSON(fiber.Map{"error": "cannot parse JSON"})
	}

	// Add room to queue if it doesn't exist
	err := helper.AddRoomToQueue(payload.RoomID)
	if err != nil {
		fmt.Printf("Error adding room to queue: %v\n", err)
		return c.JSON(fiber.Map{"error": "failed to add room to queue"})
	}

	go func(roomId string) {
		// Get available agents
		agents, err := helper.GetAvailableAgents(roomId)
		if err != nil {
			fmt.Printf("Error getting available agents: %v\n", err)
			return
		}

		// Select agent with least customers and assign
		err = helper.AssignAgentToRoom(roomId, agents)
		if err != nil {
			fmt.Printf("Error assigning agent: %v\n", err)
		}
	}(payload.RoomID)

	return c.JSON(fiber.Map{"message": "Room added to queue"})
}

func HandlerMarkAsResolved(c *fiber.Ctx) error {
	payload := new(models.MarkAsResolvedPayload)
	if err := c.BodyParser(payload); err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		return c.JSON(fiber.Map{"error": "cannot parse JSON"})
	}

	go func() {
		err := helper.MarkRoomAsResolved(payload.Service.RoomID)
		if err != nil {
			fmt.Printf("Error marking room as resolved: %v\n", err)
		}

		// Assign the next unassigned room in the queue
		err = helper.AssignNextRoomFromQueue()
		if err != nil {
			fmt.Printf("Error assigning next room: %v\n", err)
		}
	}()

	return c.JSON(fiber.Map{"message": "Room resolved and next room assigned"})
}
