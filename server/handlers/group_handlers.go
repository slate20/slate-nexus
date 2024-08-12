package handlers

import (
	"encoding/json"
	"net/http"
	"slate-rmm/database"
	"strconv"

	"github.com/gorilla/mux"
)

// GetAllGroups handles the GET /api/groups route
func GetAllGroups(w http.ResponseWriter, r *http.Request) {
	groups, err := database.GetAllGroups()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(groups)
}

// GetGroup handles the GET /api/groups/{group_id} route
func GetGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupID := vars["group_id"]

	group, err := database.GetGroup(groupID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(group)
}

// CreateGroup handles the POST /api/groups route
func CreateGroup(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		GroupName string `json:"group_name"`
	}

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = database.CreateGroup(payload.GroupName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Group created"))
}

// UpdateGroup handles the PUT /api/groups/{group_id} route
func UpdateGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupID := vars["group_id"]

	var payload struct {
		GroupName string `json:"group_name"`
	}

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = database.UpdateGroup(groupID, payload.GroupName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Group " + groupID + " updated"))
}

// DeleteGroup handles the DELETE /api/groups/{group_id} route
func DeleteGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupID := vars["group_id"]

	err := database.DeleteGroup(groupID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Group " + groupID + " deleted"))
}

// GetHostsInGroup handles the GET /api/groups/{group_id}/hosts route
func GetHostsInGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupID, _ := strconv.Atoi(vars["group_id"])

	hosts, err := database.GetHostsInGroup(groupID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(hosts)
}

// AddHostToGroup handles the POST /api/groups/{group_id}/add/{host_id} route
func AddHostToGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupID, err := strconv.Atoi(vars["group_id"])
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	hostID, err := strconv.Atoi(vars["host_id"])
	if err != nil {
		http.Error(w, "Invalid host ID", http.StatusBadRequest)
		return
	}

	err = database.AddHostToGroup(hostID, groupID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Host added to group" + vars["group_id"]))
}

// RemoveHostFromGroup handles the DELETE /api/groups/{group_id}/remove/{host_id} route
func RemoveHostFromGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupID, err := strconv.Atoi(vars["group_id"])
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	hostID, err := strconv.Atoi(vars["host_id"])
	if err != nil {
		http.Error(w, "Invalid host ID", http.StatusBadRequest)
		return
	}

	err = database.RemoveHostFromGroup(hostID, groupID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Host removed from group" + vars["group_id"]))
}

// MoveHostToGroup handles the POST /api/groups/{group_id}/move/{host_id} route
func MoveHostToGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupID, err := strconv.Atoi(vars["group_id"])
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	hostID, err := strconv.Atoi(vars["host_id"])
	if err != nil {
		http.Error(w, "Invalid host ID", http.StatusBadRequest)
		return
	}

	err = database.MoveHostToGroup(hostID, groupID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Host moved to group" + vars["group_id"]))
}
