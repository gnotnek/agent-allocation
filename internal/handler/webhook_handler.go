package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
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

type MarkAsResolvedPayload struct {
	Service struct {
		ID             int     `json:"id"`
		RoomID         string  `json:"room_id"`
		IsResolved     bool    `json:"is_resolved"`
		Notes          *string `json:"notes"` // Nullable field
		FirstCommentID string  `json:"first_comment_id"`
		LastCommentID  int     `json:"last_comment_id"`
		Source         string  `json:"source"`
	} `json:"service"`
	ResolvedBy struct {
		ID          int    `json:"id"`
		Email       string `json:"email"`
		Name        string `json:"name"`
		Type        string `json:"type"`
		IsAvailable bool   `json:"is_available"`
	} `json:"resolved_by"`
	Customer struct {
		UserID string `json:"user_id"`
	} `json:"customer"`
}

func HandleAllocateAgent(c *fiber.Ctx) error {
	payload := new(QiscusWebhookPayload)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	if payload.IsResolved {
		return c.JSON(fiber.Map{"message": "Conversation is already resolved."})
	}

	// Fetch available agents
	agents, err := getAvailableAgents(payload.RoomID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch available agents"})
	}

	maxCustomers, err := strconv.Atoi(os.Getenv("MAX_CUSTOMERS"))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to parse MAX_CUSTOMERS from environment"})
	}

	// Find an agent with less than maxCustomers
	var selectedAgent models.Agent
	for _, agent := range agents {
		if agent.CurrentCustomerCount < maxCustomers {
			selectedAgent = agent
			break
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
	err = assignAgent(payload.RoomID, selectedAgent.ID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to assign agent"})
	}

	fmt.Print("done")
	return c.JSON(fiber.Map{"message": "Agent assigned successfully", "agent_id": selectedAgent.ID})
}

func assignAgent(roomID string, agentID int) error {
	url := fmt.Sprintf("%s/api/v1/admin/service/assign_agent", os.Getenv("QISCUS_BASE_URL"))
	data := fmt.Sprintf("room_id=%s&agent_id=%d&replace_latest_agent=true", roomID, agentID)

	req, err := http.NewRequest("POST", url, strings.NewReader(data))
	if err != nil {
		return err
	}

	req.Header.Set("Qiscus-App-Id", os.Getenv("QISCUS_APP_ID"))
	req.Header.Set("Qiscus-Secret-Key", os.Getenv("QISCUS_SECRET_KEY"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to assign agent, status code: %d", resp.StatusCode)
	}

	// Increase the agent's active room count
	err = updateAgentRoomCount(agentID, 1)
	if err != nil {
		return fmt.Errorf("failed to update agent's room count: %v", err)
	}

	return nil
}

func getAvailableAgents(roomID string) ([]models.Agent, error) {
	url := fmt.Sprintf("%s/api/v2/admin/service/available_agents?room_id=%s", os.Getenv("QISCUS_BASE_URL"), roomID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Qiscus-App-Id", os.Getenv("QISCUS_APP_ID"))
	req.Header.Set("Qiscus-Secret-Key", os.Getenv("QISCUS_SECRET_KEY"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result models.QiscusResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return result.Data.Agents, nil
}

func HandlerMarkAsResolvedWebhook(c *fiber.Ctx) error {
	payload := new(MarkAsResolvedPayload)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	if !payload.Service.IsResolved {
		return c.Status(400).JSON(fiber.Map{"error": "Service is not resolved"})
	}

	// Decrease the agent's active room count using the ResolvedBy ID
	err := updateAgentRoomCount(payload.ResolvedBy.ID, -1)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update agent's room count"})
	}

	// Assign the next room in the queue
	err = assignNextRoomInQueue()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to assign the next room in queue"})
	}

	return c.JSON(fiber.Map{"message": "Service marked as resolved successfully"})
}

func updateAgentRoomCount(agentID int, delta int) error {
	agentRoomCount := make(map[int]int)

	// Update the agent's room count
	agentRoomCount[agentID] += delta

	if agentRoomCount[agentID] < 0 {
		agentRoomCount[agentID] = 0
	}

	return nil
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

func assignNextRoomInQueue() error {
	var queue models.RoomQueue
	result := database.DB.Order("position").First(&queue).Error
	if result != nil {
		return result
	}

	// Fetch available agents
	agents, err := getAvailableAgents(queue.RoomID)
	if err != nil {
		return err
	}

	maxCustomers, err := strconv.Atoi(os.Getenv("MAX_CUSTOMERS"))
	if err != nil {
		return err
	}
	var selectedAgent models.Agent
	for _, agent := range agents {
		if agent.CurrentCustomerCount < maxCustomers {
			selectedAgent = agent
			break
		}
	}

	if selectedAgent.ID == 0 {
		return nil
	}

	err = assignAgent(queue.RoomID, selectedAgent.ID)
	if err != nil {
		return err
	}

	return database.DB.Delete(&queue).Error
}
