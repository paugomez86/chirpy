package apiutils

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/paugomez86/chirpy/internal/auth"
)

// Model struct
type accessToken struct {
	Token string `json:"token"`
}

// Handler for POST /api/refresh
// TODO
func (api *ApiUtils) HandlerRefresh(w http.ResponseWriter, r *http.Request) {
	// Getting bearer token from headers
	refreshTokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		RespondWithError(w, 401, fmt.Sprintf("%v", err))
		return
	}

	// Querying refresh token
	queryResult, err := api.DbQueries.GetRefresh(r.Context(), refreshTokenString)
	if err != nil {
		if err == sql.ErrNoRows {
			RespondWithError(w, 401, "Refresh token not found")
			return
		}
		RespondWithError(w, 500, fmt.Sprintf("Error fetching refresh token: %s", err))
		return
	}

	// Checking if token is revoked.
	// sql.NullTime type has Valid field that is true if the query returned a not NULL value
	if queryResult.RevokedAt.Valid {
		RespondWithError(w, 401, "Refresh token is revoked")
		return
	}

	// Checking if token expired
	if queryResult.ExpiresAt.Before(time.Now()) {
		RespondWithError(w, 401, "Refresh token is expired")
		return
	}

	userQueryResult, err := api.DbQueries.GetUserFromRefreshToken(r.Context(), refreshTokenString)
	if err != nil {
		RespondWithError(w, 500, fmt.Sprintf("Error fetching user from refresh token: %v", err))
		return
	}

	newAccessToken, err := auth.MakeJWT(userQueryResult.ID, api.Cfg.JwtSecret, api.Cfg.AccessTokenDuration)
	if err != nil {
		RespondWithError(w, 500, fmt.Sprintf("Error generating JWT: %s", err))
		return
	}

	// Mapping query result to payload
	payload := accessToken{
		Token: newAccessToken,
	}

	// Response OK
	RespondWithJSON(w, 200, payload)
}

// Handler POST /api/revoke
// Revokes a refresh token by updating the database record
func (api *ApiUtils) HandlerRevoke(w http.ResponseWriter, r *http.Request) {
	// Geting bearer from headers
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		RespondWithError(w, 500, fmt.Sprintf("%v", err))
		return
	}

	// Querying
	_, err = api.DbQueries.RevokeRefresh(r.Context(), tokenString)
	if err != nil {
		if err == sql.ErrNoRows {
			RespondWithError(w, 404, "Refresh token not found")
			return
		}
		RespondWithError(w, 500, fmt.Sprintf("Error revoking token: %v", err))
		return
	}

	// Response with no body
	w.WriteHeader(204)
}

// Helper function to validate password strength
func passwordIsStrong(password string) error {
	// Checking password length. More than 3 characters.
	// Ideally, stronger passwords should be enforced.
	// For testing purposes, this works.
	if len(password) <= 3 {
		return errors.New("Password should be more than 3 chars long\n")
	}
	return nil
}
