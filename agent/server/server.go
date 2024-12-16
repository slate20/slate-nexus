package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slate-rmm-agent/collectors"
	"slate-rmm-agent/logger"
)

// Register sends a POST request to the server to register the agent
func Register(data collectors.AgentData, ServerURL string, apiKey string) (int32, error) {
	url := ServerURL + "/agents/register"
	// Convert data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return 0, err
	}

	// Send a POST request to the AgentRegister endpoint
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}

	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.LogError("error reading response body: %v", err)
		return 0, err
	}

	// Decode the response and get the host ID
	var result struct {
		HostID int32 `json:"host_id"`
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		logger.LogError("Error decoding response: %v", err)
		return 0, err
	}

	// Check the response status code
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		logger.LogError("Unexpected status code: %d", resp.StatusCode)
		return 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return result.HostID, nil
}

// Heartbeat sends a POST request to the server with updated agent information
func Heartbeat(hostID int32, ServerURL string, apiKey string) error {
	// Collect agent data
	data, err := collectors.CollectData()
	if err != nil {
		return err
	}

	url := ServerURL + "/agents/" + fmt.Sprint(hostID)

	// Convert data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Debug: Print the JSON data
	// fmt.Printf("Sending JSON data: %s\n", jsonData)

	// Send a PUT request to the AgentHeartbeat endpoint
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	// Set the headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %s", resp.Status)
	} else {
		fmt.Println("Successful Heartbeat")
	}

	return nil
}
