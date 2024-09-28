package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/switcherapi/switcher-gitops/src/config"
	"github.com/switcherapi/switcher-gitops/src/model"
)

type GraphQLRequest struct {
	Query string `json:"query"`
}

type PushChangeResponse struct {
	Message string `json:"message"`
	Error   string `json:"error"`
	Version int    `json:"version"`
}

type IAPIService interface {
	FetchSnapshotVersion(domainId string, environment string) (string, error)
	FetchSnapshot(domainId string, environment string) (string, error)
	PushChanges(domainId string, diff model.DiffResult) (PushChangeResponse, error)
	NewDataFromJson(jsonData []byte) model.Data
}

type ApiService struct {
	apiKey string
	apiUrl string
}

func NewApiService(apiKey string, apiUrl string) *ApiService {
	return &ApiService{
		apiKey: apiKey,
		apiUrl: apiUrl,
	}
}

func (c *ApiService) NewDataFromJson(jsonData []byte) model.Data {
	var data model.Data
	json.Unmarshal(jsonData, &data)
	return data
}

func (a *ApiService) FetchSnapshotVersion(domainId string, environment string) (string, error) {
	query := createQuerySnapshotVersion(domainId)
	responseBody, err := a.doGraphQLRequest(domainId, query)

	if err != nil {
		return "", err
	}

	return responseBody, nil
}

func (a *ApiService) FetchSnapshot(domainId string, environment string) (string, error) {
	query := createQuery(domainId, environment)
	responseBody, err := a.doGraphQLRequest(domainId, query)

	if err != nil {
		return "", err
	}

	return responseBody, nil
}

func (a *ApiService) PushChanges(domainId string, diff model.DiffResult) (PushChangeResponse, error) {
	reqBody, _ := json.Marshal(diff)
	responseBody, status, err := a.doPostRequest(a.apiUrl+config.GetEnv("SWITCHER_PATH_PUSH"), domainId, reqBody)

	if err != nil {
		return PushChangeResponse{}, err
	}

	var response PushChangeResponse
	json.Unmarshal([]byte(responseBody), &response)

	if status != http.StatusOK {
		return PushChangeResponse{}, errors.New(response.Error)
	}

	return response, nil
}

func (a *ApiService) doGraphQLRequest(domainId string, query string) (string, error) {
	// Generate a bearer token
	token := generateBearerToken(a.apiKey, domainId)

	// Create a new request
	reqBody, _ := json.Marshal(GraphQLRequest{Query: query})
	req, _ := http.NewRequest("POST", a.apiUrl+config.GetEnv("SWITCHER_PATH_GRAPHQL"), bytes.NewBuffer(reqBody))

	// Set the request headers
	setHeaders(req, token)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	responseBody, _ := io.ReadAll(resp.Body)
	return string(responseBody), nil
}

func (a *ApiService) doPostRequest(url string, domainId string, body []byte) (string, int, error) {
	// Generate a bearer token
	token := generateBearerToken(a.apiKey, domainId)

	// Create a new request
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))

	// Set the request headers
	setHeaders(req, token)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	responseBody, _ := io.ReadAll(resp.Body)
	return string(responseBody), resp.StatusCode, nil
}

func generateBearerToken(apiKey string, subject string) string {
	// Define the claims for the JWT token
	claims := jwt.MapClaims{
		"iss":     "GitOps Service",
		"sub":     "/resource",
		"subject": subject,
		"exp":     time.Now().Add(time.Minute).Unix(),
	}

	// Create the JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the API key
	signedToken, _ := token.SignedString([]byte(apiKey))

	return signedToken
}

func setHeaders(req *http.Request, token string) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
}

func createQuerySnapshotVersion(domainId string) string {
	return fmt.Sprintf(`
    {
        domain(_id: "%s") {
            version
        }
    }`, domainId)
}

func createQuery(domainId string, environment string) string {
	return fmt.Sprintf(`
    {
        domain(_id: "%s", environment: "%s") {
            name
            version
            group {
                name
                description
                activated
                config {
                    key
                    description
                    activated
                    strategies {
                        strategy
                        activated
                        operation
                        values
                    }
                    components
                }
            }
        }
    }`, domainId, environment)
}
