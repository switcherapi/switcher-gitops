package utils

import (
	"testing"
)

func TestLog(t *testing.T) {
	LogDebug("This is a debug message")
	LogInfo("This is an info message")
	LogError("This is an error message")
}
