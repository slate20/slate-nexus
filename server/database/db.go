package database

import (
	"errors"
	"slate-rmm/models"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

// InitDB initializes the database connection
func InitDB(dataSourceName string) error {
	var err error
	db, err = gorm.Open(postgres.Open(dataSourceName), &gorm.Config{})
	if err != nil {
		return err
	}

	return nil
}

// --- Agent functions ---
// RegisterNewAgent stores a new agent in the database
func RegisterNewAgent(agent *models.Agent) error {
	return db.Create(agent).Error
}

// GetAllAgents returns all the agents in the database
func GetAllAgents() ([]models.Agent, error) {
	var agents []models.Agent
	err := db.Find(&agents).Error
	return agents, err
}

// GetAgent returns a single agent from the database
func GetAgent(id string) (*models.Agent, error) {
	var agent models.Agent
	err := db.First(&agent, id).Error
	return &agent, err
}

// UpdateAgent updates an agent in the database
func UpdateAgent(agent *models.Agent) error {
	return db.Save(agent).Error
}

// DeleteAgent deletes an agent from the database
func DeleteAgent(id string) error {
	return db.Delete(&models.Agent{}, id).Error
}

// AgentHeartbeat updates the last_seen field of an agent
func AgentHeartbeat(id string) error {
	return db.Model(&models.Agent{}).Where("host_id = ?", id).Update("last_seen", time.Now()).Error
}

// --- Group functions ---
// CreateGroup creates a new group in the database
func CreateGroup(groupName string) error {
	return db.Create(&models.Group{GroupName: groupName}).Error
}

// GetAllGroups returns all the groups in the database
func GetAllGroups() ([]models.Group, error) {
	var groups []models.Group
	err := db.Find(&groups).Error
	return groups, err
}

// GetGroup returns a single group from the database
func GetGroup(id string) (*models.Group, error) {
	var group models.Group
	err := db.First(&group, id).Error
	return &group, err
}

// UpdateGroup updates a group in the database
func UpdateGroup(id string, groupName string) error {
	return db.Model(&models.Group{}).Where("group_id = ?", id).Update("group_name", groupName).Error
}

// DeleteGroup deletes a group from the database
func DeleteGroup(id string) error {
	return db.Delete(&models.Group{}, id).Error
}

// GetHostsInGroup returns all the hosts in a group
func GetHostsInGroup(groupID int) ([]models.Agent, error) {
	var agents []models.Agent
	err := db.Joins("JOIN device_group_members ON agents.host_id = device_group_members.host_id").Where("device_group_members.group_id = ?", groupID).Find(&agents).Error
	return agents, err
}

// AddHostToGroup adds a host to a group
func AddHostToGroup(hostID, groupID int) error {
	return db.Create(&models.DeviceGroupMember{HostID: uint(hostID), GroupID: uint(groupID)}).Error
}

// RemoveHostFromGroup removes a host from a group
func RemoveHostFromGroup(hostID, groupID int) error {
	return db.Where("host_id = ? AND group_id = ?", hostID, groupID).Delete(&models.DeviceGroupMember{}).Error
}

// MoveHostToGroup moves a host from one group to another
func MoveHostToGroup(hostID, newGroupID int) error {
	// First, remove the host from its current group(s)
	err := db.Where("host_id = ?", hostID).Delete(&models.DeviceGroupMember{}).Error
	if err != nil {
		return err
	}
	// Then, add the host to the new group
	return db.Create(&models.DeviceGroupMember{HostID: uint(hostID), GroupID: uint(newGroupID)}).Error
}

// --- Command functions ---
// CreateCommand creates a new command in the database
func CreateCommand(command *models.AgentCommand) error {
	return db.Create(command).Error
}

// GetPendingCommand returns the oldest pending command for an agent (FIFO)
func GetPendingCommand(hostID int) (*models.AgentCommand, error) {
	var command models.AgentCommand
	err := db.Where("host_id = ? AND status = 'pending'", hostID).Order("created_at ASC").First(&command).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &command, nil
}

// UpdateCommand updates the status and output of a command in the database
func UpdateCommand(command *models.AgentCommand) error {
	return db.Save(command).Error
}
