package models

import "time"

// AgentCommand represents a command to be executed on an agent
type AgentCommand struct {
	ID        uint      `json:"ID"`
	HostID    int       `json:"HostID"`
	Command   string    `json:"Command"`
	Status    string    `json:"Status"`
	Output    string    `json:"Output"`
	CreatedAt time.Time `json:"CreatedAt"`
	UpdatedAt time.Time `json:"UpdatedAt"`
}
