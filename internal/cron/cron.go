package cron

import (
	"github.com/gnotnek/agent-allocation/internal/database"
	"github.com/gnotnek/agent-allocation/internal/helper"
	"github.com/gnotnek/agent-allocation/internal/models"
)

func CronAssignAgent() error {
	var RoomQueues []models.RoomQueue
	//read from database yang id agen null
	err := database.DB.Where("agent_id IS NULL").Find(&RoomQueues).Error
	if err != nil {
		return err
	}

	for _, queue := range RoomQueues {
		// Get available agents
		agents, err := helper.GetAvailableAgents(queue.RoomID)
		if err != nil {
			return err
		}

		// Select agent with least customers and assign
		err = helper.AssignAgentToRoom(queue.RoomID, agents)
		if err != nil {
			return err
		}
	}

	return nil
}
