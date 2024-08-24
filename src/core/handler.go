package core

import (
	"sync"
	"time"

	"github.com/switcherapi/switcher-gitops/src/model"
	"github.com/switcherapi/switcher-gitops/src/repository"
	"github.com/switcherapi/switcher-gitops/src/utils"
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
	c.updateDomainStatus(account, model.StatusOutSync, "Syncing up...")

	// Check for changes
	diff, snapshotApi, err := c.checkForChanges(account.Domain.ID, account.Environment, repositoryData.Content)

	if err != nil {
		c.updateDomainStatus(account, model.StatusError, "Failed to check for changes")
		return
	}

	if snapshotApi.Domain.Version > account.Domain.Version {
		account = c.applyChangesToRepository(account, snapshotApi)
	} else if len(diff.Changes) > 0 {
		account = c.applyChangesToAPI(account, repositoryData)
	}

	// Update account status: Synced
	c.updateDomainStatus(account, model.StatusSynced, "Synced successfully")
}

func (c *CoreHandler) checkForChanges(domainId string, environment string, content string) (model.DiffResult, model.Snapshot, error) {
	// Get Snapshot from API
	snapshotJsonFromApi, err := c.ApiService.FetchSnapshot(domainId, environment)

	if err != nil {
		return model.DiffResult{}, model.Snapshot{}, err
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

	return c.ComparatorService.MergeResults([]model.DiffResult{diffNew, diffChanged, diffDeleted}), snapshotApi.Snapshot, nil
}

func (c *CoreHandler) applyChangesToAPI(account model.Account, repositoryData *model.RepositoryData) model.Account {
	// Push changes to API
	println("Pushing changes to API")

	// Update domain
	account.Domain.Version = "2"
	account.Domain.LastCommit = repositoryData.CommitHash

	return account
}

func (c *CoreHandler) applyChangesToRepository(account model.Account, snapshot model.Snapshot) model.Account {
	// Remove version from domain
	snapshotContent := snapshot
	snapshotContent.Domain.Version = ""

	lastCommit, _ := c.GitService.PushChanges(account.Environment, utils.ToJsonFromObject(snapshotContent))

	// Update domain
	account.Domain.Version = snapshot.Domain.Version
	account.Domain.LastCommit = lastCommit

	return account
}

func isRepositoryOutSync(account model.Account, lastCommit string) bool {
	return account.Domain.LastCommit == "" || account.Domain.LastCommit != lastCommit
}

func getTimeWindow(window string) (int, time.Duration) {
	duration, _ := time.ParseDuration(window)
	return 1, duration
}

func (c *CoreHandler) updateDomainStatus(account model.Account, status string, message string) {
	account.Domain.Status = status
	account.Domain.Message = message
	account.Domain.LastDate = time.Now().Format(time.ANSIC)
	c.AccountRepository.Update(&account)
}
