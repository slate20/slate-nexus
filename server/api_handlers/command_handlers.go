package api_handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"slate-rmm/database"
	"slate-rmm/handlers"
	"slate-rmm/models"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

// GetCommand returns a single pending command from the database
func GetCommand(w http.ResponseWriter, r *http.Request) {
	// Get the host ID from the URL
	vars := mux.Vars(r)
	hostID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid host ID", http.StatusBadRequest)
		return
	}

	// Get the command from the database
	command, err := database.GetPendingCommand(hostID)
	if err != nil {
		// Check for "not found" error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		// handle other errors
		log.Printf("error getting command: %v", err)
		http.Error(w, "error getting command", http.StatusInternalServerError)
		return
	}

	// Respond with the command as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(command)
}

// UpdateCommand updates the status of a command in the database
func UpdateCommand(w http.ResponseWriter, r *http.Request) {
	var command models.AgentCommand
	err := json.NewDecoder(r.Body).Decode(&command)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = database.UpdateCommand(&command)
	if err != nil {
		http.Error(w, "error updating command", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// CommandModal returns the HTML for the command modal
func CommandModal(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hostID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid host ID", http.StatusBadRequest)
		return
	}

	data := map[string]string{
		"host_id": strconv.Itoa(hostID),
	}
	handlers.RenderTemplate(w, "command-modal.html", data)
}

// RunCommand makes an entry to the agent_commands table
func RunCommand(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hostID, err := strconv.Atoi(vars["id"])
	commandStr := r.FormValue("command")
	if err != nil {
		http.Error(w, "Invalid host ID", http.StatusBadRequest)
		return
	}

	// Create the command in the database
	command := &models.AgentCommand{
		HostID:    hostID,
		Command:   commandStr,
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = database.CreateCommand(command)
	if err != nil {
		http.Error(w, "error creating command", http.StatusInternalServerError)
		return
	}

	// Return the command to be displayed in the history
	fmt.Fprintf(w, "<div>&gt; %s</div>", commandStr)
}
