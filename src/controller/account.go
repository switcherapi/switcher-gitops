package controller

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/switcherapi/switcher-gitops/src/model"
	"github.com/switcherapi/switcher-gitops/src/repository"
	"github.com/switcherapi/switcher-gitops/src/utils"
)

type AccountController struct {
	AccountRepository repository.AccountRepository
	RouteAccountPath  string
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewAccountController(repo repository.AccountRepository) *AccountController {
	return &AccountController{
		AccountRepository: repo,
		RouteAccountPath:  "/account",
	}
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
		log.Println(err)
		utils.ResponseJSON(w, ErrorResponse{Error: "Error creating account"}, http.StatusInternalServerError)
		return
	}

	utils.ResponseJSON(w, accountCreated, http.StatusCreated)
}

func (controller *AccountController) FetchAccountHandler(w http.ResponseWriter, r *http.Request) {
	domainId := r.URL.Path[len(controller.RouteAccountPath+"/"):]
	account, err := controller.AccountRepository.FetchByDomainId(domainId)
	if err != nil {
		log.Println(err)
		utils.ResponseJSON(w, ErrorResponse{Error: "Account not found"}, http.StatusNotFound)
		return
	}

	utils.ResponseJSON(w, account, http.StatusOK)
}

func (controller *AccountController) UpdateAccountHandler(w http.ResponseWriter, r *http.Request) {
	var accountRequest model.Account
	err := json.NewDecoder(r.Body).Decode(&accountRequest)
	if err != nil {
		log.Println(err)
		utils.ResponseJSON(w, ErrorResponse{Error: "Invalid request"}, http.StatusBadRequest)
		return
	}

	accountUpdated, err := controller.AccountRepository.Update(&accountRequest)
	if err != nil {
		log.Println(err)
		utils.ResponseJSON(w, ErrorResponse{Error: "Error updating account"}, http.StatusInternalServerError)
		return
	}

	utils.ResponseJSON(w, accountUpdated, http.StatusOK)
}
