package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

type Hardware struct {
	CPU     string `json:"cpu"`
	Memory  string `json:"memory"`
	Storage string `json:"storage"`
}

func (h Hardware) Value() (driver.Value, error) {
	return json.Marshal(h)
}

func (h *Hardware) Scan(value any) error {
	return json.Unmarshal(value.([]byte), h)
}

// Agent represents a system that the RMM tool will monitor.
type Agent struct {
	ID            int32     `gorm:"primaryKey;column:host_id" json:"host_id"`
	Hostname      string    `json:"hostname"`
	IPAddress     string    `json:"ip_address"`
	OS            string    `json:"os"`
	OSVersion     string    `json:"os_version"`
	HardwareSpecs Hardware  `gorm:"type:jsonb" json:"hardware_specs"`
	AgentVersion  string    `json:"agent_version"`
	LastSeen      time.Time `json:"last_seen"`
	LastUser      string    `json:"last_user"`
	RemotelyID    string    `json:"remotely_id"`
}

func (Agent) TableName() string {
	return "agents"
}

// Group represents a group of agents
type Group struct {
	GroupID   int32   `gorm:"primaryKey;column:group_id" json:"group_id"`
	GroupName string  `gorm:"unique;not null" json:"group_name"`
	Agents    []Agent `gorm:"many2many:device_group_members;foreignKey:GroupID;references:ID;joinForeignKey:group_id;joinReferences:host_id"`
}

func (Group) TableName() string {
	return "device_groups"
}

// DeviceGroupMember represents the join table for agents and groups
type DeviceGroupMember struct {
	HostID  uint `gorm:"primaryKey"`
	GroupID uint `gorm:"primaryKey"`
}

func (DeviceGroupMember) TableName() string {
	return "device_group_members"
}

// AgentCommand represents a command to be executed on an agent.
type AgentCommand struct {
	ID        uint   `gorm:"primaryKey"`
	HostID    int    `gorm:"not null"`
	Agent     Agent  `gorm:"foreignKey:HostID"`
	Command   string `gorm:"not null"`
	Status    string `gorm:"not null"`
	Output    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (AgentCommand) TableName() string {
	return "agent_commands"
}
