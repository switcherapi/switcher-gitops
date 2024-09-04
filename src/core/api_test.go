package core

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/switcherapi/switcher-gitops/src/config"
	"github.com/switcherapi/switcher-gitops/src/model"
)

func TestFetchSnapshot(t *testing.T) {
	if !canRunIntegratedTests() {
		t.Skip(SkipMessage)
	}

	t.Run("Should return snapshot", func(t *testing.T) {
		apiService := NewApiService(config.GetEnv("SWITCHER_API_JWT_SECRET"), config.GetEnv("SWITCHER_API_URL"))
		snapshot, _ := apiService.FetchSnapshot(config.GetEnv("API_DOMAIN_ID"), "default")

		assert.Contains(t, snapshot, "domain", "Missing domain in snapshot")
		assert.Contains(t, snapshot, "version", "Missing version in snapshot")
		assert.Contains(t, snapshot, "group", "Missing groups in snapshot")
		assert.Contains(t, snapshot, "config", "Missing config in snapshot")
	})

	t.Run("Should return data from snapshot", func(t *testing.T) {
		apiService := NewApiService(config.GetEnv("SWITCHER_API_JWT_SECRET"), config.GetEnv("SWITCHER_API_URL"))
		snapshot, _ := apiService.FetchSnapshot(config.GetEnv("API_DOMAIN_ID"), "default")
		data := apiService.NewDataFromJson([]byte(snapshot))

		assert.NotNil(t, data.Snapshot.Domain, "domain", "Missing domain in data")
		assert.NotNil(t, data.Snapshot.Domain.Group, "group", "Missing groups in data")
		assert.NotNil(t, data.Snapshot.Domain.Group[0].Config, "config", "Missing config in data")
	})

	t.Run("Should return error - invalid API key", func(t *testing.T) {
		apiService := NewApiService("INVALID_KEY", config.GetEnv("SWITCHER_API_URL"))
		snapshot, _ := apiService.FetchSnapshot(config.GetEnv("API_DOMAIN_ID"), "default")

		assert.Contains(t, snapshot, "Invalid API token")
	})

	t.Run("Should return error - invalid domain", func(t *testing.T) {
		apiService := NewApiService(config.GetEnv("SWITCHER_API_JWT_SECRET"), config.GetEnv("SWITCHER_API_URL"))
		snapshot, _ := apiService.FetchSnapshot("INVALID_DOMAIN", "default")

		assert.Contains(t, snapshot, "errors")
	})

	t.Run("Should return error - invalid API URL", func(t *testing.T) {
		apiService := NewApiService(config.GetEnv("SWITCHER_API_JWT_SECRET"), "http://localhost:8080")
		_, err := apiService.FetchSnapshot(config.GetEnv("API_DOMAIN_ID"), "default")

		AssertNotNil(t, err)
	})
}

func TestApplyChangesToAPI(t *testing.T) {
	t.Run("Should apply changes to API", func(t *testing.T) {
		// Given
		diff := givenDiffResult()
		fakeApiServer := givenApiResponse(http.StatusOK, `{ 
			"version": "2", 
			"message": "Changes applied successfully" 
		}`)
		defer fakeApiServer.Close()

		apiService := NewApiService("[SWITCHER_API_JWT_SECRET]", fakeApiServer.URL)

		// Test
		response, _ := apiService.ApplyChangesToAPI("domainId", "default", diff)

		// Assert
		assert.NotNil(t, response)
		assert.Equal(t, "2", response.Version)
		assert.Equal(t, "Changes applied successfully", response.Message)
	})
}

// Helpers

func givenDiffResult() model.DiffResult {
	diffResult := model.DiffResult{}
	diffResult.Changes = append(diffResult.Changes, model.DiffDetails{
		Action:  string(NEW),
		Diff:    string(CONFIG),
		Path:    []string{"Release 1", "MY_SWITCHER"},
		Content: map[string]interface{}{"activated": true},
	})
	return diffResult
}

func givenApiResponse(status int, response string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(response))
	}))
}
