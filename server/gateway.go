package main

import (
	"net/http"
	"slate-rmm/handlers"

	"github.com/gorilla/mux"
)

// NewGateway creates a new router and defines the routes for the microservices
func NewGateway() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	// Define routes for each microservice
	agentRoutes(router.PathPrefix("/api/agents").Subrouter())
	groupRoutes(router.PathPrefix("/api/groups").Subrouter())

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
	router.HandleFunc("/api/livestatus", handlers.QueryLivestatusHandler).Methods("GET")

	return router

}

// agentRoutes defines the routes for the agent database microservice
func agentRoutes(router *mux.Router) {
	router.HandleFunc("/register", handlers.AgentRegistration).Methods("POST")
	router.HandleFunc("/cmksvcd", handlers.CMKSvcDiscovery).Methods("POST")
	router.HandleFunc("", handlers.GetAllAgents).Methods("GET")
	router.HandleFunc("/{id}", handlers.GetAgent).Methods("GET")
	router.HandleFunc("/secret", handlers.VerifyAgentToken).Methods("POST")
	router.HandleFunc("/{id}", handlers.UpdateAgent).Methods("PUT")
	router.HandleFunc("/{id}", handlers.DeleteAgent).Methods("DELETE")
	router.HandleFunc("/{id}/heartbeat", handlers.AgentHeartbeat).Methods("POST")
}

// groupRoutes defines the routes for the group database microservice
func groupRoutes(router *mux.Router) {
	router.HandleFunc("", handlers.GetAllGroups).Methods("GET")
	router.HandleFunc("/{group_id}", handlers.GetGroup).Methods("GET")
	router.HandleFunc("", handlers.CreateGroup).Methods("POST")
	router.HandleFunc("/{group_id}", handlers.UpdateGroup).Methods("PUT")
	router.HandleFunc("/{group_id}", handlers.DeleteGroup).Methods("DELETE")
	router.HandleFunc("/{group_id}/hosts", handlers.GetHostsInGroup).Methods("GET")
	router.HandleFunc("/{group_id}/add/{host_id}", handlers.AddHostToGroup).Methods("POST")
	router.HandleFunc("/{group_id}/remove/{host_id}", handlers.RemoveHostFromGroup).Methods("DELETE")
	router.HandleFunc("/{group_id}/move/{host_id}", handlers.MoveHostToGroup).Methods("PUT")
}
