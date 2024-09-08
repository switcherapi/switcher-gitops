package controller

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/switcherapi/switcher-gitops/src/core"
	"github.com/switcherapi/switcher-gitops/src/model"
	"github.com/switcherapi/switcher-gitops/src/repository"
	"github.com/switcherapi/switcher-gitops/src/utils"
)

type AccountController struct {
	CoreHandler       *core.CoreHandler
	AccountRepository repository.AccountRepository
	RouteAccountPath  string
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewAccountController(repo repository.AccountRepository, coreHandler *core.CoreHandler) *AccountController {
	return &AccountController{
		CoreHandler:       coreHandler,
		AccountRepository: repo,
		RouteAccountPath:  "/account",
	}
}

func (controller *AccountController) RegisterRoutes(r *mux.Router) http.Handler {
	const routesDomainVar = "/{domainId}"

	r.NewRoute().Path(controller.RouteAccountPath + routesDomainVar).Name("GetAccount").HandlerFunc(controller.FetchAccountHandler).Methods(http.MethodGet)
	r.NewRoute().Path(controller.RouteAccountPath).Name("CreateAccount").HandlerFunc(controller.CreateAccountHandler).Methods(http.MethodPost)
	r.NewRoute().Path(controller.RouteAccountPath + routesDomainVar).Name("UpdateAccount").HandlerFunc(controller.UpdateAccountHandler).Methods(http.MethodPut)
	r.NewRoute().Path(controller.RouteAccountPath + routesDomainVar).Name("DeleteAccount").HandlerFunc(controller.DeleteAccountHandler).Methods(http.MethodDelete)

	return r
}

func (controller *AccountController) CreateAccountHandler(w http.ResponseWriter, r *http.Request) {
	var accountRequest model.Account
	err := json.NewDecoder(r.Body).Decode(&accountRequest)
	if err != nil {
		utils.ResponseJSON(w, ErrorResponse{Error: "Invalid request"}, http.StatusBadRequest)
		return
	}

	accountCreated, err := controller.AccountRepository.Create(&accountRequest)
	if err != nil {
		utils.Log(utils.LogLevelError, "Error creating account: %s", err.Error())
		utils.ResponseJSON(w, ErrorResponse{Error: "Error creating account"}, http.StatusInternalServerError)
		return
	}

	// Initialize account handler
	gitService := core.NewGitService(accountCreated.Repository, accountCreated.Token, accountCreated.Branch)
	go controller.CoreHandler.StartAccountHandler(accountCreated.ID.Hex(), gitService)

	utils.ResponseJSON(w, accountCreated, http.StatusCreated)
}

func (controller *AccountController) FetchAccountHandler(w http.ResponseWriter, r *http.Request) {
	domainId := r.URL.Path[len(controller.RouteAccountPath+"/"):]
	account, err := controller.AccountRepository.FetchByDomainId(domainId)
	if err != nil {
		utils.Log(utils.LogLevelError, "Error fetching account: %s", err.Error())
		utils.ResponseJSON(w, ErrorResponse{Error: "Account not found"}, http.StatusNotFound)
		return
	}

	utils.ResponseJSON(w, account, http.StatusOK)
}

func (controller *AccountController) UpdateAccountHandler(w http.ResponseWriter, r *http.Request) {
	var accountRequest model.Account
	err := json.NewDecoder(r.Body).Decode(&accountRequest)
	if err != nil {
		utils.Log(utils.LogLevelError, "Error updating account: %s", err.Error())
		utils.ResponseJSON(w, ErrorResponse{Error: "Invalid request"}, http.StatusBadRequest)
		return
	}

	accountUpdated, err := controller.AccountRepository.Update(&accountRequest)
	if err != nil {
		utils.Log(utils.LogLevelError, "Error updating account: %s", err.Error())
		utils.ResponseJSON(w, ErrorResponse{Error: "Error updating account"}, http.StatusInternalServerError)
		return
	}

	utils.ResponseJSON(w, accountUpdated, http.StatusOK)
}

func (controller *AccountController) DeleteAccountHandler(w http.ResponseWriter, r *http.Request) {
	domainId := r.URL.Path[len(controller.RouteAccountPath+"/"):]
	err := controller.AccountRepository.DeleteByDomainId(domainId)
	if err != nil {
		utils.Log(utils.LogLevelError, "Error deleting account: %s", err.Error())
		utils.ResponseJSON(w, ErrorResponse{Error: "Error deleting account: " + err.Error()}, http.StatusInternalServerError)
		return
	}

	utils.ResponseJSON(w, nil, http.StatusNoContent)
}
