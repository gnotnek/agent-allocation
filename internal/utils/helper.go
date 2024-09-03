package helper

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gnotnek/agent-allocation/internal/database"
	"github.com/gnotnek/agent-allocation/internal/models"
)

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

func UpdateAgentRoomCount(agentID int, delta int) error {
	agentRoomCount := make(map[int]int)

	// Update the agent's room count
	agentRoomCount[agentID] += delta

	if agentRoomCount[agentID] < 0 {
		agentRoomCount[agentID] = 0
	}

	return nil
}

func AssignNextRoomInQueue() error {
	var queue models.RoomQueue
	result := database.DB.Order("position").First(&queue).Error
	if result != nil {
		return result
	}

	// Fetch available agents
	agents, err := GetAvailableAgents(queue.RoomID)
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

	err = AssignAgent(queue.RoomID, selectedAgent.ID)
	if err != nil {
		return err
	}

	return database.DB.Delete(&queue).Error
}

func AssignAgent(roomID string, agentID int) error {
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
	err = UpdateAgentRoomCount(agentID, 1)
	if err != nil {
		return fmt.Errorf("failed to update agent's room count: %v", err)
	}

	return nil
}
