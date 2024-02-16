package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func ResponseJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	encodedData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return
	}

	jsonString := string(encodedData)
	w.Write([]byte(jsonString))
}
