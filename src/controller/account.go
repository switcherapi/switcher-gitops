package controller

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/switcherapi/switcher-gitops/src/config"
	"github.com/switcherapi/switcher-gitops/src/core"
	"github.com/switcherapi/switcher-gitops/src/model"
	"github.com/switcherapi/switcher-gitops/src/repository"
	"github.com/switcherapi/switcher-gitops/src/utils"
)

type AccountController struct {
	coreHandler       *core.CoreHandler
	accountRepository repository.AccountRepository
	routeAccountPath  string
}

type AccountTokensRequest struct {
	Token        string   `json:"token"`
	DomainId     string   `json:"domainId"`
	Environments []string `json:"environments"`
}

type AccountTokensResponse struct {
	Result  bool   `json:"result"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewAccountController(repo repository.AccountRepository, coreHandler *core.CoreHandler) *AccountController {
	return &AccountController{
		coreHandler:       coreHandler,
		accountRepository: repo,
		routeAccountPath:  "/account",
	}
}

func (controller *AccountController) RegisterRoutes(r *mux.Router) http.Handler {
	r.NewRoute().Path(controller.routeAccountPath).Name("CreateAccount").Handler(
		ValidateToken(http.HandlerFunc(controller.CreateAccountHandler))).Methods(http.MethodPost)
	r.NewRoute().Path(controller.routeAccountPath).Name("UpdateAccount").Handler(
		ValidateToken(http.HandlerFunc(controller.UpdateAccountHandler))).Methods(http.MethodPut)
	r.NewRoute().Path(controller.routeAccountPath + "/{domainId}").Name("UpdateAccountTokens").Handler(
		ValidateToken(http.HandlerFunc(controller.UpdateAccountTokensHandler))).Methods(http.MethodPut)
	r.NewRoute().Path(controller.routeAccountPath + "/{domainId}").Name("GelAllAccountsByDomainId").Handler(
		ValidateToken(http.HandlerFunc(controller.FetchAllAccountsByDomainIdHandler))).Methods(http.MethodGet)
	r.NewRoute().Path(controller.routeAccountPath + "/{domainId}/{enviroment}").Name("GetAccount").Handler(
		ValidateToken(http.HandlerFunc(controller.FetchAccountHandler))).Methods(http.MethodGet)
	r.NewRoute().Path(controller.routeAccountPath + "/{domainId}/{enviroment}").Name("DeleteAccount").Handler(
		ValidateToken(http.HandlerFunc(controller.DeleteAccountHandler))).Methods(http.MethodDelete)

	return r
}

func (controller *AccountController) CreateAccountHandler(w http.ResponseWriter, r *http.Request) {
	var accountRequest model.Account
	err := json.NewDecoder(r.Body).Decode(&accountRequest)
	if err != nil {
		utils.ResponseJSON(w, ErrorResponse{Error: "Invalid request"}, http.StatusBadRequest)
		return
	}

	// Encrypt token before saving
	if accountRequest.Token != "" {
		accountRequest.Token = utils.Encrypt(accountRequest.Token, config.GetEnv("GIT_TOKEN_PRIVATE_KEY"))
	}

	accountCreated, err := controller.accountRepository.Create(&accountRequest)
	if err != nil {
		utils.LogError("Error creating account: %s", err.Error())
		utils.ResponseJSON(w, ErrorResponse{Error: "Error creating account"}, http.StatusInternalServerError)
		return
	}

	// Initialize account handler
	gitService := core.NewGitService(accountCreated.Repository, accountCreated.Token, accountCreated.Branch)
	go controller.coreHandler.StartAccountHandler(accountCreated.ID.Hex(), gitService)

	opaqueTokenFromResponse(accountCreated)
	utils.ResponseJSON(w, accountCreated, http.StatusCreated)
}

func (controller *AccountController) FetchAccountHandler(w http.ResponseWriter, r *http.Request) {
	domainId := mux.Vars(r)["domainId"]
	enviroment := mux.Vars(r)["enviroment"]

	account, err := controller.accountRepository.FetchByDomainIdEnvironment(domainId, enviroment)
	if err != nil {
		utils.LogError("Error fetching account: %s", err.Error())
		utils.ResponseJSON(w, ErrorResponse{Error: "Account not found"}, http.StatusNotFound)
		return
	}

	opaqueTokenFromResponse(account)
	utils.ResponseJSON(w, account, http.StatusOK)
}

func (controller *AccountController) FetchAllAccountsByDomainIdHandler(w http.ResponseWriter, r *http.Request) {
	domainId := mux.Vars(r)["domainId"]

	accounts := controller.accountRepository.FetchAllByDomainId(domainId)
	if accounts == nil {
		utils.LogError("Not found accounts for domain: %s", domainId)
		utils.ResponseJSON(w, ErrorResponse{Error: "Not found accounts for domain: " + domainId}, http.StatusNotFound)
		return
	}

	var accountsResponse []model.Account
	for _, account := range accounts {
		opaqueTokenFromResponse(&account)
		accountsResponse = append(accountsResponse, account)
	}

	utils.ResponseJSON(w, accountsResponse, http.StatusOK)
}

func (controller *AccountController) UpdateAccountHandler(w http.ResponseWriter, r *http.Request) {
	var accountRequest model.Account
	err := json.NewDecoder(r.Body).Decode(&accountRequest)
	if err != nil {
		utils.LogError("Error updating account: %s", err.Error())
		utils.ResponseJSON(w, ErrorResponse{Error: "Invalid request"}, http.StatusBadRequest)
		return
	}

	// Encrypt token before saving
	if accountRequest.Token != "" {
		accountRequest.Token = utils.Encrypt(accountRequest.Token, config.GetEnv("GIT_TOKEN_PRIVATE_KEY"))
	}

	accountUpdated, err := controller.accountRepository.Update(&accountRequest)
	if err != nil {
		utils.LogError("Error updating account: %s", err.Error())
		utils.ResponseJSON(w, ErrorResponse{Error: "Error updating account"}, http.StatusInternalServerError)
		return
	}

	opaqueTokenFromResponse(accountUpdated)
	utils.ResponseJSON(w, accountUpdated, http.StatusOK)
}

func (controller *AccountController) UpdateAccountTokensHandler(w http.ResponseWriter, r *http.Request) {
	var accountTokensRequest AccountTokensRequest
	err := json.NewDecoder(r.Body).Decode(&accountTokensRequest)
	if err != nil {
		utils.LogError("Error updating account tokens: %s", err.Error())
		utils.ResponseJSON(w, ErrorResponse{Error: "Invalid request"}, http.StatusBadRequest)
		return
	}

	if accountTokensRequest.Token == "" {
		utils.LogError("Error updating account tokens: Token is required")
		utils.ResponseJSON(w, ErrorResponse{Error: "Token is required"}, http.StatusBadRequest)
		return
	}

	// Encrypt token before saving
	accountTokensRequest.Token = utils.Encrypt(accountTokensRequest.Token, config.GetEnv("GIT_TOKEN_PRIVATE_KEY"))

	// Update account tokens
	for _, environment := range accountTokensRequest.Environments {
		account, err := controller.accountRepository.FetchByDomainIdEnvironment(accountTokensRequest.DomainId, environment)
		if err != nil {
			utils.LogError("Error fetching account: %s", err.Error())
			utils.ResponseJSON(w, ErrorResponse{Error: "Error fetching account"}, http.StatusNotFound)
			return
		}

		account.Token = accountTokensRequest.Token
		controller.accountRepository.Update(account)
	}

	utils.ResponseJSON(w, AccountTokensResponse{
		Result:  true,
		Message: "Account tokens updated successfully",
	}, http.StatusOK)
}

func (controller *AccountController) DeleteAccountHandler(w http.ResponseWriter, r *http.Request) {
	domainId := mux.Vars(r)["domainId"]
	enviroment := mux.Vars(r)["enviroment"]

	err := controller.accountRepository.DeleteByDomainIdEnvironment(domainId, enviroment)
	if err != nil {
		utils.LogError("Error deleting account: %s", err.Error())
		utils.ResponseJSON(w, ErrorResponse{Error: "Error deleting account: " + err.Error()}, http.StatusInternalServerError)
		return
	}

	utils.ResponseJSON(w, nil, http.StatusNoContent)
}

func opaqueTokenFromResponse(account *model.Account) {
	if account.Token != "" {
		account.Token = "..." + account.Token[len(account.Token)-4:]
	}
}
