package utils

import (
	"bytes"
	"encoding/json"
	"os"
)

func FormatJSON(jsonString string) string {
	var prettyJSON bytes.Buffer
	error := json.Indent(&prettyJSON, []byte(jsonString), "", "  ")
	if error != nil {
		return jsonString
	}
	return string(prettyJSON.String())
}

func ReadJsonFromFile(path string) string {
	file, _ := os.Open(path)
	defer file.Close()

	stat, _ := file.Stat()
	bs := make([]byte, stat.Size())
	file.Read(bs)
	return string(bs)
}

func ToJsonFromObject(object interface{}) string {
	json, _ := json.MarshalIndent(object, "", "  ")
	return string(json)
}
