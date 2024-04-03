package core

import (
	"github.com/switcherapi/switcher-gitops/src/model"
	"github.com/switcherapi/switcher-gitops/src/repository"
)

type CoreHandler struct {
	AccountRepository repository.AccountRepository
	status            int
}

func NewCoreHandler(repo repository.AccountRepository) *CoreHandler {
	return &CoreHandler{
		AccountRepository: repo,
	}
}

func (c *CoreHandler) InitCoreHandlerCoroutine() (int, error) {
	// Load all accounts
	accounts, _ := c.AccountRepository.FetchAllActiveAccounts()

	// Iterate over accounts and start a new account handler
	for _, account := range accounts {
		c.startNewAccountHandler(account)
	}

	// Update core handler status
	c.status = 1
	return c.status, nil
}

func (c *CoreHandler) startNewAccountHandler(account model.Account) {
	c.AccountRepository.FetchByDomainId(account.Domain.ID)
}
