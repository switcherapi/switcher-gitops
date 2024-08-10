package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/switcherapi/switcher-gitops/src/config"
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
