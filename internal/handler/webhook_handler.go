package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	assignedAgentID, err := assignAgentToRoom(payload)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"agent_id": assignedAgentID,
		"message":  "Agent assigned successfully",
	})

}

func assignAgentToRoom(payload *WebhookPayload) (uint, error) {
	var selectedAgent models.Agent
	database.DB.Where("status = ? AND current_customer < max_customer", "online").First(&selectedAgent)

	if selectedAgent.ID == 0 {
		return 0, errors.New("no available agent")
	}

	arr := hitAssignmentAPI(payload.RoomID, selectedAgent.ID)
	if arr != nil {
		return 0, arr
	}

	selectedAgent.CurrentCustomer++
	database.DB.Save(&selectedAgent)

	return selectedAgent.ID, nil
}

func hitAssignmentAPI(roomID string, agentID uint) error {
	assigAgentURL := "https://multichannel.qiscus.com//api/v1/admin/service/assign_agent"

	payload := strings.NewReader(fmt.Sprintf("room_id=%s&agent_id=%d&replace_latest_agent=1&max_agent=1", roomID, agentID))

	req, err := http.NewRequest("POST", assigAgentURL, payload)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Qiscus-App-Id", "AppCode")
	req.Header.Set("Qiscus-Secret-Key", "SecretKey")

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("error assigning agent")
	}

	return nil
}
