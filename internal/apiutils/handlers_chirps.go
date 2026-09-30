package apiutils

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paugomez86/chirpy/internal/auth"
	"github.com/paugomez86/chirpy/internal/database"
)

// Model struct
type chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

// Request args struct
type chirpArgs struct {
	Body string `json:"body"`
}

// Handler for POST /api/chirps
// It checks length of the chirp. If it's correct, applies the profanity filter
// according to the blacklisted words and adds a new chirp to the database returning the whole data as JSON and code 201.
func (api *ApiUtils) HandlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	// Decoding request data
	decoder := json.NewDecoder(r.Body)
	reqArgs := chirpArgs{}

	if err := decoder.Decode(&reqArgs); err != nil {
		RespondWithError(w, 500, fmt.Sprintf("Error decoding request: %s", err))
		return
	}

	// Validating JWT token
	// Getting token from header
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		RespondWithError(w, 401, fmt.Sprintf("Invalid JWT: %s", err))
		return
	}

	// Validating token
	userId, err := auth.ValidateJWT(tokenString, api.Cfg.JwtSecret)
	if err != nil {
		RespondWithError(w, 401, fmt.Sprintf("Invalid JWT: %s", err))
		return
	}

	// Validating chirp
	// Checking length
	if len(reqArgs.Body) > 140 {
		RespondWithError(w, 400, "Chirp is too long")
		return
	}

	// Profanity filter
	bodyWords := strings.Split(reqArgs.Body, " ")

	for i, word := range bodyWords {
		if slices.Contains(api.Cfg.blacklist, strings.ToLower(word)) {
			bodyWords[i] = "****"
		}
	}

	cleanedBody := strings.Join(bodyWords, " ")

	// Building query args
	queryArgs := database.CreateChirpParams{
		Body:   cleanedBody,
		UserID: userId,
	}

	// Querying
	queryResult, err := api.DbQueries.CreateChirp(r.Context(), queryArgs)
	if err != nil {
		RespondWithError(w, 500, fmt.Sprintf("Error creating chirp: %s\n", err))
		return
	}

	// Mapping chirp to response payload
	payload := chirp{
		ID:        queryResult.ID,
		CreatedAt: queryResult.CreatedAt,
		UpdatedAt: queryResult.UpdatedAt,
		Body:      queryResult.Body,
		UserID:    queryResult.UserID,
	}

	// Response OK
	RespondWithJSON(w, 201, payload)
}

// Handler for GET /api/users
// If author_id is passed as a query parameter, returns all the chirps from that user in a JSON response
// Otherwise, returns all chirps
func (api *ApiUtils) HandlerGetChirps(w http.ResponseWriter, r *http.Request) {
	var queryResult []database.Chirp
	var userId uuid.UUID
	var err error

	// Getting query arguments
	// Parsing string to UUID if a string is passed as argument
	authorId := r.URL.Query().Get("author_id")
	if authorId != "" {
		userId, err = uuid.Parse(authorId)
		if err != nil {
			RespondWithError(w, 500, fmt.Sprintf("Error parsing author_id: %v", err))
			return
		}
	}

	// Querying
	// If userId is nil, it queries all chirps. Otherwise, it queries by user_id
	if userId == uuid.Nil {
		queryResult, err = api.DbQueries.GetChirps(r.Context())
	} else {
		queryResult, err = api.DbQueries.GetChirpsFromUser(r.Context(), userId)
	}
	if err != nil {
		RespondWithError(w, 500, fmt.Sprintf("Error fetching chirps: %s\n", err))
		return
	}

	// Mapping database result to payload
	var payload []chirp
	for _, item := range queryResult {
		payload = append(payload, chirp{
			ID:        item.ID,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
			Body:      item.Body,
			UserID:    item.UserID,
		})
	}

	RespondWithJSON(w, 200, payload)
}

// Handler for GET /api/chirps/{ID}
// Returns a single chirp if the ID given is found
func (api *ApiUtils) HandlerGetChirpFromId(w http.ResponseWriter, r *http.Request) {
	var chirpId uuid.UUID

	// Parsing ID arg
	chirpId, err := uuid.Parse(r.PathValue("chirpId"))
	if err != nil {
		RespondWithError(w, 400, fmt.Sprintf("Invalid ID provided: %s\n", err))
		return
	}

	// Querying
	queryResult, err := api.DbQueries.GetChirpFromId(r.Context(), chirpId)
	if err != nil {
		if err == sql.ErrNoRows {
			RespondWithError(w, 404, "Chirp not found\n")
			return
		}
		RespondWithError(w, 500, fmt.Sprintf("Error fetching chirp: %s\n", err))
		return
	}

	// Mapping database result to payload
	payload := chirp{
		ID:        queryResult.ID,
		CreatedAt: queryResult.CreatedAt,
		UpdatedAt: queryResult.UpdatedAt,
		Body:      queryResult.Body,
		UserID:    queryResult.UserID,
	}

	// Response OK
	RespondWithJSON(w, 200, payload)
}

// Handler for GET /api/users/{ID}/chirps
// Returns the given user's id chirps
func (api *ApiUtils) HandlerGetChirpsFromUser(w http.ResponseWriter, r *http.Request) {
	var userId uuid.UUID

	// Parsing user ID arg
	userId, err := uuid.Parse(r.PathValue("userId"))
	if err != nil {
		RespondWithError(w, 400, fmt.Sprintf("Invalid ID provided: %s\n", err))
		return
	}

	// Querying
	queryResult, err := api.DbQueries.GetChirpsFromUser(r.Context(), userId)
	if err != nil {
		if err == sql.ErrNoRows {
			RespondWithError(w, 404, "Query returned no rows\n")
			return
		}
		RespondWithError(w, 500, fmt.Sprintf("Error fetching chirps: %s\n", err))
		return
	}

	// Mapping database result to payload
	var payload []chirp

	for _, item := range queryResult {
		payload = append(payload, chirp{
			ID:        item.ID,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
			Body:      item.Body,
			UserID:    item.UserID,
		})
	}

	// Response OK
	RespondWithJSON(w, 200, payload)
}

// DELETE /chirps/{chirpID}
func (api *ApiUtils) HandlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	var userId uuid.UUID
	var chirpId uuid.UUID

	// Getting token from header
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		RespondWithError(w, 401, fmt.Sprintf("Invalid JWT: %s", err))
		return
	}

	// Validating token
	userId, err = auth.ValidateJWT(tokenString, api.Cfg.JwtSecret)
	if err != nil {
		RespondWithError(w, 401, fmt.Sprintf("Invalid JWT: %s", err))
		return
	}

	// Parsing chirp ID arg
	chirpId, err = uuid.Parse(r.PathValue("chirpId"))
	if err != nil {
		RespondWithError(w, 400, fmt.Sprintf("Invalid ID provided: %s\n", err))
		return
	}

	// Querying chirp data
	queryResult, err := api.DbQueries.GetChirpFromId(r.Context(), chirpId)
	if err != nil {
		if err == sql.ErrNoRows {
			RespondWithError(w, 404, "Chirp not found")
			return
		}
		RespondWithError(w, 500, fmt.Sprintf("Error getting chirp data: %s", err))
	}

	// Checking if the user is the owner
	if userId != queryResult.UserID {
		RespondWithError(w, 403, "Forbidden deletion")
		return
	}

	// Querying deletion
	err = api.DbQueries.DeleteChirp(r.Context(), chirpId)
	if err != nil {
		RespondWithError(w, 500, fmt.Sprintf("Error deleting chirp: %v", err))
		return
	}

	// Response with no body
	w.WriteHeader(204)
}
