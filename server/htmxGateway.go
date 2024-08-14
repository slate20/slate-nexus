package main

import (
	"net/http"
	"slate-rmm/handlers"
)

// NewHTMXGateway creates a new router and defines the routes for the HTMX gateway
func NewHTMXGateway() *http.ServeMux {
	router := http.NewServeMux()

	// HTMX routes
	router.HandleFunc("/get-devices", handlers.GetDevices)
	router.HandleFunc("/get-groups", handlers.GetGroups)

	return router
}
