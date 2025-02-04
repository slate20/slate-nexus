package api_handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"slate-rmm/database"
	"slate-rmm/models"
	"time"

	"github.com/gorilla/mux"
)

var agentTokens = make(map[string]string)

// AgentRegistration handles the registration of a new agent
func AgentRegistration(w http.ResponseWriter, r *http.Request) {
	var newAgent models.Agent
	// Decode the incoming JSON to the newAgent struct
	if err := json.NewDecoder(r.Body).Decode(&newAgent); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := database.RegisterNewAgent(&newAgent); err != nil {
		http.Error(w, "error registering agent", http.StatusInternalServerError)
		// Print the error message
		log.Printf("error registering agent: %v\n", err)
		return
	}

	// Convert the agent ID to a string
	// agentIDStr := strconv.Itoa(int(newAgent.ID))

	// Respond with the registered agent
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newAgent)

	// Sleep for 5 seconds to allow host creation to complete
	time.Sleep(5 * time.Second)
}

// Verify agent token and return $API_SECRET
func VerifyAgentToken(w http.ResponseWriter, r *http.Request) {
	log.Println("Received token for API secret request")
	// Decode the incoming JSON to get the token and agent ID
	var data map[string]string
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		log.Printf("could not decode request body: %v\n", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	agentID, ok := data["agent_id"]
	if !ok {
		log.Printf("Agent ID not provided\n")
		http.Error(w, "Agent ID not provided", http.StatusBadRequest)
		return
	}

	// Respond with the host_ID
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]string{"host_id": agentID}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("could not encode response: %v\n", err)
		http.Error(w, "could not encode response", http.StatusInternalServerError)
		return
	}

	// Log the response that was sent
	// log.Printf("Response: %v\n", response)
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
