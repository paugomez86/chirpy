package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func HashPassword(password string) (string, error) {
	// Parameters for production purposes.
	// Check https://tools.ietf.org/html/draft-irtf-cfrg-argon2-04#section-4 for guidance.
	/* params := &argon2id.Params{
		Memory:      128 * 1024,
		Iterations:  4,
		Parallelism: uint8(runtime.NumCPU()),
		SaltLength:  16,
		KeyLength:   32,
	} */

	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}
	return hash, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, err
	}
	return match, err
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	// Claims struct contains the data that is tokenized, signed and sent to the user
	jwtClaims := jwt.RegisteredClaims{
		Issuer:    "chirpy-access",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Time.Add(time.Now(), expiresIn)),
		Subject:   userID.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)

	// Token secret is used to ensure the user don't tamper with the token to cheat the server
	signedString, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}

	return signedString, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	var token *jwt.Token
	var userId uuid.UUID

	// Using tokenSecret, populates the referenced Claims struct with the data in tokenString
	// If error is nil, everything is ok. May want to check token.Valid but not strictly necessary
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (any, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return userId, err
	}

	// Subject contains the userId
	userIdString, err := token.Claims.GetSubject()
	if err != nil {
		return userId, err
	}

	userId, err = uuid.Parse(userIdString)
	if err != nil {
		return userId, err
	}

	return userId, nil
}

// Gets Authorization header and returns it if exists. Otherwise returns error.
// Expected format: "Bearer ***". Must trim prefix.
func GetBearerToken(headers http.Header) (string, error) {
	tokenString := headers.Get("Authorization")

	if tokenString == "" {
		return tokenString, fmt.Errorf("Authorization token not found")
	}

	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	return tokenString, nil
}

// Gets Authorization header and returns it if exists. Otherwise returns error.
// Expected format: "ApiKey ***". Must trim prefix.
func GetAPIKey(headers http.Header) (string, error) {
	APIKey := headers.Get("Authorization")

	if APIKey == "" {
		return APIKey, fmt.Errorf("API key not found")
	}

	APIKey = strings.TrimPrefix(APIKey, "ApiKey ")

	return APIKey, nil
}

// Generates a random 32 byte hexadecimal key.
func MakeRefreshToken() string {
	tokenString := make([]byte, 32)
	rand.Read(tokenString)

	return hex.EncodeToString(tokenString)
}
