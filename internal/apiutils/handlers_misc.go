package apiutils

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// Handler for /admin/metrics
// Returns an html view of the server metrics
func (api *ApiUtils) HandlerMetrics(w http.ResponseWriter, r *http.Request) {
	html := `
	<html>
		<body>
			<h1>Welcome, Chirpy Admin</h1>
			<p>Chirpy has been visited %d times!</p>
		</body>
	</html>	
	`
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(200)
	fmt.Fprintf(w, html, api.Cfg.fileserverHits.Load())
}

// Handler for /admin/reset
// Resets the number of hits to the file server and deletes all the users from the database
func (api *ApiUtils) HandlerReset(w http.ResponseWriter, r *http.Request) {
	if api.Cfg.Platform != "dev" {
		w.WriteHeader(403)
		return
	}
	if err := api.DbQueries.DeleteUsers(r.Context()); err != nil {
		RespondWithError(w, 500, fmt.Sprintf("Error deleting users: %s", err))
		return
	}
	api.Cfg.fileserverHits.Store(0)
	w.WriteHeader(200)
}

// Handler for /api/healthz
// Returns OK if the server is up
func HandlerHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK\n"))
}

// Helper function to respond with a JSON payload and a status code
func RespondWithJSON(w http.ResponseWriter, code int, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		RespondWithError(w, 500, fmt.Sprintf("Error marshalling JSON: %s", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(data)
}

// Helper function to respond with an error message and status code in JSON format
func RespondWithError(w http.ResponseWriter, errCode int, errMsg string) {
	type responseArgs struct {
		Error string `json:"error"`
	}

	payload := responseArgs{
		Error: errMsg,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(errCode)
	w.Write(data)

	log.Printf("%v\n", errMsg)
}
