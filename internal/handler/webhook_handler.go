package handler

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

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

type WebhookPayload struct {
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

func HandleWebhook(c *fiber.Ctx) error {
	payload := new(WebhookPayload)

	if err := c.BodyParser(payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	if payload.IsResolved {
		return c.JSON(fiber.Map{
			"message": "Room is already resolved",
		})
	}

	assignedAgentID, err := assignAgentToRoom(payload)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"agent_id": assignedAgentID,
		"message":  "Agent assigned successfully",
	})

}

func assignAgentToRoom(payload *WebhookPayload) (uint, error) {
	var customer models.Customer
	database.DB.Where("status = ?", "waiting").Order("created_at").First(&customer)

	if customer.ID == 0 {
		return 0, fmt.Errorf("no customer in queue")
	}

	var selectedAgent models.Agent
	database.DB.Where("status = ? AND current_customers < max_customers", "online").First(&selectedAgent)

	if selectedAgent.ID == 0 {
		return 0, fmt.Errorf("no available agent")
	}

	err := hitAssignmentAPI(payload.RoomID, selectedAgent.ID)
	if err != nil {
		return 0, err
	}

	selectedAgent.CurrentCustomer++
	database.DB.Save(&selectedAgent)

	customer.Status = "served"
	database.DB.Save(&customer)

	return selectedAgent.ID, nil
}

func hitAssignmentAPI(roomID string, agentID uint) error {
	assigAgentURL := "https://multichannel.qiscus.com/api/v1/admin/service/assign_agent"

	data := url.Values{}
	data.Set("room_id", roomID)
	data.Set("agent_id", fmt.Sprintf("%d", agentID))
	data.Set("replace_latest_agent", "false")
	data.Set("max_agent", "1")

	req, err := http.NewRequest("POST", assigAgentURL, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Qiscus-App-Id", os.Getenv("QISCUS_APP_ID"))
	req.Header.Set("Qiscus-Secret-Key", os.Getenv("QISCUS_SECRET"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New("error assigning agent")
	}

	return nil
}
