package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/switcherapi/switcher-gitops/src/config"
	"github.com/switcherapi/switcher-gitops/src/model"
)

func TestMain(m *testing.M) {
	os.Setenv("GO_ENV", "test")
	config.InitEnv()
	m.Run()
}

func TestToJsonFromObject(t *testing.T) {
	account := givenAccount(true)
	actual := ToJsonFromObject(account)
	assert.NotNil(t, actual)
}

func TestToMapFromObject(t *testing.T) {
	account := givenAccount(true)
	actual := ToMapFromObject(account)
	assert.NotNil(t, actual)
}

func TestFormatJSON(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		account := givenAccount(true)
		accountJSON := ToJsonFromObject(account)
		actual := FormatJSON(accountJSON)
		assert.NotNil(t, actual)
	})

	t.Run("invalid", func(t *testing.T) {
		actual := FormatJSON("invalid")
		assert.NotNil(t, actual)
	})
}

func TestIsValidJson(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		invalidJson := `{
			"domain": {
				"group": [{
					"name": "Hi There",
					"activated": true
				}]
			}
		}`

		assert.True(t, IsJsonValid(invalidJson, model.Snapshot{}))
	})

	t.Run("invalid", func(t *testing.T) {
		invalidJson := `{
			"domain": {
				"group": [{
					"name": "Hi There",
					"activated": true
				}]
			}
		`

		assert.False(t, IsJsonValid(invalidJson, model.Snapshot{}))
	})
}

func TestReadJsonFileToObject(t *testing.T) {
	json := ReadJsonFromFile("../../resources/fixtures/util/default.json")
	assert.NotNil(t, json)
	assert.Contains(t, json, "Release 1")
}

func TestEncrypDecrypt(t *testing.T) {
	privatKey := config.GetEnv("GIT_TOKEN_PRIVATE_KEY")
	text := "github_pat_XXXXXXXXXXXXXXXXXXXXXX_XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"

	encrypted := Encrypt(text, privatKey)
	assert.NotNil(t, encrypted)

	decrypted, err := Decrypt(encrypted, privatKey)
	assert.Nil(t, err)
	assert.Equal(t, text, decrypted)
}

func TestEncrypDecryptError(t *testing.T) {
	privatKey := config.GetEnv("GIT_TOKEN_PRIVATE_KEY")
	text := "github_pat..."

	encrypted := Encrypt(text, privatKey)
	assert.NotNil(t, encrypted)

	decrypted, err := Decrypt("invalid", privatKey)
	assert.NotNil(t, err)
	assert.Equal(t, "", decrypted)
}

func TestGetTimeWindow(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		actual, unit := GetTimeWindow("10m")
		assert.NotNil(t, actual)
		assert.NotNil(t, unit)
	})

	t.Run("invalid", func(t *testing.T) {
		actual, unit := GetTimeWindow("invalid")
		assert.NotNil(t, actual)
		assert.NotNil(t, unit)
	})
}

// Fixtures

func givenAccount(active bool) model.Account {
	return model.Account{
		Repository: "switcherapi/switcher-gitops",
		Branch:     "master",
		Domain: model.DomainDetails{
			ID:         "123-util-test",
			Name:       "Switcher GitOps",
			Version:    123,
			LastCommit: "123",
			Status:     model.StatusSynced,
			Message:    "Synced successfully",
		},
		Settings: &model.Settings{
			Active:     active,
			Window:     "10m",
			ForcePrune: false,
		},
	}
}

// Helpers

func GetDir() string {
	directory, err := os.Getwd()
	if err != nil {
		return ""
	}

	return directory
}
