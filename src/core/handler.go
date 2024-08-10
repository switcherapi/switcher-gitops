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
	ApiService        IAPIService
	ComparatorService IComparatorService
	status            int
}

func NewCoreHandler(repo repository.AccountRepository, gitService IGitService, apiService IAPIService, comparatorService IComparatorService) *CoreHandler {
	return &CoreHandler{
		AccountRepository: repo,
		GitService:        gitService,
		ApiService:        apiService,
		ComparatorService: comparatorService,
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
	timeWindow, unitWindow := getTimeWindow(account.Settings.Window)

	for {
		select {
		case <-quit:
			return
		default:
			repositoryData, err := c.GitService.GetRepositoryData(account.Environment)

			if err == nil && isRepositoryOutSync(account, repositoryData.CommitHash) {
				c.syncUp(account, repositoryData)
			}
		}

		time.Sleep(time.Duration(timeWindow) * unitWindow)
	}
}

func (c *CoreHandler) syncUp(account model.Account, repositoryData *model.RepositoryData) {
	// Update account status: Out of sync
	account.Domain.LastCommit = repositoryData.CommitHash
	account.Domain.LastDate = repositoryData.CommitDate
	account.Domain.Status = model.StatusOutSync
	account.Domain.Message = "Syncing up..."
	c.AccountRepository.Update(&account)

	// Check for changes
	diff, err := c.checkForChanges(account.Domain.ID, account.Environment, repositoryData.Content)

	if err != nil {
		// Update account status: Error
		account.Domain.Status = model.StatusError
		account.Domain.Message = "Error syncing up"
		c.AccountRepository.Update(&account)
		return
	}

	// Apply changes
	c.applyChanges(account, diff)

	// Update account status: Synced
	account.Domain.Status = model.StatusSynced
	account.Domain.Message = "Synced successfully"
	c.AccountRepository.Update(&account)
}

func (c *CoreHandler) checkForChanges(domainId string, environment string, content string) (model.DiffResult, error) {
	// Get Snapshot from API
	snapshotJsonFromApi, err := c.ApiService.FetchSnapshot(domainId, environment)

	if err != nil {
		return model.DiffResult{}, err
	}

	// Convert API JSON to model.Snapshot
	snapshotApi := c.ApiService.NewDataFromJson([]byte(snapshotJsonFromApi))
	left := snapshotApi.Snapshot

	// Convert content to model.Snapshot
	jsonRight := []byte(content)
	right := c.ComparatorService.NewSnapshotFromJson(jsonRight)

	// Compare Snapshots and get diff
	diffNew := c.ComparatorService.CheckSnapshotDiff(left, right, NEW)
	diffChanged := c.ComparatorService.CheckSnapshotDiff(left, right, CHANGED)
	diffDeleted := c.ComparatorService.CheckSnapshotDiff(left, right, DELETED)

	return c.ComparatorService.MergeResults([]model.DiffResult{diffNew, diffChanged, diffDeleted}), nil
}

func (c *CoreHandler) applyChanges(account model.Account, diff model.DiffResult) {
	// Apply changes
}

func isRepositoryOutSync(account model.Account, lastCommit string) bool {
	return account.Domain.LastCommit == "" || account.Domain.LastCommit != lastCommit
}

func getTimeWindow(window string) (int, time.Duration) {
	// Convert window string to time.Duration
	duration, _ := time.ParseDuration(window)
	return 1, duration
}
