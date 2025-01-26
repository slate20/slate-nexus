package handlers

import (
	"net/http"
)

// Handler for rendering the index page
func IndexHandler(w http.ResponseWriter, r *http.Request) {

	// Render the template
	RenderTemplate(w, "index.html", nil)
}
