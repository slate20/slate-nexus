package main

import (
	"net/http"
	"slate-rmm/handlers"
)

// NewHTMXGateway creates a new router and defines the routes for the HTMX gateway
func NewHTMXGateway() *http.ServeMux {
	router := http.NewServeMux()

	// HTMX routes
	router.HandleFunc("/htmx/get-devices", handlers.GetDevices)
	router.HandleFunc("/htmx/get-groups", handlers.GetGroups)

	return router
}
