package utils

import (
	"os"
	"strings"
	"testing"

	"github.com/switcherapi/switcher-gitops/src/model"
)

func TestToJsonFromObject(t *testing.T) {
	account := givenAccount(true)
	actual := ToJsonFromObject(account)
	AssertNotNil(t, actual)
}

func TestFormatJSON(t *testing.T) {
	account := givenAccount(true)
	accountJSON := ToJsonFromObject(account)
	actual := FormatJSON(accountJSON)
	AssertNotNil(t, actual)
}

func TestFormatJSONError(t *testing.T) {
	actual := FormatJSON("invalid")
	AssertNotNil(t, actual)
}

func TestReadJsonFileToObject(t *testing.T) {
	json := ReadJsonFromFile("../../resources/fixtures/default.json")
	AssertNotNil(t, json)
	AssertContains(t, json, "Release 1")
}

// Fixtures

func givenAccount(active bool) model.Account {
	return model.Account{
		Repository: "switcherapi/switcher-gitops",
		Branch:     "master",
		Domain: model.DomainDetails{
			ID:         "123",
			Name:       "Switcher GitOps",
			Version:    "123",
			LastCommit: "123",
			Status:     model.StatusSynced,
			Message:    "Synced successfully",
		},
		Settings: model.Settings{
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

func AssertNotNil(t *testing.T, object interface{}) {
	if object == nil {
		t.Errorf("Object is nil")
	}
}

func AssertNil(t *testing.T, object interface{}) {
	if object != nil {
		t.Errorf("Object is not nil")
	}
}

func AssertEqual(t *testing.T, actual interface{}, expected interface{}) {
	if actual != expected {
		t.Errorf("Expected %v, got %v", actual, expected)
	}
}

func AssertContains(t *testing.T, actual string, expected string) {
	if !strings.Contains(actual, expected) {
		t.Errorf("Expected %v to contain %v", actual, expected)
	}
}
