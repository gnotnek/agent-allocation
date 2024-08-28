package handler

import (
	"fmt"

	"github.com/gnotnek/agent-allocation/internal/database"
	"github.com/gnotnek/agent-allocation/internal/models"
	"github.com/gofiber/fiber/v2"
)

type CandidateAgent struct {
	ID                  int      `json:"id"`
	Name                string   `json:"name"`
	Email               string   `json:"email"`
	AuthenticationToken string   `json:"authentication_token"`
	SdkEmail            string   `json:"sdk_email"`
	SdkKey              string   `json:"sdk_key"`
	IsAvailable         bool     `json:"is_available"`
	Type                int      `json:"type"`
	AvatarURL           string   `json:"avatar_url"`
	AppID               int      `json:"app_id"`
	IsVerified          bool     `json:"is_verified"`
	NotificationsRoomID string   `json:"notifications_room_id"`
	BubbleColor         string   `json:"bubble_color"`
	QismoKey            string   `json:"qismo_key"`
	DirectLoginToken    string   `json:"direct_login_token"`
	TypeAsString        string   `json:"type_as_string"`
	AssignedRules       []string `json:"assigned_rules"`
}

type LatestService struct {
	ID                    int    `json:"id"`
	UserID                int    `json:"user_id"`
	RoomLogID             int    `json:"room_log_id"`
	AppID                 int    `json:"app_id"`
	RoomID                string `json:"room_id"`
	Notes                 string `json:"notes"`
	ResolvedAt            string `json:"resolved_at"`
	IsResolved            bool   `json:"is_resolved"`
	CreatedAt             string `json:"created_at"`
	UpdatedAt             string `json:"updated_at"`
	FirstCommentID        string `json:"first_comment_id"`
	LastCommentID         string `json:"last_comment_id"`
	RetrievedAt           string `json:"retrieved_at"`
	FirstCommentTimestamp string `json:"first_comment_timestamp"`
}

type QiscusWebhookPayload struct {
	AppID          string         `json:"app_id"`
	Source         string         `json:"source"`
	Name           string         `json:"name"`
	Email          string         `json:"email"`
	AvatarURL      string         `json:"avatar_url"`
	Extras         string         `json:"extras"`
	IsResolved     bool           `json:"is_resolved"`
	LatestService  LatestService  `json:"latest_service"`
	RoomID         string         `json:"room_id"`
	CandidateAgent CandidateAgent `json:"candidate_agent"`
}

func HandleAllocateAgent(c *fiber.Ctx) error {
	payload := new(QiscusWebhookPayload)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Cannot parse webhook payload",
		})
	}

	if payload.IsResolved {
		return c.Status(200).JSON(fiber.Map{
			"message": "Webhook is resolved",
		})
	}

	if payload.CandidateAgent.IsAvailable {
		err := customAgentAllocationLogic(payload)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
	}

	return c.JSON(fiber.Map{"message": "Custom agent allocation success"})
}

func customAgentAllocationLogic(payload *QiscusWebhookPayload) error {
	var agent models.Agent
	database.DB.Where("status = ? AND max_customers > current_customers", "online").Order("created_at ASC").First(&agent)
	if agent.ID == 0 {
		return fmt.Errorf("no available agent")
	}

	err := assignAgentToRoom(payload)
	if err != nil {
		return err
	}

	agent.CurrentCustomers++
	database.DB.Save(&agent)

	return nil
}

func assignAgentToRoom(payload *QiscusWebhookPayload) error {
	var customer models.Customer
	database.DB.Where("status = ?", "waiting").Order("created_at ASC").First(&customer)

	if customer.ID == 0 {
		return fmt.Errorf("no waiting customer")
	}

	var selectedAgent models.Agent
	database.DB.Where("status = ? AND current_customers < max_customers", "online").First(&selectedAgent)

	if selectedAgent.ID == 0 {
		return fmt.Errorf("no available agent")
	}

	customer.Status = "served"
	customer.RoomID = payload.RoomID
	database.DB.Save(&customer)

	return nil
}

func HandlerMarkAsResolvedWebhook(c *fiber.Ctx) error {
	payload := new(QiscusWebhookPayload)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	if !payload.IsResolved {
		return c.Status(400).JSON(fiber.Map{"error": "Service is not resolved"})
	}

	// Find the service by RoomID and update its status
	var service models.Service
	result := database.DB.Where("room_id = ?", payload.LatestService.RoomID).First(&service)

	if result.Error != nil {
		// If the service doesn't exist, create a new one
		service = models.Service{
			RoomID:          payload.LatestService.RoomID,
			IsResolved:      payload.LatestService.IsResolved,
			Notes:           payload.LatestService.Notes,
			FirstCommentID:  payload.LatestService.FirstCommentID,
			LastCommentID:   payload.LatestService.LastCommentID,
			Source:          payload.Source,
			ResolvedBy:      payload.CandidateAgent.Name,
			ResolvedByEmail: payload.CandidateAgent.Email,
		}
		database.DB.Create(&service)
	} else {
		// Update the existing service
		service.IsResolved = payload.LatestService.IsResolved
		service.Notes = payload.LatestService.Notes
		service.FirstCommentID = payload.LatestService.FirstCommentID
		service.LastCommentID = payload.LatestService.LastCommentID
		service.Source = payload.Source
		service.ResolvedBy = payload.CandidateAgent.Name
		service.ResolvedByEmail = payload.CandidateAgent.Email
		database.DB.Save(&service)
	}

	return c.JSON(fiber.Map{
		"message":    "Service marked as resolved successfully",
		"service_id": service.ID,
	})
}
