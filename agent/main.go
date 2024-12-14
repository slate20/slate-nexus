package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"slate-rmm-agent/collectors"
	"slate-rmm-agent/logger"
	"slate-rmm-agent/server"
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
	APIKey    string `json:"api_key"`
}

func main() {
	err := logger.SetupLogger()
	if err != nil {
		log.Fatalf("could not setup logger: %v", err)
	}
	defer logger.LogFile.Close()

	logger.LogInfo("Starting SlateNexusAgent...")

	// Run as a service
	err = svc.Run("SlateNexusAgent", &Service{})
	if err != nil {
		logger.LogError("Service failed: %v", err)
	}
}

func runAgent(stop <-chan struct{}) {
	// Add a defer statement to recover from panics
	defer func() {
		if r := recover(); r != nil {
			logger.LogError("recovered from panic: %v", r)
		}
	}()

	logger.LogInfo("Agent starting...")

	config, err := loadConfig()
	if err != nil {
		logger.LogError("could not load config: %v", err)
		return
	}

	// If HostID is 0, run agentSetup and reload config
	if config.HostID == 0 {
		configFile := "C:\\Program Files\\SlateNexus\\config.json"
		err = agentSetup(config, configFile)
		if err != nil {
			logger.LogError("could not setup agent: %v", err)
			return
		}
		// Reload config after setup
		config, err = loadConfig()
		if err != nil {
			logger.LogError("could not reload config after setup: %v", err)
			return
		}
	}

	// Send a heartbeat every minute
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := server.Heartbeat(config.HostID, config.ServerURL, config.APIKey); err != nil {
				logger.LogError("could not send heartbeat: %v", err)
			} else {
				logger.LogInfo("Heartbeat sent successfully")
			}
		case <-stop:
			logger.LogInfo("Agent stopping...")
			return
		}
	}
}

func loadConfig() (Config, error) {
	var config Config

	// Define the path to the config file
	configFile := "C:\\Program Files\\SlateNexus\\config.json"

	// Read the config file
	data, err := os.ReadFile(configFile)
	if err != nil {
		logger.LogError("could not read config file: %v", err)
		return config, err
	}

	// Remove BOM
	data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))

	// Unmarshal the config file
	err = json.Unmarshal(data, &config)
	if err != nil {
		logger.LogError("could not unmarshal config file: %v", err)
	}

	return config, nil
}

func agentSetup(config Config, configPath string) error {

	// Collect data
	fmt.Println("Collecting system data...")
	data, err := collectors.CollectData()
	if err != nil {
		logger.LogError("could not collect data: %v", err)
		return err
	}

	// Register the agent and get the token for AUTOMATION_SECRET
	fmt.Println("Registering agent...")
	HostID, err := server.Register(data, config.ServerURL, config.APIKey)
	if err != nil {
		logger.LogError("could not register with the server: %v", err)
	} else {
		logger.LogInfo("Registered with the server with HostID: %d\n", HostID)
	}

	// Save the HostID in the config
	config.HostID = HostID

	// Save the config in the config file
	configBytes, err := json.Marshal(config)
	if err != nil {
		logger.LogError("could not marshal config: %v", err)
	}

	err = os.WriteFile(configPath, configBytes, 0644)
	if err != nil {
		logger.LogError("could not write config file: %v", err)
	}

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
