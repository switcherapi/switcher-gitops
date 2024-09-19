package core

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/switcherapi/switcher-gitops/src/config"
	"github.com/switcherapi/switcher-gitops/src/model"
	"github.com/switcherapi/switcher-gitops/src/utils"
)

const SWITCHER_API_JWT_SECRET = "SWITCHER_API_JWT_SECRET"

func TestFetchSnapshotVersion(t *testing.T) {
	t.Run("Should return snapshot version", func(t *testing.T) {
		responsePayload := utils.ReadJsonFromFile("../../resources/fixtures/api/default_snapshot_version.json")
		fakeApiServer := givenApiResponse(http.StatusOK, responsePayload)
		defer fakeApiServer.Close()

		apiService := NewApiService(SWITCHER_API_JWT_SECRET, fakeApiServer.URL)
		version, _ := apiService.FetchSnapshotVersion("domainId", "default")

		assert.Contains(t, version, "version", "Missing version in response")
		assert.Contains(t, version, "domain", "Missing domain in response")
	})

	t.Run("Should return error - invalid API key", func(t *testing.T) {
		fakeApiServer := givenApiResponse(http.StatusUnauthorized, `{ "error": "Invalid API token" }`)
		defer fakeApiServer.Close()

		apiService := NewApiService("INVALID_KEY", fakeApiServer.URL)
		version, _ := apiService.FetchSnapshotVersion("domainId", "default")

		assert.Contains(t, version, "Invalid API token")
	})

	t.Run("Should return error - invalid domain", func(t *testing.T) {
		responsePayload := utils.ReadJsonFromFile("../../resources/fixtures/api/error_invalid_domain.json")
		fakeApiServer := givenApiResponse(http.StatusUnauthorized, responsePayload)
		defer fakeApiServer.Close()

		apiService := NewApiService(SWITCHER_API_JWT_SECRET, fakeApiServer.URL)
		version, _ := apiService.FetchSnapshotVersion("INVALID_DOMAIN", "default")

		assert.Contains(t, version, "errors")
	})

	t.Run("Should return error - invalid API URL", func(t *testing.T) {
		apiService := NewApiService(config.GetEnv(SWITCHER_API_JWT_SECRET), "http://localhost:8080")
		_, err := apiService.FetchSnapshotVersion("domainId", "default")

		assert.NotNil(t, err)
	})
}

func TestFetchSnapshot(t *testing.T) {
	t.Run("Should return snapshot", func(t *testing.T) {
		responsePayload := utils.ReadJsonFromFile("../../resources/fixtures/api/default_snapshot.json")
		fakeApiServer := givenApiResponse(http.StatusOK, responsePayload)
		defer fakeApiServer.Close()

		apiService := NewApiService(SWITCHER_API_JWT_SECRET, fakeApiServer.URL)
		snapshot, _ := apiService.FetchSnapshot("domainId", "default")

		assert.Contains(t, snapshot, "domain", "Missing domain in snapshot")
		assert.Contains(t, snapshot, "version", "Missing version in snapshot")
		assert.Contains(t, snapshot, "group", "Missing groups in snapshot")
		assert.Contains(t, snapshot, "config", "Missing config in snapshot")
	})

	t.Run("Should return data from snapshot", func(t *testing.T) {
		responsePayload := utils.ReadJsonFromFile("../../resources/fixtures/api/default_snapshot.json")
		fakeApiServer := givenApiResponse(http.StatusOK, responsePayload)
		defer fakeApiServer.Close()

		apiService := NewApiService(SWITCHER_API_JWT_SECRET, fakeApiServer.URL)
		snapshot, _ := apiService.FetchSnapshot("domainId", "default")
		data := apiService.NewDataFromJson([]byte(snapshot))

		assert.NotNil(t, data.Snapshot.Domain, "domain", "Missing domain in data")
		assert.NotNil(t, data.Snapshot.Domain.Group, "group", "Missing groups in data")
		assert.NotNil(t, data.Snapshot.Domain.Group[0].Config, "config", "Missing config in data")
	})

	t.Run("Should return error - invalid API key", func(t *testing.T) {
		fakeApiServer := givenApiResponse(http.StatusUnauthorized, `{ "error": "Invalid API token" }`)
		defer fakeApiServer.Close()

		apiService := NewApiService("INVALID_KEY", fakeApiServer.URL)
		snapshot, _ := apiService.FetchSnapshot("domainId", "default")

		assert.Contains(t, snapshot, "Invalid API token")
	})

	t.Run("Should return error - invalid domain", func(t *testing.T) {
		responsePayload := utils.ReadJsonFromFile("../../resources/fixtures/api/error_invalid_domain.json")
		fakeApiServer := givenApiResponse(http.StatusUnauthorized, responsePayload)
		defer fakeApiServer.Close()

		apiService := NewApiService(SWITCHER_API_JWT_SECRET, fakeApiServer.URL)
		snapshot, _ := apiService.FetchSnapshot("INVALID_DOMAIN", "default")

		assert.Contains(t, snapshot, "errors")
	})

	t.Run("Should return error - invalid API URL", func(t *testing.T) {
		apiService := NewApiService(config.GetEnv(SWITCHER_API_JWT_SECRET), "http://localhost:8080")
		_, err := apiService.FetchSnapshot("domainId", "default")

		assert.NotNil(t, err)
	})
}

func TestPushChangesToAPI(t *testing.T) {
	t.Run("Should push changes to API", func(t *testing.T) {
		// Given
		diff := givenDiffResult()
		fakeApiServer := givenApiResponse(http.StatusOK, `{ 
			"version": 2, 
			"message": "Changes applied successfully" 
		}`)
		defer fakeApiServer.Close()

		apiService := NewApiService(SWITCHER_API_JWT_SECRET, fakeApiServer.URL)

		// Test
		response, _ := apiService.PushChanges("domainId", "default", diff)

		// Assert
		assert.NotNil(t, response)
		assert.Equal(t, 2, response.Version)
		assert.Equal(t, "Changes applied successfully", response.Message)
	})

	t.Run("Should return error - invalid API key", func(t *testing.T) {
		// Given
		diff := givenDiffResult()
		fakeApiServer := givenApiResponse(http.StatusUnauthorized, `{ "message": "Invalid API token" }`)
		defer fakeApiServer.Close()

		apiService := NewApiService("[INVALID_KEY]", fakeApiServer.URL)

		// Test
		response, _ := apiService.PushChanges("domainId", "default", diff)

		// Assert
		assert.NotNil(t, response)
		assert.Contains(t, response.Message, "Invalid API token")
	})

	t.Run("Should return error - API not accessible", func(t *testing.T) {
		// Given
		diff := givenDiffResult()
		apiService := NewApiService("[SWITCHER_API_JWT_SECRET]", "http://localhost:8080")

		// Test
		_, err := apiService.PushChanges("domainId", "default", diff)

		// Assert
		assert.NotNil(t, err)
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
