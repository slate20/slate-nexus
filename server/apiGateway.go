package main

import (
	"net/http"
	"slate-rmm/api_handlers"

	"github.com/gorilla/mux"
)

// NewGateway creates a new router and defines the routes for the microservices
func NewGateway() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	// Define routes for each microservice
	agentRoutes(router.PathPrefix("/agents").Subrouter())
	groupRoutes(router.PathPrefix("/groups").Subrouter())

	// Serve the agent executable
	router.HandleFunc("/download/agent", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Disposition", "attachment; filename=slate-rmm-agent.exe")
		http.ServeFile(w, r, "../agent/slate-rmm-agent.exe")
	})

	// Serve the Remotely Windows agent
	router.HandleFunc("/download/remotely-win", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Disposition", "attachment; Install-Remotely.ps1")
		http.ServeFile(w, r, "../agent/Install-Remotely.ps1")
	})

	// Route for Livestatus queries
	router.HandleFunc("/api/livestatus", api_handlers.QueryLivestatusHandler).Methods("GET")

	return router

}

// agentRoutes defines the routes for the agent database microservice
func agentRoutes(router *mux.Router) {
	router.HandleFunc("/register", api_handlers.AgentRegistration).Methods("POST")
	router.HandleFunc("", api_handlers.GetAllAgents).Methods("GET")
	router.HandleFunc("/{id}", api_handlers.GetAgent).Methods("GET")
	router.HandleFunc("/secret", api_handlers.VerifyAgentToken).Methods("POST")
	router.HandleFunc("/{id}", api_handlers.UpdateAgent).Methods("PUT")
	router.HandleFunc("/{id}", api_handlers.DeleteAgent).Methods("DELETE")
	router.HandleFunc("/{id}/heartbeat", api_handlers.AgentHeartbeat).Methods("POST")
}

// groupRoutes defines the routes for the group database microservice
func groupRoutes(router *mux.Router) {
	router.HandleFunc("", api_handlers.GetAllGroups).Methods("GET")
	router.HandleFunc("/{group_id}", api_handlers.GetGroup).Methods("GET")
	router.HandleFunc("", api_handlers.CreateGroup).Methods("POST")
	router.HandleFunc("/{group_id}", api_handlers.UpdateGroup).Methods("PUT")
	router.HandleFunc("/{group_id}", api_handlers.DeleteGroup).Methods("DELETE")
	router.HandleFunc("/{group_id}/hosts", api_handlers.GetHostsInGroup).Methods("GET")
	router.HandleFunc("/{group_id}/add/{host_id}", api_handlers.AddHostToGroup).Methods("POST")
	router.HandleFunc("/{group_id}/remove/{host_id}", api_handlers.RemoveHostFromGroup).Methods("DELETE")
	router.HandleFunc("/{group_id}/move/{host_id}", api_handlers.MoveHostToGroup).Methods("PUT")
}
