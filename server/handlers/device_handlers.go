package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slate-rmm/database"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
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
	templates := template.Must(template.New("").Funcs(CommonFuncMap).ParseGlob(filepath.Join("templates", "*.html")))

	// Render the template with the fetched data
	err = templates.ExecuteTemplate(w, "device-list.html", agents)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Template execution failed:", err)
		return
	}
}

func GetRemoteControlURL(w http.ResponseWriter, r *http.Request) {
	//load .env file
	err := godotenv.Load()
	if err != nil {
		http.Error(w, "could not load .env file", http.StatusInternalServerError)
		return
	}

	apiURL := os.Getenv("REMOTELY_API_URL")
	apiToken := os.Getenv("REMOTELY_API_TOKEN")
	apiID := os.Getenv("REMOTELY_API_ID")

	// get the RemotelyID using the host ID from the end of request URL
	hostID := strings.TrimPrefix(r.URL.Path, "/htmx/remoterequest/")
	if _, err := strconv.Atoi(hostID); err != nil || hostID == "" {
		http.Error(w, "invalid host ID", http.StatusBadRequest)
		return
	}

	agent, err := database.GetAgent(hostID)
	if err != nil {
		http.Error(w, "could not get agent", http.StatusInternalServerError)
		return
	}

	// Validate RemotelyID
	if agent.RemotelyID == "" || !regexp.MustCompile(`[a-zA-Z0-9]+$`).MatchString(agent.RemotelyID) {
		http.Error(w, "invalid RemotelyID", http.StatusBadRequest)
		return
	}

	baseURL, err := url.Parse(apiURL)
	if err != nil {
		http.Error(w, "invalid API URL", http.StatusBadRequest)
		return
	}

	// build the Remotely API URL
	baseURL.Path = path.Join(baseURL.Path, "RemoteControl", agent.RemotelyID)
	url := baseURL.String()

	// build the Remotely API request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		http.Error(w, "could not create request", http.StatusInternalServerError)
		return
	}

	// set the headers
	req.Header.Set("X-Api-Key", apiID+":"+apiToken)

	// send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "could not send request", http.StatusInternalServerError)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Print the response to the console
	fmt.Println(resp)

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "could not read response body", http.StatusInternalServerError)
		return
	}

	// Set HTMX headers
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("HX-Trigger", "remoterequest")

	// Return the URL in a JSON object
	json.NewEncoder(w).Encode(map[string]string{"url": string(body)})
}
