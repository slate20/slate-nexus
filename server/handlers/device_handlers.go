package handlers

import (
	"html/template"
	"net/http"
	"path/filepath"
	"slate-rmm/models"
)

// Handler to get all devices
func GetDevices(w http.ResponseWriter, r *http.Request) {
	// Load templates
	templates := template.Must(template.ParseGlob(filepath.Join("server", "templates", "*.html")))

	// Mock data TODO: Replace with fetched data
	devices := []models.Agent{
		{Hostname: "DESKTOP-KJH345", Status: "Online", IPAddress: "192.168.1.10", OS: "Windows 11", LastSeen: "8/12/2024 2:00 PM", LastUser: "mvenhaus"},
	}

	// Render the template
	err := templates.ExecuteTemplate(w, "device-list.html", devices)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
