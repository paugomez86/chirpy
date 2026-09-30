package apiutils

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/paugomez86/chirpy/internal/auth"
	"github.com/paugomez86/chirpy/internal/database"
)

// Model struct
type user struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	AccessToken  string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	IsChirpyRed  bool      `json:"is_chirpy_red"`
}

// Request args struct
type userArgs struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type polkaWebhookArgs struct {
	Event string `json:"event"`
	Data  struct {
		UserId uuid.UUID `json:"user_id"`
	} `json:"data"`
}

// GET /api/users
// Returns all the users in a JSON response
func (api *ApiUtils) HandlerGetUsers(w http.ResponseWriter, r *http.Request) {
	// Querying
	queryResult, err := api.DbQueries.GetUsers(r.Context())
	if err != nil {
		RespondWithError(w, 500, fmt.Sprintf("Error fetching users: %s\n", err))
		return
	}

	// Mapping database result to payload
	var payload []user
	for _, item := range queryResult {
		payload = append(payload, user{
			ID:          item.ID,
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
			Email:       item.Email,
			IsChirpyRed: item.IsChirpyRed,
		})
	}

	// Response OK
	RespondWithJSON(w, 200, payload)
}

// POST /api/users
// Expects a JSON with the new user email and returns a JSON with the newly created user data.
func (api *ApiUtils) HandlerCreateUser(w http.ResponseWriter, r *http.Request) {
	// Decoding request data
	decoder := json.NewDecoder(r.Body)
	reqArgs := userArgs{}

	if err := decoder.Decode(&reqArgs); err != nil {
		RespondWithError(w, 500, fmt.Sprintf("Error decoding request: %s", err))
		return
	}

	// Checking password strength.
	err := passwordIsStrong(reqArgs.Password)
	if err != nil {
		RespondWithError(w, 400, fmt.Sprintf("Password strength error: %v", err))
		return
	}

	// Hashing password arg
	hashedPassword, err := auth.HashPassword(reqArgs.Password)
	if err != nil {
		RespondWithError(w, 500, fmt.Sprintf("Error hashing password: %s", err))
		return
	}

	// Building query args
	queryArgs := database.CreateUserParams{
		Email:          reqArgs.Email,
		HashedPassword: hashedPassword,
	}

	// Querying
	queryResult, err := api.DbQueries.CreateUser(r.Context(), queryArgs)
	if err != nil {
		RespondWithError(w, 500, fmt.Sprintf("Error creating user: %s\n", err))
		return
	}

	// Mapping new user data to response payload
	payload := user{
		ID:          queryResult.ID,
		CreatedAt:   queryResult.CreatedAt,
		UpdatedAt:   queryResult.UpdatedAt,
		Email:       queryResult.Email,
		IsChirpyRed: queryResult.IsChirpyRed,
	}

	// Response OK
	RespondWithJSON(w, 201, payload)
}

// PUT /api/users
// Updates user data.
func (api *ApiUtils) HandlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	// Getting access token from headers
	accessTokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		RespondWithError(w, 401, fmt.Sprintf("Invalid JWT: %s", err))
		return
	}
	// Validating JWT
	userId, err := auth.ValidateJWT(accessTokenString, api.Cfg.JwtSecret)
	if err != nil {
		RespondWithError(w, 401, fmt.Sprintf("Invalid JWT: %s", err))
		return
	}

	// Decoding request data
	decoder := json.NewDecoder(r.Body)
	reqArgs := userArgs{}

	if err := decoder.Decode(&reqArgs); err != nil {
		RespondWithError(w, 500, fmt.Sprintf("Error decoding request: %s", err))
		return
	}

	// Checking password strength.
	err = passwordIsStrong(reqArgs.Password)
	if err != nil {
		RespondWithError(w, 400, fmt.Sprintf("Password strength error: %v", err))
		return
	}

	// Hashing password arg
	hashedPassword, err := auth.HashPassword(reqArgs.Password)
	if err != nil {
		RespondWithError(w, 500, fmt.Sprintf("Error hashing password: %s", err))
		return
	}

	//Building query args
	queryArgs := database.UpdateUserParams{
		ID:             userId,
		Email:          reqArgs.Email,
		HashedPassword: hashedPassword,
	}

	// Querying
	queryResult, err := api.DbQueries.UpdateUser(r.Context(), queryArgs)
	if err != nil {
		if err == sql.ErrNoRows {
			RespondWithJSON(w, 404, "User not found")
			return
		}
		RespondWithError(w, 500, fmt.Sprintf("Error updating user: %s", err))
		return
	}

	// Mapping user data to response payload
	payload := user{
		ID:          queryResult.ID,
		CreatedAt:   queryResult.CreatedAt,
		UpdatedAt:   queryResult.UpdatedAt,
		Email:       queryResult.Email,
		AccessToken: accessTokenString,
		IsChirpyRed: queryResult.IsChirpyRed,
	}

	// Response OK
	RespondWithJSON(w, 200, payload)

}

// POST /api/login
// Gets loginArgs JSON, queries user by email, validates password and generates a JWT token
// Returns user in JSON response
func (api *ApiUtils) HandlerLogin(w http.ResponseWriter, r *http.Request) {
	// Decoding request data
	decoder := json.NewDecoder(r.Body)
	reqArgs := userArgs{}

	if err := decoder.Decode(&reqArgs); err != nil {
		RespondWithError(w, 500, fmt.Sprintf("Error decoding request: %s", err))
		return
	}

	// Querying user by email
	userQueryResult, err := api.DbQueries.GetUserFromEmail(r.Context(), reqArgs.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			RespondWithError(w, 404, "User not found")
			return
		}
		RespondWithError(w, 500, fmt.Sprintf("Error fetching user: %s", err))
		return
	}

	// Validating given password matches database hash
	match, err := auth.CheckPasswordHash(reqArgs.Password, userQueryResult.HashedPassword)

	if err != nil {
		RespondWithError(w, 500, fmt.Sprintf("Error hashing password: %s", err))
		return
	}

	if !match {
		RespondWithError(w, 401, "Incorrect email or password")
		return
	}

	// Generating JWT access token
	accessTokenString, err := auth.MakeJWT(userQueryResult.ID, api.Cfg.JwtSecret, api.Cfg.AccessTokenDuration)
	if err != nil {
		RespondWithError(w, 500, fmt.Sprintf("Error generating JWT: %s", err))
		return
	}

	// Generating refresh token
	refreshTokenString := auth.MakeRefreshToken()
	refreshTokenExpiration := time.Now().Add(api.Cfg.RefreshTokenDuration)

	// Building refresh token query args struct
	refreshQueryArgs := database.CreateRefreshParams{
		Token:     refreshTokenString,
		UserID:    userQueryResult.ID,
		ExpiresAt: refreshTokenExpiration,
	}

	// Querying refresh token creation
	// TODO: Check function return. Remove
	_, err = api.DbQueries.CreateRefresh(r.Context(), refreshQueryArgs)
	if err != nil {
		RespondWithError(w, 500, fmt.Sprintf("Error creating refresh token: %s", err))
		return
	}

	// Mapping user to payload
	payload := user{
		ID:           userQueryResult.ID,
		CreatedAt:    userQueryResult.CreatedAt,
		UpdatedAt:    userQueryResult.UpdatedAt,
		Email:        userQueryResult.Email,
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		IsChirpyRed:  userQueryResult.IsChirpyRed,
	}

	// Response OK
	RespondWithJSON(w, 200, payload)
}

// POST /api/polka/webhooks
// Gets a request from the fictional 3rd party payment service Polka.
// The request is a confirmation that the user paid for Chirpy Red so it can update the user account status.
func (api *ApiUtils) HandlerUpgradeUserRed(w http.ResponseWriter, r *http.Request) {
	// Decoding request body
	decoder := json.NewDecoder(r.Body)
	reqArgs := polkaWebhookArgs{}

	if err := decoder.Decode(&reqArgs); err != nil {
		RespondWithError(w, 500, fmt.Sprintf("Error decoding request: %s", err))
		return
	}

	// Checking provided API key
	APIKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		RespondWithError(w, 401, fmt.Sprintf("Error getting API key: %s", err))
		return
	}

	if APIKey != api.Cfg.PolkaKey {
		RespondWithError(w, 401, fmt.Sprintf("Invalid API key: %s", err))
		return
	}

	// Checking event from JSON data
	if reqArgs.Event != "user.upgraded" {
		w.WriteHeader(204)
		return
	}

	// Querying user status update
	_, err = api.DbQueries.UpgradeUserRed(r.Context(), reqArgs.Data.UserId)
	if err != nil {
		if err == sql.ErrNoRows {
			RespondWithError(w, 404, "User not found")
			return
		}
		RespondWithError(w, 500, fmt.Sprintf("Error upgrading user to Chirpy Red: %v", err))
		return
	}

	// Response with no content
	w.WriteHeader(204)
}
