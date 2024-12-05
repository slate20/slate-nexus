package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"slate-rmm-agent/collectors"
	"slate-rmm-agent/server"
	"time"
)

// Config represents the configuration for the agent
type Config struct {
	ServerURL string `json:"server_url"`
	HostID    int32  `json:"host_id"`
}

func main() {
	// Add a defer statement to recover from panics
	defer func() {
		if r := recover(); r != nil {
			log.Printf("recovered from panic: %v", r)
		}
	}()

	var config Config
	var err error

	//Get the directory of the executable
	exe, err := os.Executable()
	if err != nil {
		log.Printf("could not get the directory of the executable: %v", err)
	}
	dir := filepath.Dir(exe)

	// Define the path to the config file
	configPath := filepath.Join(dir, "config.json")

	// Check if the config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// If the config file does not exist, setup the agent
		agentSetup(config, configPath)
	} else {
		// If the config file exists, read the config from the file
		configBytes, err := os.ReadFile(configPath)
		if err != nil {
			log.Printf("could not read config file: %v", err)
		}

		err = json.Unmarshal(configBytes, &config)
		if err != nil {
			log.Printf("could not unmarshal config: %v", err)
		}
	}

	// Send a heartbeat every minute
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		if err := server.Heartbeat(config.HostID, config.ServerURL); err != nil {
			log.Printf("host id: %d, server url: %s", config.HostID, config.ServerURL)
			log.Printf("could not send heartbeat: %v", err)
		}
	}

}

func agentSetup(config Config, configPath string) {
	// Prompt the user for the server URL
	fmt.Print("Enter the server IP or Hostname (If DNS is configured): ")
	var serverURL string
	_, err := fmt.Scan(&serverURL)
	if err != nil {
		log.Printf("could not read server IP/Hostname: %v", err)
	}

	// Append "http://" to the server URL
	config.ServerURL = "http://" + serverURL

	// Validate the server URL
	_, err = url.ParseRequestURI(config.ServerURL)
	if err != nil {
		log.Printf("invalid server URL: %v", err)
		return
	}

	// Install Remotely agent
	err = installRemotely(config)
	if err != nil {
		log.Printf("could not install Remotely agent: %v", err)
	}

	// Collect data
	fmt.Println("Collecting system data...")
	data, err := collectors.CollectData()
	if err != nil {
		log.Printf("could not collect data: %v", err)
	}

	// Register the agent and get the token for AUTOMATION_SECRET
	fmt.Println("Registering agent...")
	HostID, token, err := server.Register(data, config.ServerURL)
	if err != nil {
		log.Printf("could not register with the server: %v", err)
	} else {
		fmt.Printf("Registered with the server with HostID: %d\n", HostID)
	}

	// Save the HostID in the config
	config.HostID = HostID

	// Save the config in the config file
	configBytes, err := json.Marshal(config)
	if err != nil {
		log.Printf("could not marshal config: %v", err)
	}

	err = os.WriteFile(configPath, configBytes, 0644)
	if err != nil {
		log.Printf("could not write config file: %v", err)
	}

	// Use the token to get the AUTOMATION_SECRET
	url := config.ServerURL + "/api/agents/secret"

	body := map[string]string{
		"token":    token,
		"agent_id": fmt.Sprint(HostID),
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		log.Printf("could not encode request body: %v", err)
	}
	// log.Printf("Sending request: %s", string(jsonBody))
	// deepcode ignore Ssrf: Validation performed after user input
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Printf("could not send request: %v", err)
	} else {
		defer resp.Body.Close()
	}

	// Decode the response to get the AUTOMATION_SECRET
	var result struct {
		Secret string `json:"secret"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("could not decode response: %v", err)
	}

	// Delete the AUTOMATION_SECRET
	result.Secret = ""

}

// downloadFile downloads a file from the given URL and saves it to the given path
func downloadFile(url string, path string) error {
	// Create the file
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	// deepcode ignore Ssrf: Validation performed after user input
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the data to the file
	_, err = io.Copy(out, resp.Body)
	return err
}

func installRemotely(config Config) error {
	// Download the Remotely Windows installer if Remotely is not installed

	//Check if "C:\Program Files\Remotely\Remotely_Agent.exe" exists
	_, err := os.Stat("C:\\Program Files\\Remotely\\Remotely_Agent.exe")
	if err == nil {
		fmt.Println("Remotely is already installed")
	} else {
		fmt.Println("Downloading Remotely agent...")
		err = downloadFile(config.ServerURL+"/api/download/remotely-win", "Install-Remotely.ps1")
		if err != nil {
			log.Printf("could not download Remotely agent: %v", err)
		}

		// Install the Remotely agent
		fmt.Println("Installing Remotely agent...")

		// Prepare the command with parameters
		cmd := exec.Command("powershell", "-ExecutionPolicy", "Bypass", "-File", "Install-Remotely.ps1", "-install")

		// Capture the output
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("could not install Remotely agent: %v\nOutput: %s", err, output)
		}
	}
	return nil
}
