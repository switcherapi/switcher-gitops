package utils

import (
	"bytes"
	"encoding/json"
)

func FormatJSON(jsonString string) string {
	var prettyJSON bytes.Buffer
	error := json.Indent(&prettyJSON, []byte(jsonString), "", "  ")
	if error != nil {
		return jsonString
	}
	return string(prettyJSON.String())
}

func ToJsonFromObject(object interface{}) string {
	json, _ := json.MarshalIndent(object, "", "  ")
	return string(json)
}
