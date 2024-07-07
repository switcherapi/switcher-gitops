package core

import (
	"sync"
	"time"

	"github.com/switcherapi/switcher-gitops/src/model"
	"github.com/switcherapi/switcher-gitops/src/repository"
)

type CoreHandler struct {
	AccountRepository repository.AccountRepository
	GitService        IGitService
	status            int
}

func NewCoreHandler(repo repository.AccountRepository, gitService IGitService) *CoreHandler {
	return &CoreHandler{
		AccountRepository: repo,
		GitService:        gitService,
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
			lastCommit, date, content, err := c.GitService.GetRepositoryData(account.Environment)

			if err == nil && isRepositoryOutSync(account, lastCommit) {
				c.syncUp(account, lastCommit, date, content)
			}
		}

		time.Sleep(time.Duration(sleep) * time.Second)
	}
}

func (c *CoreHandler) syncUp(account model.Account, lastCommit string, date string, content string) {
	status, message := c.GitService.CheckForChanges(account, lastCommit, date, content)

	account.Domain.LastCommit = lastCommit
	account.Domain.Status = status
	account.Domain.Message = message
	c.AccountRepository.Update(&account)
}

func isRepositoryOutSync(account model.Account, lastCommit string) bool {
	return account.Domain.LastCommit == "" || account.Domain.LastCommit != lastCommit
}
