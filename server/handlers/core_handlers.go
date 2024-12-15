package handlers

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

// Handler for rendering the index page
func IndexHandler(w http.ResponseWriter, r *http.Request) {

	// Load templates
	templates := template.Must(template.New("").Funcs(CommonFuncMap).ParseGlob(filepath.Join("templates", "*.html")))

	// Render the template
	err := templates.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Template execution failed:", err)
		return
	}
}
