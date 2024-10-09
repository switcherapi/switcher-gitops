package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

var mr = mux.NewRouter()

func TestTokenMiddleware(t *testing.T) {

	t.Run("Should return success when token is valid", func(t *testing.T) {
		fakeServer := givenFakeServer(http.StatusOK)
		defer fakeServer.Close()

		token := generateToken("test", time.Minute)
		req, _ := http.NewRequest(http.MethodGet, fakeServer.URL+"/api/fake", nil)
		response := executeRequest(req, mr, token)

		assert.Equal(t, http.StatusOK, response.Code)
	})

	t.Run("Should return unauthorized - expired token", func(t *testing.T) {
		fakeServer := givenFakeServer(http.StatusOK)
		defer fakeServer.Close()

		token := generateToken("test", -time.Minute)
		req, _ := http.NewRequest(http.MethodGet, fakeServer.URL+"/api/fake", nil)
		response := executeRequest(req, mr, token)

		assert.Equal(t, http.StatusUnauthorized, response.Code)
		assert.Contains(t, response.Body.String(), "Invalid token")
	})

	t.Run("Should return unauthorized - invalid token", func(t *testing.T) {
		fakeServer := givenFakeServer(http.StatusUnauthorized)
		defer fakeServer.Close()

		req, _ := http.NewRequest(http.MethodGet, fakeServer.URL+"/api/fake", nil)
		response := executeRequest(req, mr, "invalid-token")

		assert.Equal(t, http.StatusUnauthorized, response.Code)
		assert.Contains(t, response.Body.String(), "Invalid token")
	})

	t.Run("Should return unauthorized - missing token", func(t *testing.T) {
		fakeServer := givenFakeServer(http.StatusOK)
		defer fakeServer.Close()

		req, _ := http.NewRequest(http.MethodGet, fakeServer.URL+"/api/fake", nil)
		response := executeRequest(req, mr, "")

		assert.Equal(t, http.StatusUnauthorized, response.Code)
		assert.Contains(t, response.Body.String(), "Invalid token")
	})

	t.Run("Should return unauthorized - invalid token format", func(t *testing.T) {
		fakeServer := givenFakeServer(http.StatusOK)
		defer fakeServer.Close()

		req, _ := http.NewRequest(http.MethodGet, fakeServer.URL+"/api/fake", nil)
		req.Header.Set("Authorization", "invalid-token")
		response := executeRequest(req, mr, "")

		assert.Equal(t, http.StatusUnauthorized, response.Code)
		assert.Contains(t, response.Body.String(), "Invalid token")
	})
}

// Helpers

func givenFakeServer(status int) *httptest.Server {
	mr.Path("/api/fake").Handler(ValidateToken(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		w.Header().Set("Content-Type", "application/json")
	})))

	return httptest.NewServer(mr)
}
