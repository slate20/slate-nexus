package handlers

import (
	"log"
	"net/http"
	"slate-rmm/database"
	"strings"
)

// Handler for rendering the group items
func GetGroups(w http.ResponseWriter, r *http.Request) {
	// call the GetAllGroups function from the database package
	groups, err := database.GetAllGroups()
	if err != nil {
		http.Error(w, "Failed to fetch groups", http.StatusInternalServerError)
		log.Println("Failed to fetch groups:", err)
		return
	}

	// Render the template
	RenderTemplate(w, "group-items.html", groups)
}

func EditGroupsModal(w http.ResponseWriter, r *http.Request) {
	groups, err := database.GetAllGroups()
	if err != nil {
		http.Error(w, "Failed to fetch groups", http.StatusInternalServerError)
		return
	}

	// Render the template
	RenderTemplate(w, "edit-groups-modal.html", groups)
}

func CloseModal(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("HX-Trigger", "closeModal, group-list-update")
	w.WriteHeader(http.StatusOK)
}

func CreateGroupModal(w http.ResponseWriter, r *http.Request) {
	// Render the template
	RenderTemplate(w, "create-group-modal.html", nil)
}

func CreateNewGroup(w http.ResponseWriter, r *http.Request) {
	// Call the CreateGroup function from the database package
	err := database.CreateGroup(r.FormValue("group_name"))
	if err != nil {
		http.Error(w, "Failed to create group", http.StatusInternalServerError)
		log.Println("Failed to create group:", err)
		return
	}

	// Call EditGroupsModal
	EditGroupsModal(w, r)
}

func RenameGroup(w http.ResponseWriter, r *http.Request) {
	// Extract the group ID from the URL
	groupID := strings.TrimPrefix(r.URL.Path, "/htmx/rename-group/")

	// Extract the new name from the HX-Prompt header
	newName := r.Header.Get("HX-Prompt")

	// Call the RenameGroup function from the database package
	err := database.UpdateGroup(groupID, newName)
	if err != nil {
		http.Error(w, "Failed to rename group", http.StatusInternalServerError)
		log.Println("Failed to rename group:", err)
		return
	}

	// Call EditGroupsModal
	EditGroupsModal(w, r)
}

func DeleteGroup(w http.ResponseWriter, r *http.Request) {
	// Extract the group ID from the URL
	groupID := strings.TrimPrefix(r.URL.Path, "/htmx/delete-group/")

	// Call the DeleteGroup function from the database package
	err := database.DeleteGroup(groupID)
	if err != nil {
		http.Error(w, "Failed to delete group", http.StatusInternalServerError)
		log.Println("Failed to delete group:", err)
		return
	}

	// Call EditGroupsModal
	EditGroupsModal(w, r)
}
