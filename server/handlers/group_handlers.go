package handlers

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
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

	// Load templates
	templates := template.Must(template.ParseGlob(filepath.Join("templates", "*.html")))

	// Render the template
	err = templates.ExecuteTemplate(w, "group-items.html", groups)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Template execution failed:", err)
		return
	}
}
