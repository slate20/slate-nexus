package handlers

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"slate-rmm/database"
)

// Handler to get all devices
func GetDevices(w http.ResponseWriter, r *http.Request) {
	// Call the GetAllAgents function from the database package
	agents, err := database.GetAllAgents()
	if err != nil {
		http.Error(w, "Failed to fetch devices", http.StatusInternalServerError)
		log.Println("Failed to fetch devices:", err)
		return
	}

	// Load templates
	templates := template.Must(template.ParseGlob(filepath.Join("templates", "*.html")))

	// Mock data TODO: Replace with fetched data
	// devices := []models.Agent{
	// 	{Hostname: "DESKTOP-KJH345", Status: "Online", IPAddress: "192.168.1.10", OS: "Windows 11", LastSeen: "8/12/2024 2:00 PM", LastUser: "mvenhaus"},
	// }

	// Render the template with the fetched data
	err = templates.ExecuteTemplate(w, "device-list.html", agents)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Template execution failed:", err)
		return
	}
}
