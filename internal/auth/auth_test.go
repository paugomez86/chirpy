package auth

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestPasswordHash(t *testing.T) {
	cases := []string{
		"1234",
		"strong_password1234!",
		"weakpassword",
		"$$password_for_testing7658**+-",
	}

	for _, c := range cases {
		hash, err := HashPassword(c)
		if err != nil {
			fmt.Println(err)
		} else {
			match, err := CheckPasswordHash(c, hash)
			if err != nil {
				fmt.Println(err)
			}
			if !match {
				t.Errorf("Expecting match")
			}
		}
	}
}

func TestJWT(t *testing.T) {

	cases := []struct {
		userIdString   string
		tokenSecret    string
		tokenSecretMod string
		duration       time.Duration
		expectedError  error
		actualError    error
	}{
		{
			userIdString:   "10cdfe4a-7bb4-40b0-82b5-d4e8741a54f4",
			tokenSecret:    "sEcrEt_key&1234",
			tokenSecretMod: "sEcrEt_key&1234",
			duration:       time.Second * 5,
			expectedError:  errors.New(""),
			actualError:    errors.New(""),
		},
		{
			userIdString:   "10cdfe4a-7bb4-40b0-82b5-d4e8741a54f4",
			tokenSecret:    "sEcrEt_key&1234",
			tokenSecretMod: "sEcrEt_key&1234",
			duration:       time.Nanosecond * 1,
			expectedError:  errors.Join(jwt.ErrTokenInvalidClaims, jwt.ErrTokenExpired),
			actualError:    nil,
		},
		{
			userIdString:   "10cdfe4a-7bb4-40b0-82b5-d4e8741a54f4",
			tokenSecret:    "sEcrEt_key&1234",
			tokenSecretMod: "sEcrEt_key&12345",
			duration:       time.Second * 5,
			expectedError:  jwt.ErrInvalidKey,
			actualError:    nil,
		},
	}

	for i, c := range cases {
		// Getting userId
		actualUserId, err := uuid.Parse(c.userIdString)
		if err != nil {
			cases[i].actualError = err
			continue
		}

		// Getting JWT token
		tokenString, err := MakeJWT(actualUserId, c.tokenSecret, c.duration)
		if err != nil {
			cases[i].actualError = err
			continue
		}

		// Validating JWT token
		receivedUserId, err := ValidateJWT(tokenString, c.tokenSecret)
		if err != nil {
			cases[i].actualError = err
			continue
		}

		// Comparing users
		if receivedUserId != actualUserId {
			cases[i].actualError = errors.New("User mismatch")
			continue
		}
	}

	// Checking errors in each case
	for i, c := range cases {
		if errors.Is(c.actualError, c.expectedError) {
			t.Errorf("Case %d: Expecting error: %v - Got error: %v", i, c.expectedError, c.actualError)
		}
	}
}

func TestGetBearerToken(t *testing.T) {
	cases := []struct {
		headers       http.Header
		expectedToken string
		expectedError error
	}{
		{
			headers:       http.Header{"Authorization": []string{"Bearer *******"}},
			expectedToken: "*******",
			expectedError: nil,
		},
		{
			headers:       http.Header{"Authorization": []string{""}},
			expectedToken: "*******",
			expectedError: errors.New("Authorization token not found"),
		},
		{
			headers:       http.Header{"Content-Type": []string{"application/json"}},
			expectedToken: "*******",
			expectedError: errors.New("Authorization token not found"),
		},
	}

	for i, c := range cases {
		stringToken, err := GetBearerToken(c.headers)
		if err == nil {
			if c.expectedToken != stringToken {
				t.Errorf("Test %d. Expected token: %s - Actual token: %s", i, c.expectedToken, stringToken)
			}
		} else {
			if err.Error() != c.expectedError.Error() {
				t.Errorf("Test %d. Expected error: %v - Actual error: %s", i, c.expectedError, err)
			}
		}
	}
}
