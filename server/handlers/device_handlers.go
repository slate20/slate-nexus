package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"slate-rmm/database"
	"slate-rmm/models"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

// Handler to get all devices
func GetDevices(w http.ResponseWriter, r *http.Request) {
	// Call the GetAllAgents function from the database package
	agents, err := database.GetAllAgents()
	if err != nil {
		http.Error(w, "Failed to fetch devices", http.StatusInternalServerError)
		log.Println("Failed to fetch devices:", err)
		return
	}

	// Render the template
	RenderTemplate(w, "device-list.html", agents)
}

func GetRemoteControlURL(w http.ResponseWriter, r *http.Request) {
	//load .env file
	err := godotenv.Load()
	if err != nil {
		http.Error(w, "could not load .env file", http.StatusInternalServerError)
		return
	}

	apiURL := os.Getenv("REMOTELY_API_URL")
	apiToken := os.Getenv("REMOTELY_API_TOKEN")
	apiID := os.Getenv("REMOTELY_API_ID")

	// get the RemotelyID using the host ID from the end of request URL
	vars := mux.Vars(r)
	hostID := vars["id"]
	if _, err := strconv.Atoi(hostID); err != nil || hostID == "" {
		http.Error(w, "invalid host ID", http.StatusBadRequest)
		return
	}

	agent, err := database.GetAgent(hostID)
	if err != nil {
		http.Error(w, "could not get agent", http.StatusInternalServerError)
		return
	}

	// Validate RemotelyID
	if agent.RemotelyID == "" || !regexp.MustCompile(`[a-zA-Z0-9]+$`).MatchString(agent.RemotelyID) {
		http.Error(w, "invalid RemotelyID", http.StatusBadRequest)
		return
	}

	baseURL, err := url.Parse(apiURL)
	if err != nil {
		http.Error(w, "invalid API URL", http.StatusBadRequest)
		return
	}

	// build the Remotely API URL
	baseURL.Path = path.Join(baseURL.Path, "RemoteControl", agent.RemotelyID)
	url := baseURL.String()

	// build the Remotely API request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		http.Error(w, "could not create request", http.StatusInternalServerError)
		return
	}

	// set the headers
	req.Header.Set("X-Api-Key", apiID+":"+apiToken)

	// send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "could not send request", http.StatusInternalServerError)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Print the response to the console
	fmt.Println(resp)

	// Read the response body
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "could not read response body", http.StatusInternalServerError)
		return
	}

	// Set HTMX headers
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("HX-Trigger", "remoterequest")

	// Return the URL in a JSON object

}

// GetCommands returns a list of commands for a given agent
func GetCommands(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement get commands
}

// QueueCommand queues a command to be executed on an agent
func QueueCommand(w http.ResponseWriter, r *http.Request) {
	//Get the agent ID from the request url
	vars := mux.Vars(r)
	hostID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid host ID", http.StatusBadRequest)
		return
	}

	// Parse the form data and get the command string
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	command := r.FormValue("command")

	// Create the command object
	cmd := &models.AgentCommand{
		HostID:    hostID,
		Command:   command,
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save the command to the database
	err = database.CreateCommand(cmd)
	if err != nil {
		http.Error(w, "Failed to save command", http.StatusInternalServerError)
		return
	}

	// Send a succes response for now
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Command queued for agent %d", hostID)
}

// GetCommandResults returns the status of a command
func GetCommandResults(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement get command results
}
