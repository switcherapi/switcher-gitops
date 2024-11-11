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
const DEFAULT_SNAPSHOT = "../../resources/fixtures/api/default_snapshot.json"
const DEFAULT_SNAPSHOT_INVALID = "../../resources/fixtures/api/error_invalid_domain.json"
const DEFAULT_SNAPSHOT_VERSION = "../../resources/fixtures/api/default_snapshot_version.json"
const DEFAULT_SNAPSHOT_VERSION_INVALID = "../../resources/fixtures/api/error_invalid_domain.json"

func TestFetchSnapshotVersion(t *testing.T) {
	t.Run("Should return snapshot version", func(t *testing.T) {
		responsePayload := utils.ReadJsonFromFile(DEFAULT_SNAPSHOT_VERSION)
		fakeApiServer := givenApiResponse(http.StatusOK, responsePayload)
		defer fakeApiServer.Close()

		apiService := NewApiService(SWITCHER_API_JWT_SECRET, fakeApiServer.URL, "")
		version, _ := apiService.FetchSnapshotVersion("domainId", "default")

		assert.Contains(t, version, "version", "Missing version in response")
		assert.Contains(t, version, "domain", "Missing domain in response")
	})

	t.Run("Should return error - invalid API key", func(t *testing.T) {
		fakeApiServer := givenApiResponse(http.StatusUnauthorized, `{ "error": "Invalid API token" }`)
		defer fakeApiServer.Close()

		apiService := NewApiService("INVALID_KEY", fakeApiServer.URL, "")
		version, _ := apiService.FetchSnapshotVersion("domainId", "default")

		assert.Contains(t, version, "Invalid API token")
	})

	t.Run("Should return error - invalid domain", func(t *testing.T) {
		responsePayload := utils.ReadJsonFromFile(DEFAULT_SNAPSHOT_VERSION_INVALID)
		fakeApiServer := givenApiResponse(http.StatusUnauthorized, responsePayload)
		defer fakeApiServer.Close()

		apiService := NewApiService(SWITCHER_API_JWT_SECRET, fakeApiServer.URL, "")
		version, _ := apiService.FetchSnapshotVersion("INVALID_DOMAIN", "default")

		assert.Contains(t, version, "errors")
	})

	t.Run("Should return error - invalid API URL", func(t *testing.T) {
		apiService := NewApiService(config.GetEnv(SWITCHER_API_JWT_SECRET), "http://localhost:8080", "")
		_, err := apiService.FetchSnapshotVersion("domainId", "default")

		assert.NotNil(t, err)
	})
}

func TestFetchSnapshot(t *testing.T) {
	t.Run("Should return snapshot", func(t *testing.T) {
		responsePayload := utils.ReadJsonFromFile(DEFAULT_SNAPSHOT)
		fakeApiServer := givenApiResponse(http.StatusOK, responsePayload)
		defer fakeApiServer.Close()

		apiService := NewApiService(SWITCHER_API_JWT_SECRET, fakeApiServer.URL, "")
		snapshot, _ := apiService.FetchSnapshot("domainId", "default")

		assert.Contains(t, snapshot, "domain", "Missing domain in snapshot")
		assert.Contains(t, snapshot, "version", "Missing version in snapshot")
		assert.Contains(t, snapshot, "group", "Missing groups in snapshot")
		assert.Contains(t, snapshot, "config", "Missing config in snapshot")
		assert.Contains(t, snapshot, "relay", "Missing relay in snapshot")
	})

	t.Run("Should return data from snapshot", func(t *testing.T) {
		responsePayload := utils.ReadJsonFromFile(DEFAULT_SNAPSHOT)
		fakeApiServer := givenApiResponse(http.StatusOK, responsePayload)
		defer fakeApiServer.Close()

		apiService := NewApiService(SWITCHER_API_JWT_SECRET, fakeApiServer.URL, "")
		snapshot, _ := apiService.FetchSnapshot("domainId", "default")
		data := apiService.NewDataFromJson([]byte(snapshot))

		assert.NotNil(t, data.Snapshot.Domain, "domain", "Missing domain in data")
		assert.NotNil(t, data.Snapshot.Domain.Group, "group", "Missing groups in data")
		assert.NotNil(t, data.Snapshot.Domain.Group[0].Config, "config", "Missing config in data")
		assert.NotNil(t, data.Snapshot.Domain.Group[0].Config[0].Strategies, "strategies", "Missing strategies in data")
		assert.NotNil(t, data.Snapshot.Domain.Group[0].Config[0].Relay, "relay", "Missing relay in data")
		assert.Contains(t, data.Snapshot.Domain.Group[0].Config[0].Relay.Type, "NOTIFICATION", "Missing relay type in data")
	})

	t.Run("Should return error - invalid API key", func(t *testing.T) {
		fakeApiServer := givenApiResponse(http.StatusUnauthorized, `{ "error": "Invalid API token" }`)
		defer fakeApiServer.Close()

		apiService := NewApiService("INVALID_KEY", fakeApiServer.URL, "")
		snapshot, _ := apiService.FetchSnapshot("domainId", "default")

		assert.Contains(t, snapshot, "Invalid API token")
	})

	t.Run("Should return error - invalid domain", func(t *testing.T) {
		responsePayload := utils.ReadJsonFromFile(DEFAULT_SNAPSHOT_INVALID)
		fakeApiServer := givenApiResponse(http.StatusUnauthorized, responsePayload)
		defer fakeApiServer.Close()

		apiService := NewApiService(SWITCHER_API_JWT_SECRET, fakeApiServer.URL, "")
		snapshot, _ := apiService.FetchSnapshot("INVALID_DOMAIN", "default")

		assert.Contains(t, snapshot, "errors")
	})

	t.Run("Should return error - invalid API URL", func(t *testing.T) {
		apiService := NewApiService(config.GetEnv(SWITCHER_API_JWT_SECRET), "http://localhost:8080", "")
		_, err := apiService.FetchSnapshot("domainId", "default")

		assert.NotNil(t, err)
	})
}

func TestPushChangesToAPI(t *testing.T) {
	t.Run("Should push changes to API", func(t *testing.T) {
		// Given
		diff := givenDiffResult("default")
		fakeApiServer := givenApiResponse(http.StatusOK, `{ 
			"version": 2, 
			"message": "Changes applied successfully" 
		}`)
		defer fakeApiServer.Close()

		apiService := NewApiService(SWITCHER_API_JWT_SECRET, fakeApiServer.URL, "")

		// Test
		response, _ := apiService.PushChanges("domainId", diff)

		// Assert
		assert.NotNil(t, response)
		assert.Equal(t, 2, response.Version)
		assert.Equal(t, "Changes applied successfully", response.Message)
	})

	t.Run("Should return error - invalid payload (400)", func(t *testing.T) {
		// Given
		diff := givenDiffResult("default")
		fakeApiServer := givenApiResponse(http.StatusBadRequest, `{ "error": "Config already exists" }`)
		defer fakeApiServer.Close()

		apiService := NewApiService(SWITCHER_API_JWT_SECRET, fakeApiServer.URL, "")

		// Test
		_, err := apiService.PushChanges("domainId", diff)

		// Assert
		assert.NotNil(t, err)
		assert.Equal(t, "Config already exists", err.Error())

	})

	t.Run("Should return error - invalid API key (401)", func(t *testing.T) {
		// Given
		diff := givenDiffResult("default")
		fakeApiServer := givenApiResponse(http.StatusUnauthorized, `{ "error": "Invalid API token" }`)
		defer fakeApiServer.Close()

		apiService := NewApiService("[INVALID_KEY]", fakeApiServer.URL, "")

		// Test
		_, err := apiService.PushChanges("domainId", diff)

		// Assert
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "Invalid API token")
	})

	t.Run("Should return error - API not accessible", func(t *testing.T) {
		// Given
		diff := givenDiffResult("default")
		apiService := NewApiService("[SWITCHER_API_JWT_SECRET]", "http://localhost:8080", "")

		// Test
		_, err := apiService.PushChanges("domainId", diff)

		// Assert
		assert.NotNil(t, err)
	})
}

func TestFetchSnapshotWithCaCert(t *testing.T) {
	t.Run("Should return snapshot", func(t *testing.T) {
		responsePayload := utils.ReadJsonFromFile("../../resources/fixtures/api/default_snapshot.json")
		fakeApiServer := givenApiResponse(http.StatusOK, responsePayload)
		defer fakeApiServer.Close()

		apiService := NewApiService(SWITCHER_API_JWT_SECRET, fakeApiServer.URL, "../../resources/fixtures/api/dummy.pem")
		snapshot, _ := apiService.FetchSnapshot("domainId", "default")

		assert.Contains(t, snapshot, "domain", "Missing domain in snapshot")
		assert.Contains(t, snapshot, "version", "Missing version in snapshot")
		assert.Contains(t, snapshot, "group", "Missing groups in snapshot")
		assert.Contains(t, snapshot, "config", "Missing config in snapshot")
	})

	t.Run("Should return error - certificate not found", func(t *testing.T) {
		responsePayload := utils.ReadJsonFromFile("../../resources/fixtures/api/default_snapshot.json")
		fakeApiServer := givenApiResponse(http.StatusOK, responsePayload)
		defer fakeApiServer.Close()

		apiService := NewApiService(SWITCHER_API_JWT_SECRET, fakeApiServer.URL, "invalid.pem")
		_, err := apiService.FetchSnapshot("domainId", "default")

		assert.NotNil(t, err)
	})
}

// Helpers

func givenDiffResult(environment string) model.DiffResult {
	diffResult := model.DiffResult{}
	diffResult.Environment = environment
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
