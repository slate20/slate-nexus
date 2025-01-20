package main

import (
	"net/http"
	"slate-rmm/handlers"
)

// NewHTMXGateway creates a new router and defines the routes for the HTMX gateway
func NewHTMXGateway() *http.ServeMux {
	router := http.NewServeMux()

	// Servce static files
	fs := http.FileServer(http.Dir("../dashboard"))
	router.Handle("/", fs)

	// Serve the Nexus agent zip file
	router.HandleFunc("/download/agent", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Disposition", "attachment; filename=NexusAgent_win.zip")
		http.ServeFile(w, r, "../agent/NexusAgent_win.zip")
	})

	// HTMX routes
	router.HandleFunc("/htmx/get-devices", handlers.GetDevices)
	router.HandleFunc("/htmx/get-groups", handlers.GetGroups)
	router.HandleFunc("/htmx/remoterequest/{id}", handlers.GetRemoteControlURL)

	return router
}
