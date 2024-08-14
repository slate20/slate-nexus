package handlers

import (
	"html/template"
	"net/http"
	"path/filepath"
	"slate-rmm/models"
)

// Handler for rendering the group items
func GetGroups(w http.ResponseWriter, r *http.Request) {
	// Load templates
	templates := template.Must(template.ParseGlob(filepath.Join("server", "templates", "*.html")))

	// Mock data TODO: Replace with fetched data
	groups := []models.Group{
		{GroupName: "Group 1"},
		{GroupName: "Group 2"},
		{GroupName: "Group 3"},
	}

	// Render the template
	err := templates.ExecuteTemplate(w, "group-items.html", groups)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
