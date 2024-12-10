package handlers

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// Handler for rendering the index page
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	hostname, err := os.Hostname()
	if err != nil {
		http.Error(w, "Failed to get hostname", http.StatusInternalServerError)
		log.Println("Failed to get hostname:", err)
		return
	}

	remotelyURL := "https://" + hostname + ":4000"

	// Load templates
	templates := template.Must(template.New("").Funcs(CommonFuncMap).ParseGlob(filepath.Join("templates", "*.html")))

	// Render the template
	err = templates.ExecuteTemplate(w, "index.html", remotelyURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Template execution failed:", err)
		return
	}
}
