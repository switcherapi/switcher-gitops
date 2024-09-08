package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/switcherapi/switcher-gitops/src/config"
)

const (
	LogLevelInfo  = "INFO"
	LogLevelError = "ERROR"
	LogLevelDebug = "DEBUG"
)

func Log(logLevel string, message string, args ...interface{}) {
	currentLogLevel := config.GetEnv("LOG_LEVEL")

	if currentLogLevel == "DEBUG" || currentLogLevel == "ERROR" || currentLogLevel == logLevel {
		log.Printf("[%s] %s\n", logLevel, fmt.Sprintf(message, args...))
	}
}

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

func ToMapFromObject(obj interface{}) map[string]interface{} {
	var result map[string]interface{}
	jsonData, _ := json.Marshal(obj)
	json.Unmarshal(jsonData, &result)
	return result
}

func Encrypt(plaintext string, privateKey string) string {
	aes, _ := aes.NewCipher([]byte(privateKey))
	gcm, _ := cipher.NewGCM(aes)

	nonce := make([]byte, gcm.NonceSize())
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	return base64.StdEncoding.EncodeToString(ciphertext)
}

func Decrypt(encodedPlaintext string, privateKey string) (string, error) {
	decodedText, _ := base64.StdEncoding.DecodeString(encodedPlaintext)

	aes, _ := aes.NewCipher([]byte(privateKey))
	gcm, _ := cipher.NewGCM(aes)

	nonceSize := gcm.NonceSize()
	if len(decodedText) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := decodedText[:nonceSize], decodedText[nonceSize:]
	plaintext, _ := gcm.Open(nil, []byte(nonce), []byte(ciphertext), nil)

	return string(plaintext), nil
}
