package controller

import (
	"encoding/json"
	"net/http"

	"github.com/switcherapi/switcher-gitops/src/model"
	"github.com/switcherapi/switcher-gitops/src/repository"
	"github.com/switcherapi/switcher-gitops/src/utils"
)

type AccountController struct {
	AccountRepository repository.AccountRepository
}

func (controller *AccountController) CreateAccountHandler(w http.ResponseWriter, r *http.Request) {
	var accountRequest model.Account
	err := json.NewDecoder(r.Body).Decode(&accountRequest)
	if err != nil {
		utils.ResponseJSON(w, err.Error(), http.StatusBadRequest)
		return
	}

	accountCreated, err := controller.AccountRepository.Create(&accountRequest)
	if err != nil {
		utils.ResponseJSON(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.ResponseJSON(w, accountCreated, http.StatusCreated)
}
