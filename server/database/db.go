package database

import (
	"database/sql"
	"encoding/json"
	"slate-rmm/models"
	"time"

	_ "github.com/lib/pq"
)

var db *sql.DB

// InitDB initializes the database connection
func InitDB(dataSourceName string) error {
	var err error
	db, err = sql.Open("postgres", dataSourceName)
	if err != nil {
		return err
	}

	if err = db.Ping(); err != nil {
		return err
	}

	return nil
}

// RegisterNewAgent stores a new agent in the database
func RegisterNewAgent(agent *models.Agent) error {

	// Marshal the hardware specs
	hardwareSpecsJSON, err := json.Marshal(agent.HardwareSpecs)
	if err != nil {
		return err
	}

	// Prepare for SQL Statement
	stmt, err := db.Prepare(`
		INSERT INTO agents (hostname, ip_address, os, os_version, hardware_specs, agent_version, last_seen, last_user)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING host_id
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Execute the SQL Statement. For 'last_seen', we use the current time
	err = stmt.QueryRow(
		agent.Hostname,
		agent.IPAddress,
		agent.OS,
		agent.OSVersion,
		hardwareSpecsJSON,
		agent.AgentVersion,
		time.Now(),
		agent.LastUser,
	).Scan(&agent.ID)

	if err != nil {
		return err
	}

	return nil
}

// GetAllAgents returns all the agents in the database
func GetAllAgents() ([]models.Agent, error) {
	rows, err := db.Query("SELECT * FROM agents")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate over the rows and add the agents to the slice
	var agents []models.Agent
	for rows.Next() {
		var agent models.Agent
		var hardwareSpecsRaw sql.RawBytes
		if err := rows.Scan(&agent.ID, &agent.Hostname, &agent.IPAddress, &agent.OS, &agent.OSVersion, &hardwareSpecsRaw, &agent.AgentVersion, &agent.LastSeen, &agent.LastUser); err != nil {
			return nil, err
		}

		// Unmarshal the hardware specs
		if len(hardwareSpecsRaw) > 0 {
			if err := json.Unmarshal([]byte(hardwareSpecsRaw), &agent.HardwareSpecs); err != nil {
				return nil, err
			}
		}

		agents = append(agents, agent)
	}

	return agents, nil
}

// GetAgent returns a single agent from the database
func GetAgent(id string) (*models.Agent, error) {
	row := db.QueryRow("SELECT * FROM agents WHERE host_id = $1", id)

	// Scan the row into an Agent struct
	var agent models.Agent
	var hardwareSpecsRaw []byte
	if err := row.Scan(&agent.ID, &agent.Hostname, &agent.IPAddress, &agent.OS, &agent.OSVersion, &hardwareSpecsRaw, &agent.AgentVersion, &agent.LastSeen); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Unmarshal the hardware specs
	if len(hardwareSpecsRaw) > 0 {
		if err := json.Unmarshal([]byte(hardwareSpecsRaw), &agent.HardwareSpecs); err != nil {
			return nil, err
		}
	}

	return &agent, nil
}

// UpdateAgent updates an agent in the database
func UpdateAgent(id string, agent *models.Agent) error {
	_, err := db.Exec("UPDATE agents SET hostname = $1, ip_address = $2, os = $3, os_version = $4, agent_version = $5, last_seen = $6, last_user = $7 WHERE host_id = $8",
		agent.Hostname, agent.IPAddress, agent.OS, agent.OSVersion, agent.AgentVersion, time.Now(), agent.LastUser, id)
	return err
}

// DeleteAgent deletes an agent from the database
func DeleteAgent(id string) error {
	_, err := db.Exec("DELETE FROM agents WHERE host_id = $1", id)
	return err
}

// AgentHeartbeat updates the last_seen field of an agent
func AgentHeartbeat(id string) error {
	_, err := db.Exec("UPDATE agents SET last_seen = $1 WHERE host_id = $2", time.Now(), id)
	return err
}

// CreateGroup creates a new group in the database
func CreateGroup(groupName string) error {
	_, err := db.Exec("INSERT INTO device_groups (group_name) VALUES ($1)", groupName)
	return err
}

// GetAllGroups returns all the groups in the database
func GetAllGroups() ([]models.Group, error) {
	rows, err := db.Query("SELECT group_id, group_name FROM device_groups")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []models.Group
	for rows.Next() {
		var group models.Group
		if err := rows.Scan(&group.GroupID, &group.GroupName); err != nil {
			return nil, err
		}
		groups = append(groups, group)
	}

	return groups, nil
}

// GetGroup returns a single group from the database
func GetGroup(id string) (string, error) {
	row := db.QueryRow("SELECT group_name FROM device_groups WHERE group_id = $1", id)

	var group string
	if err := row.Scan(&group); err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}

	return group, nil
}

// UpdateGroup updates a group in the database
func UpdateGroup(id string, groupName string) error {
	_, err := db.Exec("UPDATE device_groups SET group_name = $1 WHERE group_id = $2", groupName, id)
	return err
}

// DeleteGroup deletes a group from the database
func DeleteGroup(id string) error {
	_, err := db.Exec("DELETE FROM device_groups WHERE group_id = $1", id)
	return err
}

// GetHostsInGroup returns all the hosts in a group
func GetHostsInGroup(groupID int) ([]models.Agent, error) {
	rows, err := db.Query("SELECT a.* FROM agents a JOIN device_group_members dgm ON a.host_id = dgm.host_id WHERE dgm.group_id = $1", groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var agents []models.Agent
	for rows.Next() {
		var agent models.Agent
		var hardwareSpecsRaw sql.RawBytes
		if err := rows.Scan(&agent.ID, &agent.Hostname, &agent.IPAddress, &agent.OS, &agent.OSVersion, &hardwareSpecsRaw, &agent.AgentVersion, &agent.LastSeen, &agent.LastUser); err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}
			return nil, err
		}
		agents = append(agents, agent)
	}

	return agents, nil
}

// AddhostToGroup adds a host to a group
func AddHostToGroup(hostID, groupID int) error {
	_, err := db.Exec("INSERT INTO device_group_members (host_id, group_id) VALUES ($1, $2)", hostID, groupID)
	return err
}

// RemoveHostFromGroup removes a host from a group
func RemoveHostFromGroup(hostID, groupID int) error {
	_, err := db.Exec("DELETE FROM device_group_members WHERE host_id = $1 AND group_id = $2", hostID, groupID)
	return err
}

// MoveHostToGroup moves a host from one group to another
func MoveHostToGroup(hostID, newGroupID int) error {
	_, err := db.Exec("UPDATE device_group_members SET group_id = $1 WHERE host_id = $2", newGroupID, hostID)
	return err
}
