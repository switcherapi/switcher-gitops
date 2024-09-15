package core

import (
	"fmt"
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
	Status            int
}

func NewCoreHandler(accountRepository repository.AccountRepository, apiService IAPIService, comparatorService IComparatorService) *CoreHandler {
	return &CoreHandler{
		AccountRepository: accountRepository,
		ApiService:        apiService,
		ComparatorService: comparatorService,
	}
}

func (c *CoreHandler) InitCoreHandlerGoroutine() (int, error) {
	// Check if core handler is already running
	if c.Status == CoreHandlerStatusRunning {
		return c.Status, nil
	}

	// Update core handler status
	c.Status = CoreHandlerStatusInit

	// Load all accounts
	accounts := c.AccountRepository.FetchAllActiveAccounts()

	// Iterate over accounts and start account handlers
	for _, account := range accounts {
		go c.StartAccountHandler(account.ID.Hex(), NewGitService(
			account.Repository,
			account.Token,
			account.Branch,
		))
	}

	// Update core handler status
	c.Status = CoreHandlerStatusRunning
	return c.Status, nil
}

func (c *CoreHandler) StartAccountHandler(accountId string, gitService IGitService) {
	utils.Log(utils.LogLevelInfo, "[%s] Starting account handler", accountId)

	for {
		// Refresh account settings
		account, _ := c.AccountRepository.FetchByAccountId(accountId)

		if account == nil {
			// Terminate the goroutine (account was deleted)
			utils.Log(utils.LogLevelInfo, "[%s] Account was deleted, terminating account handler", accountId)
			return
		}

		// Wait for account to be active
		if !account.Settings.Active {
			utils.Log(utils.LogLevelInfo, "[%s - %s (%s)] Account is not active, waiting for activation",
				accountId, account.Domain.Name, account.Environment)

			c.updateDomainStatus(*account, model.StatusPending, "Account was deactivated")
			time.Sleep(1 * time.Minute)
			continue
		}

		// Refresh account repository settings
		gitService.UpdateRepositorySettings(account.Repository, account.Token, account.Branch)

		// Fetch repository data
		repositoryData, err := gitService.GetRepositoryData(account.Environment)

		if err != nil {
			utils.Log(utils.LogLevelError, "[%s - %s (%s)] Failed to fetch repository data - %s",
				accountId, account.Domain.Name, account.Environment, err.Error())

			c.updateDomainStatus(*account, model.StatusError, "Failed to fetch repository data - "+err.Error())
			time.Sleep(1 * time.Minute)
			continue
		}

		// Fetch snapshot version from API
		snapshotVersionPayload, err := c.ApiService.FetchSnapshotVersion(account.Domain.ID, account.Environment)

		if err != nil {
			utils.Log(utils.LogLevelError, "[%s - %s (%s)] Failed to fetch snapshot version - %s",
				accountId, account.Domain.Name, account.Environment, err.Error())

			c.updateDomainStatus(*account, model.StatusError, "Failed to fetch snapshot version - "+err.Error())
			time.Sleep(1 * time.Minute)
			continue
		}

		// Check if repository is out of sync
		if c.isOutSync(*account, repositoryData.CommitHash, snapshotVersionPayload) {
			c.syncUp(*account, repositoryData, gitService)
		}

		// Wait for the next cycle
		timeWindow, unitWindow := getTimeWindow(account.Settings.Window)
		time.Sleep(time.Duration(timeWindow) * unitWindow)
	}
}

func (c *CoreHandler) syncUp(account model.Account, repositoryData *model.RepositoryData, gitService IGitService) {
	utils.Log(utils.LogLevelInfo, "[%s - %s (%s)] Syncing up", account.ID.Hex(), account.Domain.Name, account.Environment)

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

	utils.Log(utils.LogLevelDebug, "[%s - %s (%s)] SnapshotAPI version: %s - SnapshotRepo version: %s",
		account.ID.Hex(), account.Domain.Name, account.Environment, fmt.Sprint(snapshotApi.Domain.Version), fmt.Sprint(account.Domain.Version))

	// Apply changes
	changeSource := ""
	if snapshotApi.Domain.Version > account.Domain.Version {
		changeSource = "Repository"
		if c.isRepositoryOutSync(repositoryData, diff) {
			account, err = c.applyChangesToRepository(account, snapshotApi, gitService)
		} else {
			utils.Log(utils.LogLevelInfo, "[%s - %s (%s)] Repository is up to date",
				account.ID.Hex(), account.Domain.Name, account.Environment)

			account.Domain.Version = snapshotApi.Domain.Version
			account.Domain.LastCommit = repositoryData.CommitHash
		}
	} else if len(diff.Changes) > 0 {
		changeSource = "API"
		account = c.applyChangesToAPI(account, repositoryData, diff)
	}

	if err != nil {
		utils.Log(utils.LogLevelError, "[%s - %s (%s)] Failed to apply changes [%s] - %s",
			account.ID.Hex(), account.Domain.Name, account.Environment, changeSource, err.Error())

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
	diffDeleted := c.ComparatorService.CheckSnapshotDiff(fromApi, fromRepo, DELETED)

	return c.ComparatorService.MergeResults([]model.DiffResult{diffNew, diffChanged, diffDeleted}), snapshotApi.Snapshot, nil
}

func (c *CoreHandler) applyChangesToAPI(account model.Account, repositoryData *model.RepositoryData, diff model.DiffResult) model.Account {
	utils.Log(utils.LogLevelInfo, "[%s - %s (%s)] Pushing changes to API", account.ID.Hex(), account.Domain.Name, account.Environment)

	// Removed deleted if force prune is disabled
	if !account.Settings.ForcePrune {
		c.ComparatorService.RemoveDeleted(diff)
	}

	// Push changes to API

	// Update domain
	account.Domain.Version = 2
	account.Domain.LastCommit = repositoryData.CommitHash

	return account
}

func (c *CoreHandler) applyChangesToRepository(account model.Account, snapshot model.Snapshot, gitService IGitService) (model.Account, error) {
	utils.Log(utils.LogLevelInfo, "[%s - %s (%s)] Pushing changes to repository", account.ID.Hex(), account.Domain.Name, account.Environment)

	// Remove version from domain
	snapshotContent := snapshot
	snapshotContent.Domain.Version = 0

	lastCommit, err := gitService.PushChanges(account.Environment, utils.ToJsonFromObject(snapshotContent))

	if err != nil {
		return account, err
	}

	// Update domain
	account.Domain.Version = snapshot.Domain.Version
	account.Domain.LastCommit = lastCommit

	return account, nil
}

func (c *CoreHandler) isOutSync(account model.Account, lastCommit string, snapshotVersionPayload string) bool {
	snapshotVersion := c.ApiService.NewDataFromJson([]byte(snapshotVersionPayload)).Snapshot.Domain.Version

	utils.Log(utils.LogLevelDebug, "[%s - %s (%s)] Checking account - Last commit: %s - Domain Version: %d - Snapshot Version: %d",
		account.ID.Hex(), account.Domain.Name, account.Environment, account.Domain.LastCommit, account.Domain.Version, snapshotVersion)

	return account.Domain.LastCommit == "" || // First sync
		account.Domain.LastCommit != lastCommit || // Repository out of sync
		account.Domain.Version != snapshotVersion // API out of sync
}

func (c *CoreHandler) isRepositoryOutSync(repositoryData *model.RepositoryData, diff model.DiffResult) bool {
	return len(repositoryData.Content) <= 1 || // File is empty
		len(diff.Changes) > 0 // Changes detected
}

func (c *CoreHandler) updateDomainStatus(account model.Account, status string, message string) {
	account.Domain.Status = status
	account.Domain.Message = message
	account.Domain.LastDate = time.Now().Format(time.ANSIC)
	c.AccountRepository.Update(&account)
}

func getTimeWindow(window string) (int, time.Duration) {
	duration, _ := time.ParseDuration(window)
	return 1, duration
}
