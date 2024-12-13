package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"slate-rmm/database"
	"sync"
	"time"
)

func main() {
	// Initialize the database connection
	dsn := "host=localhost user=postgres password=slatermm dbname=RMM_db sslmode=disable"
	database.InitDB(dsn)

	// Create a new API router
	apiRouter := NewGateway()

	// Create a new router for the HTMX gateway
	htmxRouter := NewHTMXGateway()

	var wg sync.WaitGroup
	wg.Add(2)

	// Start the API server on port 8123
	go func() {
		defer wg.Done()
		fmt.Println("Starting API server on port 8123...")
		srv := &http.Server{
			Addr:    ":8123",
			Handler: CORSMiddleware(apiRouter),
		}
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("API server failed: %v", err)
		}
	}()

	// Start the HTMX server on port 8080
	go func() {
		defer wg.Done()
		fmt.Println("Starting HTMX server on port 8080...")
		srv := &http.Server{
			Addr:    ":8080",
			Handler: CORSMiddleware(htmxRouter),
		}
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTMX server failed: %v", err)
		}
	}()

	// Wait a moment for servers to start
	time.Sleep(time.Second)

	// Check if servers are responsive
	checkServer("http://localhost:8123")
	checkServer("http://localhost:8080")

	wg.Wait()
}

func checkServer(url string) {
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error checking %s: %v", url, err)
		return
	}
	defer resp.Body.Close()
	log.Printf("%s is responsive. Status: %s", url, resp.Status)
}

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// DEBUG: Print the request headers
		// for key, value := range r.Header {
		// 	fmt.Printf("%s: %s\n", key, value)
		// }

		//Only allow requests from the Nginx Container
		ip, _, err := net.SplitHostPort(r.RemoteAddr)

		if err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		if ip != "172.20.0.100" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		log.Printf("Received %s request to %s from %s", r.Method, r.URL.Path, ip)

		// Set the headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With")

		// If it's just an OPTIONS request, we don't need to go any further
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
