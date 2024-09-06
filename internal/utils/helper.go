package helper

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/gnotnek/agent-allocation/internal/database"
	"github.com/gnotnek/agent-allocation/internal/models"
	"gorm.io/gorm"
)

// AssignAgent assigns an agent to a room
func AssignAgent(roomID string, agentID int) error {
	url := fmt.Sprintf("%s/api/v1/admin/service/assign_agent", os.Getenv("QISCUS_BASE_URL"))
	data := fmt.Sprintf("room_id=%s&agent_id=%d", roomID, agentID)

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

	return nil
}

// GetAvailableAgents fetches available agents for a specific room
func GetAvailableAgents(roomID string) ([]models.Agent, error) {
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

func AddRoomToQueue(roomID string) error {
	var count int64
	database.DB.Model(&models.RoomQueue{}).Where("room_id = ?", roomID).Count(&count)
	if count == 0 {
		// Add new room to queue
		roomQueue := models.RoomQueue{
			RoomID: roomID,
		}
		return database.DB.Create(&roomQueue).Error
	}
	return nil
}

func IsRoomAlreadyExists(roomID string) (bool, error) {
	var queue models.RoomQueue
	err := database.DB.Where("room_id = ?", roomID).First(&queue).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	return err == nil, err
}

func AssignAgentToRoom(roomID string, agents []models.Agent) error {
	maxCustomers, err := strconv.Atoi(os.Getenv("MAX_CUSTOMERS"))
	if err != nil {
		return err
	}

	// Filter eligible agents
	var eligibleAgents []models.Agent
	for _, agent := range agents {
		if agent.IsAvailable && agent.CurrentCustomerCount < maxCustomers {
			eligibleAgents = append(eligibleAgents, agent)
		}
	}

	if len(eligibleAgents) == 0 {
		return nil // No available agents, leave room unassigned
	}

	// Sort agents by least number of customers
	sort.Slice(eligibleAgents, func(i, j int) bool {
		return eligibleAgents[i].CurrentCustomerCount < eligibleAgents[j].CurrentCustomerCount
	})

	selectedAgent := eligibleAgents[0]
	err = AssignAgent(roomID, selectedAgent.ID)
	if err != nil {
		return err
	}

	// Assign the room to the agent in the database
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		err = tx.Model(&models.RoomQueue{}).Where("room_id = ?", roomID).Update("agent_id", selectedAgent.ID).Error
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func ResolveRoom(roomID string) error {
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		// Remove the room from the queue
		err := tx.Where("room_id = ?", roomID).Delete(&models.RoomQueue{}).Error
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func AssignNextRoomFromQueue() error {
	// Fetch the oldest unassigned room
	var queue models.RoomQueue
	err := database.DB.Where("agent_id IS NULL").Order("created_at").First(&queue).Error
	if err != nil {
		return nil // No unassigned rooms
	}

	// Get available agents
	agents, err := GetAvailableAgents(queue.RoomID)
	if err != nil {
		return err
	}

	// Assign the room to the agent
	return AssignAgentToRoom(queue.RoomID, agents)
}
