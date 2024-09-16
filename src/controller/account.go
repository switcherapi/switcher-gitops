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
	r.NewRoute().Path(controller.routeAccountPath).Name("CreateAccount").HandlerFunc(controller.CreateAccountHandler).Methods(http.MethodPost)
	r.NewRoute().Path(controller.routeAccountPath).Name("UpdateAccount").HandlerFunc(controller.UpdateAccountHandler).Methods(http.MethodPut)
	r.NewRoute().Path(controller.routeAccountPath + "/{domainId}").Name("GelAllAccountsByDomainId").HandlerFunc(controller.FetchAllAccountsByDomainIdHandler).Methods(http.MethodGet)
	r.NewRoute().Path(controller.routeAccountPath + "/{domainId}/{enviroment}").Name("GetAccount").HandlerFunc(controller.FetchAccountHandler).Methods(http.MethodGet)
	r.NewRoute().Path(controller.routeAccountPath + "/{domainId}/{enviroment}").Name("DeleteAccount").HandlerFunc(controller.DeleteAccountHandler).Methods(http.MethodDelete)

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
		utils.Log(utils.LogLevelError, "Error creating account: %s", err.Error())
		utils.ResponseJSON(w, ErrorResponse{Error: "Error creating account"}, http.StatusInternalServerError)
		return
	}

	// Initialize account handler
	gitService := core.NewGitService(accountCreated.Repository, accountCreated.Token, accountCreated.Branch)
	go controller.coreHandler.StartAccountHandler(accountCreated.ID.Hex(), gitService)

	utils.ResponseJSON(w, accountCreated, http.StatusCreated)
}

func (controller *AccountController) FetchAccountHandler(w http.ResponseWriter, r *http.Request) {
	domainId := mux.Vars(r)["domainId"]
	enviroment := mux.Vars(r)["enviroment"]

	account, err := controller.accountRepository.FetchByDomainIdEnvironment(domainId, enviroment)
	if err != nil {
		utils.Log(utils.LogLevelError, "Error fetching account: %s", err.Error())
		utils.ResponseJSON(w, ErrorResponse{Error: "Account not found"}, http.StatusNotFound)
		return
	}

	utils.ResponseJSON(w, account, http.StatusOK)
}

func (controller *AccountController) FetchAllAccountsByDomainIdHandler(w http.ResponseWriter, r *http.Request) {
	domainId := mux.Vars(r)["domainId"]

	accounts := controller.accountRepository.FetchAllByDomainId(domainId)
	if accounts == nil {
		utils.Log(utils.LogLevelError, "Not found accounts for domain: %s", domainId)
		utils.ResponseJSON(w, ErrorResponse{Error: "Not found accounts for domain: " + domainId}, http.StatusNotFound)
		return
	}

	utils.ResponseJSON(w, accounts, http.StatusOK)
}

func (controller *AccountController) UpdateAccountHandler(w http.ResponseWriter, r *http.Request) {
	var accountRequest model.Account
	err := json.NewDecoder(r.Body).Decode(&accountRequest)
	if err != nil {
		utils.Log(utils.LogLevelError, "Error updating account: %s", err.Error())
		utils.ResponseJSON(w, ErrorResponse{Error: "Invalid request"}, http.StatusBadRequest)
		return
	}

	// Encrypt token before saving
	if accountRequest.Token != "" {
		accountRequest.Token = utils.Encrypt(accountRequest.Token, config.GetEnv("GIT_TOKEN_PRIVATE_KEY"))
	}

	accountUpdated, err := controller.accountRepository.Update(&accountRequest)
	if err != nil {
		utils.Log(utils.LogLevelError, "Error updating account: %s", err.Error())
		utils.ResponseJSON(w, ErrorResponse{Error: "Error updating account"}, http.StatusInternalServerError)
		return
	}

	utils.ResponseJSON(w, accountUpdated, http.StatusOK)
}

func (controller *AccountController) DeleteAccountHandler(w http.ResponseWriter, r *http.Request) {
	domainId := mux.Vars(r)["domainId"]
	enviroment := mux.Vars(r)["enviroment"]

	err := controller.accountRepository.DeleteByDomainIdEnvironment(domainId, enviroment)
	if err != nil {
		utils.Log(utils.LogLevelError, "Error deleting account: %s", err.Error())
		utils.ResponseJSON(w, ErrorResponse{Error: "Error deleting account: " + err.Error()}, http.StatusInternalServerError)
		return
	}

	utils.ResponseJSON(w, nil, http.StatusNoContent)
}
