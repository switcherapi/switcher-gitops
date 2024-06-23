package core

import (
	"sync"
	"time"

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

	// Iterate over accounts and start account handlers
	for _, account := range accounts {
		go c.StartAccountHandler(account, make(chan bool), &sync.WaitGroup{})
	}

	// Update core handler status
	c.status = 1
	return c.status, nil
}

func (c *CoreHandler) StartAccountHandler(account model.Account, quit chan bool, wg *sync.WaitGroup) {
	defer wg.Done()

	// Check if account is not active
	if !account.Settings.Active {
		return
	}

	// Reads Window setting
	sleep := 1

	for {
		select {
		case <-quit:
			return
		default:
			lastCommit := "123"
			date := time.Now().Format("2006-01-02 15:04:05")

			if account.Domain.LastCommit == "" || account.Domain.LastCommit != lastCommit {
				c.CheckForChanges(account, lastCommit, date)
			}
		}

		time.Sleep(time.Duration(sleep) * time.Second)
	}
}

func (c *CoreHandler) CheckForChanges(account model.Account, lastCommit string, date string) {
	account.Domain.LastCommit = lastCommit
	account.Domain.Status = "Synced"
	account.Domain.Message = "Synced successfully"
	c.AccountRepository.Update(&account)
}
