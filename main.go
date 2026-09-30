package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/paugomez86/chirpy/internal/apiutils"
	"github.com/paugomez86/chirpy/internal/database"
)

func main() {
	var api *apiutils.ApiUtils
	var mux *http.ServeMux
	var server *http.Server

	// Loading environment
	godotenv.Load()

	// Database connection
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Error connecting to database\n")
	}

	// Initializing api utils struct instance
	api = &apiutils.ApiUtils{}
	api.LoadConfig()
	api.DbQueries = database.New(db)

	// Initializing ServeMux
	// Registering routes
	mux = &http.ServeMux{}
	mux.Handle("/app/", api.MiddlewareMetricsInc(apiutils.MiddlewareLog(http.StripPrefix("/app/", http.FileServer(http.Dir("."))))))

	mux.Handle("GET /admin/metrics", apiutils.MiddlewareLog(http.HandlerFunc(api.HandlerMetrics)))
	mux.Handle("POST /admin/reset", apiutils.MiddlewareLog(http.HandlerFunc(api.HandlerReset)))
	mux.Handle("GET /api/healthz", apiutils.MiddlewareLog(http.HandlerFunc(apiutils.HandlerHealthz)))

	mux.Handle("GET /api/chirps", apiutils.MiddlewareLog(http.HandlerFunc(api.HandlerGetChirps)))
	mux.Handle("POST /api/chirps", apiutils.MiddlewareLog(http.HandlerFunc(api.HandlerCreateChirp)))
	mux.Handle("GET /api/chirps/{chirpId}", apiutils.MiddlewareLog(http.HandlerFunc(api.HandlerGetChirpFromId)))
	mux.Handle("DELETE /api/chirps/{chirpId}", apiutils.MiddlewareLog(http.HandlerFunc(api.HandlerDeleteChirp)))
	mux.Handle("GET /api/users/{userId}/chirps", apiutils.MiddlewareLog(http.HandlerFunc(api.HandlerGetChirpsFromUser)))

	mux.Handle("GET /api/users", apiutils.MiddlewareLog(http.HandlerFunc(api.HandlerGetUsers)))
	mux.Handle("POST /api/users", apiutils.MiddlewareLog(http.HandlerFunc(api.HandlerCreateUser)))
	mux.Handle("PUT /api/users", apiutils.MiddlewareLog(http.HandlerFunc(api.HandlerUpdateUser)))
	mux.Handle("POST /api/login", apiutils.MiddlewareLog(http.HandlerFunc(api.HandlerLogin)))
	mux.Handle("POST /api/polka/webhooks", apiutils.MiddlewareLog(http.HandlerFunc(api.HandlerUpgradeUserRed)))

	mux.Handle("POST /api/refresh", apiutils.MiddlewareLog(http.HandlerFunc(api.HandlerRefresh)))
	mux.Handle("POST /api/revoke", apiutils.MiddlewareLog(http.HandlerFunc(api.HandlerRevoke)))

	// Initializing server
	server = &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	// Printing some feedback
	fmt.Printf("Running...\n")

	// Logging error if anything goes wrong
	log.Fatal(server.ListenAndServe())
}
