package core

import (
	"time"

	"github.com/switcherapi/switcher-gitops/src/config"
	"github.com/switcherapi/switcher-gitops/src/model"
	"github.com/switcherapi/switcher-gitops/src/repository"
	"github.com/switcherapi/switcher-gitops/src/utils"
)

const (
	CoreHandlerStatusCreated = -1
	CoreHandlerStatusInit    = 0
	CoreHandlerStatusRunning = 1
)

type CoreHandler struct {
	accountRepository repository.AccountRepository
	apiService        IAPIService
	comparatorService IComparatorService
	waitingTime       time.Duration
	Status            int
}

func NewCoreHandler(accountRepository repository.AccountRepository, apiService IAPIService,
	comparatorService IComparatorService) *CoreHandler {
	timeWindow, unitWindow := utils.GetTimeWindow(config.GetEnv("HANDLER_WAITING_TIME"))
	waitingTime := time.Duration(timeWindow) * unitWindow

	return &CoreHandler{
		accountRepository: accountRepository,
		apiService:        apiService,
		comparatorService: comparatorService,
		waitingTime:       waitingTime,
		Status:            CoreHandlerStatusCreated,
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
	accounts := c.accountRepository.FetchAllAccounts()

	// Iterate over accounts and start account handlers
	for _, account := range accounts {
		go c.StartAccountHandler(account.ID.Hex(), NewGitService(
			account.Repository,
			account.Token,
			account.Branch,
			account.Path,
		))
	}

	// Update core handler status
	c.Status = CoreHandlerStatusRunning
	return c.Status, nil
}

func (c *CoreHandler) StartAccountHandler(accountId string, gitService IGitService) {
	utils.LogInfo("[%s] Starting account handler", accountId)

	for {
		// Refresh account settings
		account, _ := c.accountRepository.FetchByAccountId(accountId)

		if account == nil {
			// Terminate the goroutine (account was deleted)
			utils.LogInfo("[%s] Account was deleted, terminating account handler", accountId)
			return
		}

		// Wait for account to be active
		if !account.Settings.Active {
			c.updateDomainStatus(*account, model.StatusPending, "Account was deactivated",
				utils.LogLevelInfo)
			time.Sleep(time.Duration(c.waitingTime))
			continue
		}

		// Refresh account repository settings
		gitService.UpdateRepositorySettings(account.Repository, account.Token, account.Branch, account.Path)

		// Fetch repository data
		repositoryData, err := gitService.GetRepositoryData(account.Environment)

		if err != nil {
			c.updateDomainStatus(*account, model.StatusError, "Failed to fetch repository data - "+err.Error(),
				utils.LogLevelError)
			time.Sleep(time.Duration(c.waitingTime))
			continue
		}

		if !utils.IsJsonValid(repositoryData.Content, &model.Snapshot{}) {
			c.updateDomainStatus(*account, model.StatusError, "Invalid JSON content", utils.LogLevelError)
			time.Sleep(time.Duration(c.waitingTime))
			continue
		}

		// Fetch snapshot version from API
		snapshotVersionPayload, err := c.apiService.FetchSnapshotVersion(account.Domain.ID, account.Environment)

		if err != nil {
			c.updateDomainStatus(*account, model.StatusError, "Failed to fetch snapshot version - "+err.Error(),
				utils.LogLevelError)
			time.Sleep(time.Duration(c.waitingTime))
			continue
		}

		// Check if repository is out of sync
		if c.isOutSync(*account, repositoryData.CommitHash, snapshotVersionPayload) {
			c.syncUp(*account, repositoryData, gitService)
		}

		// Wait for the next cycle
		timeWindow, unitWindow := utils.GetTimeWindow(account.Settings.Window)
		time.Sleep(time.Duration(timeWindow) * unitWindow)
	}
}

func (c *CoreHandler) syncUp(account model.Account, repositoryData *model.RepositoryData, gitService IGitService) {
	// Update account status: Out of sync
	account.Domain.LastCommit = repositoryData.CommitHash
	account.Domain.LastDate = repositoryData.CommitDate
	c.updateDomainStatus(account, model.StatusOutSync, model.MessageSyncingUp, utils.LogLevelInfo)

	// Check for changes
	diff, snapshotApi, err := c.checkForChanges(account, repositoryData.Content)

	if err != nil {
		c.updateDomainStatus(account, model.StatusError, "Failed to check for changes - "+err.Error(),
			utils.LogLevelError)
		return
	}

	// Push changes
	changeSource, account, err := c.pushChanges(snapshotApi, account, repositoryData, diff, gitService)

	if err != nil {
		c.updateDomainStatus(account, model.StatusError, "Failed to push changes ["+changeSource+"] - "+err.Error(),
			utils.LogLevelError)
		return
	}

	// Update account status: Synced
	c.updateDomainStatus(account, model.StatusSynced, model.MessageSynced, utils.LogLevelInfo)
}

func (c *CoreHandler) checkForChanges(account model.Account, content string) (model.DiffResult, model.Snapshot, error) {
	// Get Snapshot from API
	snapshotJsonFromApi, err := c.apiService.FetchSnapshot(account.Domain.ID, account.Environment)

	if err != nil {
		return model.DiffResult{}, model.Snapshot{}, err
	}

	// Convert API JSON to model.Snapshot
	snapshotApi := c.apiService.NewDataFromJson([]byte(snapshotJsonFromApi))
	fromApi := snapshotApi.Snapshot

	// Convert content to model.Snapshot
	jsonRight := []byte(content)
	fromRepo := c.comparatorService.NewSnapshotFromJson(jsonRight)

	// Compare Snapshots and get diff
	diffNew := c.comparatorService.CheckSnapshotDiff(fromRepo, fromApi, NEW)
	diffChanged := c.comparatorService.CheckSnapshotDiff(fromApi, fromRepo, CHANGED)
	diffDeleted := c.comparatorService.CheckSnapshotDiff(fromApi, fromRepo, DELETED)

	return c.comparatorService.MergeResults([]model.DiffResult{diffNew, diffChanged, diffDeleted}), snapshotApi.Snapshot, nil
}

func (c *CoreHandler) pushChanges(snapshotApi model.Snapshot, account model.Account,
	repositoryData *model.RepositoryData, diff model.DiffResult, gitService IGitService) (string, model.Account, error) {
	err := error(nil)

	changeSource := ""
	if snapshotApi.Domain.Version > account.Domain.Version {
		changeSource = "Repository"
		if c.isRepositoryOutSync(repositoryData, diff) {
			account, err = c.pushChangesToRepository(account, snapshotApi, gitService)
		} else {
			utils.LogInfo("[%s - %s (%s)] Repository is up to date",
				account.ID.Hex(), account.Domain.Name, account.Environment)

			account.Domain.Version = snapshotApi.Domain.Version
		}
	} else if len(diff.Changes) > 0 {
		changeSource = "API"
		account, err = c.pushChangesToAPI(account, repositoryData, diff)
	}

	return changeSource, account, err
}

func (c *CoreHandler) pushChangesToAPI(account model.Account,
	repositoryData *model.RepositoryData, diff model.DiffResult) (model.Account, error) {
	utils.LogInfo("[%s - %s (%s)] Pushing changes to API (prune: %t)", account.ID.Hex(), account.Domain.Name,
		account.Environment, account.Settings.ForcePrune)

	// Removed deleted if force prune is disabled
	if !account.Settings.ForcePrune {
		diff = c.comparatorService.RemoveDeleted(diff)
	}

	// Push changes to API
	diff.Environment = account.Environment
	apiResponse, err := c.apiService.PushChanges(account.Domain.ID, diff)

	if err != nil {
		return account, err
	}

	// Update domain
	account.Domain.Version = apiResponse.Version
	account.Domain.LastCommit = repositoryData.CommitHash

	return account, nil
}

func (c *CoreHandler) pushChangesToRepository(account model.Account, snapshot model.Snapshot,
	gitService IGitService) (model.Account, error) {
	utils.LogInfo("[%s - %s (%s)] Pushing changes to repository", account.ID.Hex(), account.Domain.Name, account.Environment)

	// Remove version from domain
	snapshotContent := snapshot
	snapshotContent.Domain.Version = 0

	// Push changes to repository
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
	snapshotVersion := c.apiService.NewDataFromJson([]byte(snapshotVersionPayload)).Snapshot.Domain.Version

	utils.LogDebug("[%s - %s (%s)] Checking account - Last commit: %s - GitOps Version: %d - API Version: %d",
		account.ID.Hex(), account.Domain.Name, account.Environment, account.Domain.LastCommit,
		account.Domain.Version, snapshotVersion)

	return account.Domain.LastCommit == "" || // First sync
		account.Domain.LastCommit != lastCommit || // Repository out of sync
		account.Domain.Version != snapshotVersion || // API out of sync
		account.Domain.Status != model.StatusSynced // Account out of sync
}

func (c *CoreHandler) isRepositoryOutSync(repositoryData *model.RepositoryData, diff model.DiffResult) bool {
	return len(repositoryData.Content) <= 1 || // File is empty
		len(diff.Changes) > 0 // Changes detected
}

func (c *CoreHandler) updateDomainStatus(account model.Account, status string, message string, logLevel string) {
	utils.Log(logLevel, "[%s - %s (%s)] %s", account.ID.Hex(), account.Domain.Name, account.Environment, message)

	account.Domain.Status = status
	account.Domain.Message = message
	account.Domain.LastDate = time.Now().Format(time.ANSIC)
	c.accountRepository.Update(&account)
}
