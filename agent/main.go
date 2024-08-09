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
		// If the config file does not exist, register the agent

		// Prompt the user for the server URL
		fmt.Print("Enter the server IP or Hostname (If DNS is configured): ")
		var serverURL string
		_, err := fmt.Scan(&serverURL)
		if err != nil {
			log.Printf("could not read server IP/Hostname: %v", err)
		}

		// Append "http://" and ":8080" to the server URL
		config.ServerURL = "http://" + serverURL + ":8080"

		// Validate the server URL
		_, err = url.ParseRequestURI(config.ServerURL)
		if err != nil {
			log.Printf("invalid server URL: %v", err)
			return
		}

		// Collect data
		fmt.Println("Collecting system data...")
		data, err := collectors.CollectData()
		if err != nil {
			log.Printf("could not collect data: %v", err)
		}

		// Download the CheckMK agent
		fmt.Println("Downloading CheckMK agent...")
		err = downloadFile("http://"+serverURL+":5000/main/check_mk/agents/windows/check_mk_agent.msi", "check_mk_agent.msi")
		if err != nil {
			log.Printf("could not download CheckMK agent: %v", err)
		}

		// Install the CheckMK agent
		fmt.Println("Installing CheckMK agent...")
		err = exec.Command("msiexec", "/i", "check_mk_agent.msi", "/qn").Run()
		if err != nil {
			log.Printf("could not install CheckMK agent: %v", err)
		}

		// Download CheckMK Inventory plugin
		err = downloadFile("http://"+serverURL+":5000/main/check_mk/agents/windows/plugins/mk_inventory.vbs", "mk_inventory.vbs")
		if err != nil {
			log.Printf("could not download CheckMK Inventory plugin: %v", err)
		}

		// Move the CheckMK Inventory plugin to the plugins directory
		err = os.Rename("mk_inventory.vbs", "C:\\ProgramData\\checkmk\\agent\\plugins\\mk_inventory.vbs")
		if err != nil {
			log.Printf("could not move CheckMK Inventory plugin: %v", err)
		}

		// Register the agent and get the token for AUTOMATION_SECRET
		fmt.Println("Registering agent...")
		HostID, token, err := server.Register(data, config.ServerURL)
		if err != nil {
			log.Printf("could not register with the server: %v", err)
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

		// Register the CheckMK agent with the CheckMK server
		fmt.Println("Registering CheckMK agent...")
		cmd := exec.Command("C:\\Program Files (x86)\\checkmk\\service\\cmk-agent-ctl.exe", "register", "--hostname", data.Hostname, "--server", serverURL+":8000", "--site", "main", "--user", "cmkadmin", "--password", result.Secret)

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err = cmd.Run()
		if err != nil {
			log.Printf("could not register CheckMK agent: %v", err)
			log.Printf("stdout: %s", stdout.String())
			log.Printf("stderr: %s", stderr.String())
		}

		// Delete the AUTOMATION_SECRET
		result.Secret = ""

		// Sleep for 1 minute to allow the agent to register with the CheckMK server
		fmt.Println("Waiting to run service discovery (1m)...")
		time.Sleep(1 * time.Minute)

		// Send service discovery request to the server
		fmt.Println("Sending service discovery request...")
		// Prepare the request body
		body = map[string]string{
			"host_name": data.Hostname,
		}
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			log.Printf("could not marshal request body: %v", err)
		}

		// Create the request
		req, err := http.NewRequest("POST", config.ServerURL+"/api/agents/cmksvcd", bytes.NewBuffer(bodyBytes))
		if err != nil {
			log.Printf("could not create request: %v", err)
		}

		// Set the content type to application/json
		req.Header.Set("Content-Type", "application/json")

		//  Print the request
		//log.Printf("Request: %v", req)

		// Send the request
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("could not send request: %v", err)
		}
		defer resp.Body.Close()
		// Log the response
		bodyRespBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("could not read response body: %v", err)
		}
		bodyResp := string(bodyRespBytes)
		log.Printf("Response: %v", bodyResp)
		log.Println("Service discovery request sent")

		// Download the Remotely Windows installer
		fmt.Println("Downloading Remotely agent...")
		err = downloadFile("http://"+serverURL+":8080/download/remotely-win", "Install-Remotely.ps1")
		if err != nil {
			log.Printf("could not download Remotely agent: %v", err)
		}

		// Install the Remotely agent
		fmt.Println("Installing Remotely agent...")

		// Prepare the command with parameters
		cmd = exec.Command("powershell", "-ExecutionPolicy", "Bypass", "-File", "Install-Remotely.ps1", "-install")

		// Capture the output
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("could not install Remotely agent: %v\nOutput: %s", err, output)
			return
		}

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

	// Pause until the user presses a key
	// fmt.Println("Press any key to exit...")
	// var input string
	// fmt.Scanln(&input)

	// Send a heartbeat every minute
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		if err := server.Heartbeat(config.HostID, config.ServerURL); err != nil {
			log.Printf("could not send heartbeat: %v", err)
		}
	}

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
