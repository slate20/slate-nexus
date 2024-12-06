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
	"strings"
	"time"

	"golang.org/x/sys/windows/svc"
)

type Service struct{}

func (s *Service) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
	changes <- svc.Status{State: svc.StartPending}
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

	stop := make(chan struct{})

	go runAgent(stop)

	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				close(stop)
				changes <- svc.Status{State: svc.StopPending}
				return
			default:
				log.Printf("unexpected control request: #%d", c)
			}
		}
	}
}

// Config represents the configuration for the agent
type Config struct {
	ServerURL string `json:"server_url"`
	HostID    int32  `json:"host_id"`
}

func main() {
	err := logger.setupLogger()
	if err != nil {
		log.Fatalf("could not setup logger: %v", err)
	}
	defer logFile.Close()

	logInfo("Starting SlateNexusAgent...")

	// Run as a service
	err = svc.Run("SlateRMMAgent", &Service{})
	if err != nil {
		logError("Service failed: %v", err)
	}
}

func runAgent(stop <-chan struct{}) {
	// Add a defer statement to recover from panics
	defer func() {
		if r := recover(); r != nil {
			logError("recovered from panic: %v", r)
		}
	}()

	logInfo("Agent starting...")

	config, err := loadConfig()
	if err != nil {
		logError("could not load config: %v", err)
		return
	}

	// Send a heartbeat every minute
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := server.Heartbeat(config.HostID, config.ServerURL); err != nil {
				logError("could not send heartbeat: %v", err)
			} else {
				logInfo("Heartbeat sent successfully")
			}
		case <-stop:
			logInfo("Agent stopping...")
			return
		}
	}
}

func loadConfig() (Config, error) {
	var config Config

	//Get the directory of the executable
	exe, err := os.Executable()
	if err != nil {
		logError("could not get the directory of the executable: %v", err)
		return config, fmt.Errorf("could not get the directory of the executable: %w", err)
	}
	dir := filepath.Dir(exe)

	// Define the path to the config file
	configPath := filepath.Join(dir, "config.json")

	// Check if the config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// If the config file does not exist, setup the agent
		err = agentSetup(config, configPath)
		if err != nil {
			logError("could not setup the agent: %v", err)
			return config, err
		}
	}

	// If the config file exists, read the config from the file
	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		logError("could not read config file: %v", err)
		return config, fmt.Errorf("could not read config file: %w", err)
	}

	err = json.Unmarshal(configBytes, &config)
	if err != nil {
		logError("could not unmarshal config: %v", err)
		return config, fmt.Errorf("could not unmarshal config: %v", err)
	}

	return config, nil
}

func agentSetup(config Config, configPath string) error {
	// Check for tmp server url file
	tempURLPath := strings.ReplaceAll(configPath, "config.json", "server_url.tmp")
	if _, err := os.Stat(tempURLPath); err == nil {
		// Read the server URL from the file
		serverURL, err := os.ReadFile(tempURLPath)
		if err != nil {
			logError("could not read server URL file: %v", err)
			return fmt.Errorf("could not read server URL file: %w", err)
		}
		config.ServerURL = strings.TrimSpace(string(serverURL))

		// Delete the file
		os.Remove(tempURLPath)
	} else {
		// If the file does not exist, prompt the user
		fmt.Print("Enter the server IP or Hostname (if DNS is configured): ")
		_, err := fmt.Scanln(&config.ServerURL)
		if err != nil {
			logError("could not read server URL: %v", err)
			return fmt.Errorf("could not read server URL: %w", err)
		}
	}

	// Append "http://" to the server URL
	config.ServerURL = "http://" + config.ServerURL

	// Validate the server URL
	_, err := url.ParseRequestURI(config.ServerURL)
	if err != nil {
		logError("invalid server URL: %v", err)
		return err
	}

	// Install Remotely agent
	err = installRemotely(config)
	if err != nil {
		logError("could not install Remotely agent: %v", err)
		return err
	}

	// Collect data
	fmt.Println("Collecting system data...")
	data, err := collectors.CollectData()
	if err != nil {
		logError("could not collect data: %v", err)
		return err
	}

	// Register the agent and get the token for AUTOMATION_SECRET
	fmt.Println("Registering agent...")
	HostID, token, err := server.Register(data, config.ServerURL)
	if err != nil {
		logError("could not register with the server: %v", err)
	} else {
		logInfo("Registered with the server with HostID: %d\n", HostID)
	}

	// Save the HostID in the config
	config.HostID = HostID

	// Save the config in the config file
	configBytes, err := json.Marshal(config)
	if err != nil {
		logError("could not marshal config: %v", err)
	}

	err = os.WriteFile(configPath, configBytes, 0644)
	if err != nil {
		logError("could not write config file: %v", err)
	}

	// Use the token to get the AUTOMATION_SECRET
	url := config.ServerURL + "/api/agents/secret"

	body := map[string]string{
		"token":    token,
		"agent_id": fmt.Sprint(HostID),
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		logError("could not encode request body: %v", err)
	}
	// log.Printf("Sending request: %s", string(jsonBody))
	// deepcode ignore Ssrf: Validation performed after user input
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		logError("could not send request: %v", err)
	} else {
		defer resp.Body.Close()
	}

	// Decode the response to get the AUTOMATION_SECRET
	var result struct {
		Secret string `json:"secret"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		logError("could not decode response: %v", err)
	}

	// Delete the AUTOMATION_SECRET
	result.Secret = ""

	return nil
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
		logInfo("Remotely is already installed")
	} else {
		logInfo("Downloading Remotely agent...")
		err = downloadFile(config.ServerURL+"/api/download/remotely-win", "Install-Remotely.ps1")
		if err != nil {
			logError("could not download Remotely agent: %v", err)
		}

		// Install the Remotely agent
		logInfo("Installing Remotely agent...")

		// Prepare the command with parameters
		cmd := exec.Command("powershell", "-ExecutionPolicy", "Bypass", "-File", "Install-Remotely.ps1", "-install")

		// Capture the output
		output, err := cmd.CombinedOutput()
		if err != nil {
			logError("could not install Remotely agent: %v\nOutput: %s", err, output)
		}
	}
	return nil
}
