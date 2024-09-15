package utils

import (
	"encoding/json"
	"net/http"
)

func ResponseJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	encodedData, err := json.Marshal(data)
	if err != nil {
		Log(LogLevelError, "Error encoding JSON: %s", err.Error())
		return
	}

	jsonString := string(encodedData)
	w.Write([]byte(jsonString))
}
