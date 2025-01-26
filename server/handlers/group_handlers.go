package handlers

import (
	"log"
	"net/http"
	"slate-rmm/database"
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
