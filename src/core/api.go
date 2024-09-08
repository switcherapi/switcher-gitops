package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/switcherapi/switcher-gitops/src/model"
)

type GraphQLRequest struct {
	Query string `json:"query"`
}

type ApplyChangeResponse struct {
	Message string `json:"message"`
	Version int    `json:"version"`
}

type IAPIService interface {
	FetchSnapshot(domainId string, environment string) (string, error)
	ApplyChangesToAPI(domainId string, environment string, diff model.DiffResult) (ApplyChangeResponse, error)
	NewDataFromJson(jsonData []byte) model.Data
}

type ApiService struct {
	ApiKey string
	ApiUrl string
}

func NewApiService(apiKey string, apiUrl string) *ApiService {
	return &ApiService{
		ApiKey: apiKey,
		ApiUrl: apiUrl,
	}
}

func (c *ApiService) NewDataFromJson(jsonData []byte) model.Data {
	var data model.Data
	json.Unmarshal(jsonData, &data)
	return data
}

func (a *ApiService) FetchSnapshot(domainId string, environment string) (string, error) {
	// Generate a bearer token
	token := generateBearerToken(a.ApiKey, domainId)

	// Define the GraphQL query
	query := createQuery(domainId, environment)

	// Create a new request
	reqBody, _ := json.Marshal(GraphQLRequest{Query: query})
	req, _ := http.NewRequest("POST", a.ApiUrl+"/gitops-graphql", bytes.NewBuffer(reqBody))

	// Set the request headers
	setHeaders(req, token)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read and print the response
	body, _ := io.ReadAll(resp.Body)
	return string(body), nil
}

func (a *ApiService) ApplyChangesToAPI(domainId string, environment string, diff model.DiffResult) (ApplyChangeResponse, error) {
	// Generate a bearer token
	token := generateBearerToken(a.ApiKey, domainId)

	// Create a new request
	reqBody, _ := json.Marshal(diff)
	req, _ := http.NewRequest("POST", a.ApiUrl+"/gitops/apply", bytes.NewBuffer(reqBody))

	// Set the request headers
	setHeaders(req, token)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return ApplyChangeResponse{}, err
	}
	defer resp.Body.Close()

	// Read and print the response
	body, _ := io.ReadAll(resp.Body)
	var response ApplyChangeResponse
	json.Unmarshal(body, &response)
	return response, nil
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
