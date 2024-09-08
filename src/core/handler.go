package core

import (
	"time"

	"github.com/switcherapi/switcher-gitops/src/model"
	"github.com/switcherapi/switcher-gitops/src/repository"
	"github.com/switcherapi/switcher-gitops/src/utils"
)

const (
	CoreHandlerStatusInit    = 0
	CoreHandlerStatusRunning = 1
)

type CoreHandler struct {
	AccountRepository repository.AccountRepository
	ApiService        IAPIService
	ComparatorService IComparatorService
	status            int
}

func NewCoreHandler(accountRepository repository.AccountRepository, apiService IAPIService, comparatorService IComparatorService) *CoreHandler {
	return &CoreHandler{
		AccountRepository: accountRepository,
		ApiService:        apiService,
		ComparatorService: comparatorService,
	}
}

func (c *CoreHandler) InitCoreHandlerCoroutine() (int, error) {
	// Check if core handler is already running
	if c.status == CoreHandlerStatusRunning {
		return c.status, nil
	}

	// Update core handler status
	c.status = CoreHandlerStatusInit

	// Load all accounts
	accounts, _ := c.AccountRepository.FetchAllActiveAccounts()

	// Iterate over accounts and start account handlers
	for _, account := range accounts {
		go c.StartAccountHandler(account.ID.Hex(), NewGitService(
			account.Repository,
			account.Token,
			account.Branch,
		))
	}

	// Update core handler status
	c.status = CoreHandlerStatusRunning
	return c.status, nil
}

func (c *CoreHandler) StartAccountHandler(accountId string, gitService IGitService) {
	for {
		// Refresh account settings
		account, _ := c.AccountRepository.FetchByAccountId(accountId)

		if account == nil {
			// Terminate the goroutine (account was deleted)
			return
		}

		// Wait for account to be active
		if !account.Settings.Active {
			c.updateDomainStatus(*account, model.StatusPending, "Account was deactivated")
			time.Sleep(1 * time.Minute)
			continue
		}

		// Refresh account repository settings
		gitService.UpdateRepositorySettings(account.Repository, account.Token, account.Branch)

		// Fetch repository data
		repositoryData, err := gitService.GetRepositoryData(account.Environment)

		// Check if repository is out of sync
		if err == nil && isRepositoryOutSync(*account, repositoryData.CommitHash) {
			c.syncUp(*account, repositoryData, gitService)
		}

		// Wait for the next cycle
		timeWindow, unitWindow := getTimeWindow(account.Settings.Window)
		time.Sleep(time.Duration(timeWindow) * unitWindow)
	}
}

func (c *CoreHandler) syncUp(account model.Account, repositoryData *model.RepositoryData, gitService IGitService) {
	// Update account status: Out of sync
	account.Domain.LastCommit = repositoryData.CommitHash
	account.Domain.LastDate = repositoryData.CommitDate
	c.updateDomainStatus(account, model.StatusOutSync, model.MessageSyncingUp)

	// Check for changes
	diff, snapshotApi, err := c.checkForChanges(account, repositoryData.Content)

	if err != nil {
		c.updateDomainStatus(account, model.StatusError, "Failed to check for changes - "+err.Error())
		return
	}

	// Apply changes
	changeSource := ""
	if snapshotApi.Domain.Version > account.Domain.Version {
		changeSource = "Repository"
		account, err = c.applyChangesToRepository(account, snapshotApi, gitService)
	} else if len(diff.Changes) > 0 {
		changeSource = "API"
		account = c.applyChangesToAPI(account, repositoryData)
	}

	if err != nil {
		c.updateDomainStatus(account, model.StatusError, "Failed to apply changes ["+changeSource+"] - "+err.Error())
		return
	}

	// Update account status: Synced
	c.updateDomainStatus(account, model.StatusSynced, model.MessageSynced)
}

func (c *CoreHandler) checkForChanges(account model.Account, content string) (model.DiffResult, model.Snapshot, error) {
	// Get Snapshot from API
	snapshotJsonFromApi, err := c.ApiService.FetchSnapshot(account.Domain.ID, account.Environment)

	if err != nil {
		return model.DiffResult{}, model.Snapshot{}, err
	}

	// Convert API JSON to model.Snapshot
	snapshotApi := c.ApiService.NewDataFromJson([]byte(snapshotJsonFromApi))
	fromApi := snapshotApi.Snapshot

	// Convert content to model.Snapshot
	jsonRight := []byte(content)
	fromRepo := c.ComparatorService.NewSnapshotFromJson(jsonRight)

	// Compare Snapshots and get diff
	diffNew := c.ComparatorService.CheckSnapshotDiff(fromRepo, fromApi, NEW)
	diffChanged := c.ComparatorService.CheckSnapshotDiff(fromApi, fromRepo, CHANGED)

	var diffDeleted model.DiffResult
	if account.Settings.ForcePrune {
		diffDeleted = c.ComparatorService.CheckSnapshotDiff(fromApi, fromRepo, DELETED)
	}

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

func (c *CoreHandler) applyChangesToRepository(account model.Account, snapshot model.Snapshot, gitService IGitService) (model.Account, error) {
	// Remove version from domain
	snapshotContent := snapshot
	snapshotContent.Domain.Version = ""

	lastCommit, err := gitService.PushChanges(account.Environment, utils.ToJsonFromObject(snapshotContent))

	// Update domain
	account.Domain.Version = snapshot.Domain.Version
	account.Domain.LastCommit = lastCommit

	return account, err
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
