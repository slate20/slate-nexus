package api_handlers

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"os"
	"os/exec"
	"slate-rmm/database"
	"slate-rmm/livestatus"
	"slate-rmm/models"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

var agentTokens = make(map[string]string)

// AgentRegistration handles the registration of a new agent
func AgentRegistration(w http.ResponseWriter, r *http.Request) {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		http.Error(w, "could not load .env file", http.StatusInternalServerError)
		return
	}

	apiURL := os.Getenv("API_URL")
	if apiURL == "" {
		log.Fatal("API_URL is not set in .env file")
	}
	siteName := os.Getenv("SITE_NAME")
	if siteName == "" {
		log.Fatal("SITE_NAME is not set in .env file")
	}
	apiUser := os.Getenv("API_USER")
	if apiUser == "" {
		log.Fatal("API_USER is not set in .env file")
	}
	apiPass := os.Getenv("AUTOMATION_SECRET")
	if apiPass == "" {
		log.Fatal("AUTOMATION_SECRET is not set in .env file")
	}

	var newAgent models.Agent
	// Decode the incoming JSON to the newAgent struct
	if err := json.NewDecoder(r.Body).Decode(&newAgent); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := database.RegisterNewAgent(&newAgent); err != nil {
		http.Error(w, "error registering agent", http.StatusInternalServerError)
		return
	}

	// Prepare payload for CheckMK host creation
	payload := map[string]interface{}{
		"folder":    "/",
		"host_name": newAgent.Hostname,
		"attributes": map[string]string{
			"ipaddress": newAgent.IPAddress,
		},
	}
	payloadStr, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "could not marshal payload", http.StatusInternalServerError)
		return
	}

	log.Printf("Payload: %s\n", payloadStr)

	// Create a new request
	req, err := http.NewRequest("POST", apiURL+"/domain-types/host_config/collections/all", strings.NewReader(string(payloadStr)))
	if err != nil {
		http.Error(w, "error creating request", http.StatusInternalServerError)
		return
	}

	// Set the content type to application/json
	req.Header.Set("Content-Type", "application/json")
	// Set the authorization header
	req.Header.Set("Authorization", "Bearer "+apiUser+" "+apiPass)
	// req.Header.Set
	req.Header.Set("Accept", "application/json")

	// log.Printf("Request: %v\n", req)

	// Send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "error sending request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	//log.Printf("Response: %v\n", resp)

	//Generate a one-time token for the agent
	token := uuid.New().String()
	newAgent.Token = token

	// Convert the agent ID to a string
	agentIDStr := strconv.Itoa(int(newAgent.ID))

	// Store the token and the agent ID in the agentTokens map
	agentTokens[agentIDStr] = token

	// Respond with the registered agent
	w.WriteHeader(http.StatusCreated)
	newAgentWithToken := newAgent
	newAgentWithToken.Token = token
	json.NewEncoder(w).Encode(newAgentWithToken)

	// Sleep for 5 seconds to allow host creation to complete
	time.Sleep(5 * time.Second)

	// Run the CheckMK service discovery script
	cmd := exec.Command("./handlers/cmk_svcd.sh", newAgent.Hostname)
	cmd.Env = append(os.Environ(), "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		log.Printf("cmd.Run() failed with %s\n", err)
	}
}

// Verify agent token and return $AUTOMATION_SECRET
func VerifyAgentToken(w http.ResponseWriter, r *http.Request) {
	log.Println("Received token for automation secret request")
	// Decode the incoming JSON to get the token and agent ID
	var data map[string]string
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		log.Printf("could not decode request body: %v\n", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	token, ok := data["token"]
	if !ok {
		log.Printf("Token not provided\n")
		http.Error(w, "Token not provided", http.StatusBadRequest)
		return
	}

	agentID, ok := data["agent_id"]
	if !ok {
		log.Printf("Agent ID not provided\n")
		http.Error(w, "Agent ID not provided", http.StatusBadRequest)
		return
	}

	//Verify the token
	storedToken, ok := agentTokens[agentID]
	if !ok || token != storedToken {
		log.Printf("Invalid token.\n")
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	} else {
		log.Printf("Token verified for agent %s", agentID)
	}

	// Delete the token from the agentTokens map
	delete(agentTokens, agentID)

	// If the token is valid, respond with the AUTOMATION_SECRET
	automationSecret := os.Getenv("AUTOMATION_SECRET")
	if automationSecret == "" {
		log.Println("AUTOMATION_SECRET not set")
		http.Error(w, "AUTOMATION_SECRET not set", http.StatusInternalServerError)
		return
	}

	// Respond with the AUTOMATION_SECRET
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]string{"secret": automationSecret}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("could not encode response: %v\n", err)
		http.Error(w, "could not encode response", http.StatusInternalServerError)
		return
	}

	// Log the response that was sent
	// log.Printf("Response: %v\n", response)
}

// CheckMKServiceDiscovery runs the CheckMK service discovery script
func CMKSvcDiscovery(w http.ResponseWriter, r *http.Request) {
	log.Println("Received service discovery request")
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		http.Error(w, "could not load .env file", http.StatusInternalServerError)
		return
	}

	apiURL := os.Getenv("API_URL")
	if apiURL == "" {
		log.Fatal("API_URL is not set in .env file")
	}
	apiUser := os.Getenv("API_USER")
	if apiUser == "" {
		log.Fatal("API_USER is not set in .env file")
	}
	apiPass := os.Getenv("AUTOMATION_SECRET")
	if apiPass == "" {
		log.Fatal("AUTOMATION_SECRET is not set in .env file")
	}

	// Decode the incoming JSON to get the hostname
	var data map[string]string
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	hostname := data["host_name"]

	// Run the CheckMK service discovery script
	log.Println("Running service discovery script")
	cmd := exec.Command("./handlers/cmk_svcd.sh", hostname)
	cmd.Env = append(os.Environ(), "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		log.Printf("cmd.Run() failed with %s\n", err)
	}
	log.Println("Service discovery script complete")
}

// GetAllAgents returns all the agents in the database
func GetAllAgents(w http.ResponseWriter, r *http.Request) {
	agents, err := database.GetAllAgents()
	if err != nil {
		log.Printf("error getting agents: %v", err)
		http.Error(w, "error getting agents", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(agents)
}

// GetAgent returns a single agent from the database
func GetAgent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	agent, err := database.GetAgent(id)
	if err != nil {
		log.Printf("error getting agent: %v", err)
		http.Error(w, "error getting agent", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(agent)
}

// UpdateAgent updates an agent in the database
func UpdateAgent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var updatedAgent models.Agent
	err := json.NewDecoder(r.Body).Decode(&updatedAgent)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = database.UpdateAgent(id, &updatedAgent)
	if err != nil {
		http.Error(w, "error updating agent", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DeleteAgent deletes an agent from the database
func DeleteAgent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	err := database.DeleteAgent(id)
	if err != nil {
		http.Error(w, "error deleting agent", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// AgentHeartbeat updates tha agent data in the database
func AgentHeartbeat(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	err := database.AgentHeartbeat(id)
	if err != nil {
		http.Error(w, "error updating agent", http.StatusInternalServerError)
		return
	}
}

// QueryLivestatusHandler handles Livestatus queries
func QueryLivestatusHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the query from the request
	query := r.URL.Query().Get("query")
	if query == "" {
		http.Error(w, "Missing query parameter", http.StatusBadRequest)
		return
	}

	// Call the QueryLivestatus function with the query
	response, err := livestatus.QueryLivestatus(html.EscapeString(query))
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to query Livestatus: %v", err), http.StatusInternalServerError)
		return
	}

	// Write the response to the client
	w.Write([]byte(response))
}
