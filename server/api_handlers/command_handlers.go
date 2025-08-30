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
	"strings"
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
	// Get the host ID from the URL
	hostID := strings.TrimPrefix(r.URL.Path, "/htmx/command-modal/")

	// Get the host from the database
	host, err := database.GetAgent(hostID)
	if err != nil {
		http.Error(w, "error getting host", http.StatusInternalServerError)
		return
	}

	// DEBUG: Print the host ID
	fmt.Println("Opening command modal for Host ID:", hostID)

	data := map[string]any{
		"Host":         host,
		"SessionStart": time.Now().Unix(),
	}
	handlers.RenderTemplate(w, "command-modal.html", data)
}

// RunCommand makes an entry to the agent_commands table
func RunCommand(w http.ResponseWriter, r *http.Request) {
	hostID, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/htmx/run-command/"))

	commandStr := r.FormValue("command")

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

// GetCommandResults returns the rendered HTML for the command history of a host
func GetCommandResults(w http.ResponseWriter, r *http.Request) {
	// Get the host ID from the URL
	hostID, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/htmx/command-results/"))
	if err != nil {
		http.Error(w, "Invalid host ID", http.StatusBadRequest)
		return
	}

	sinceStr := r.URL.Query().Get("since")
	var since int64
	if sinceStr != "" {
		since, _ = strconv.ParseInt(sinceStr, 10, 64)
	}

	commands, err := database.GetCommandResults(hostID, since)
	if err != nil {
		http.Error(w, "error getting command results", http.StatusInternalServerError)
		return
	}

	for _, cmd := range commands {
		fmt.Fprintf(w, "<div>&gt; %s</div>", cmd.Command)
		if cmd.Status == "failed" {
			fmt.Fprint(w, "<pre>Failed to execute command</pre>")
		} else {
			fmt.Fprintf(w, "<pre>%s</pre>", cmd.Output)
		}
	}
}
