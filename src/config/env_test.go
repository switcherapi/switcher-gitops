package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEnv(t *testing.T) {
	os.Setenv("GO_ENV", "test")
	InitEnv()

	switcherApiUrl := GetEnv("SWITCHER_API_URL")
	assert.NotEmpty(t, switcherApiUrl)
}
