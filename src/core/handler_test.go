package core

import (
	"context"
	"encoding/json"
	"errors"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/switcherapi/switcher-gitops/src/model"
)

func TestInitCoreHandlerGoroutine(t *testing.T) {
	t.Run("Should start account handlers for all active accounts", func(t *testing.T) {
		// Given
		fakeApiService := NewFakeApiService()
		coreHandler = NewCoreHandler(coreHandler.accountRepository, fakeApiService, NewComparatorService())

		account := givenAccount()
		account.Domain.ID = "123-init-core-handler"
		accountCreated, _ := coreHandler.accountRepository.Create(&account)

		// Test
		status, err := coreHandler.InitCoreHandlerGoroutine()

		// Terminate the goroutine
		coreHandler.accountRepository.DeleteByDomainIdEnvironment(accountCreated.Domain.ID, accountCreated.Environment)
		time.Sleep(1 * time.Second)

		// Assert
		assert.Nil(t, err)
		assert.Equal(t, CoreHandlerStatusRunning, status)

		tearDown()
	})

	t.Run("Should not start account handlers when core handler is already running", func(t *testing.T) {
		// Given
		fakeApiService := NewFakeApiService()
		coreHandler = NewCoreHandler(coreHandler.accountRepository, fakeApiService, NewComparatorService())

		account := givenAccount()
		account.Domain.ID = "123-not-init-core-handler"
		accountCreated, _ := coreHandler.accountRepository.Create(&account)

		// Test
		coreHandler.InitCoreHandlerGoroutine()
		status, _ := coreHandler.InitCoreHandlerGoroutine()

		// Terminate the goroutine
		coreHandler.accountRepository.DeleteByDomainIdEnvironment(accountCreated.Domain.ID, accountCreated.Environment)
		time.Sleep(1 * time.Second)

		// Assert
		assert.Equal(t, CoreHandlerStatusRunning, status)

		tearDown()
	})
}

func TestAccountHandlerSyncRepository(t *testing.T) {
	t.Run("Should sync successfully when API has a newer version", func(t *testing.T) {
		// Given
		fakeGitService := NewFakeGitService()
		fakeGitService.existingData.CommitHash = "111"

		fakeApiService := NewFakeApiService()
		coreHandler = NewCoreHandler(coreHandler.accountRepository, fakeApiService, NewComparatorService())

		account := givenAccount()
		account.Domain.ID = "123-newer-version"
		accountCreated, _ := coreHandler.accountRepository.Create(&account)

		// Test
		go coreHandler.StartAccountHandler(accountCreated.ID.Hex(), fakeGitService)

		// Wait for goroutine to process
		time.Sleep(1 * time.Second)

		// Assert
		accountFromDb, _ := coreHandler.accountRepository.FetchByAccountId(string(accountCreated.ID.Hex()))
		assert.Equal(t, model.StatusSynced, accountFromDb.Domain.Status)
		assert.Contains(t, accountFromDb.Domain.Message, model.MessageSynced)
		assert.Equal(t, "111", accountFromDb.Domain.LastCommit)
		assert.Equal(t, 1, accountFromDb.Domain.Version)
		assert.NotEqual(t, "", accountFromDb.Domain.LastDate)

		tearDown()
	})

	t.Run("Should sync successfully after account reactivation when it was pending", func(t *testing.T) {
		// Given
		fakeGitService := NewFakeGitService()
		fakeApiService := NewFakeApiService()
		fakeApiService.pushChanges = PushChangeResponse{
			Message: "Changes applied successfully",
			Version: 1,
		}
		coreHandler = NewCoreHandler(coreHandler.accountRepository, fakeApiService, NewComparatorService())

		account := givenAccount()
		account.Domain.ID = "123-pending"
		account.Domain.Status = model.StatusPending
		account.Domain.Message = "Account was deactivated"
		account.Domain.LastCommit = "123"
		account.Domain.Version = 1
		accountCreated, _ := coreHandler.accountRepository.Create(&account)

		// Test
		go coreHandler.StartAccountHandler(accountCreated.ID.Hex(), fakeGitService)

		// Wait for goroutine to process
		time.Sleep(1 * time.Second)

		// Assert
		accountFromDb, _ := coreHandler.accountRepository.FetchByAccountId(string(accountCreated.ID.Hex()))
		assert.Equal(t, model.StatusSynced, accountFromDb.Domain.Status)
		assert.Contains(t, accountFromDb.Domain.Message, model.MessageSynced)
		assert.Equal(t, "123", accountFromDb.Domain.LastCommit)
		assert.Equal(t, 1, accountFromDb.Domain.Version)
		assert.NotEqual(t, "", accountFromDb.Domain.LastDate)

		tearDown()
	})

	t.Run("Should sync successfully when repository is out of sync", func(t *testing.T) {
		// Given
		fakeGitService := NewFakeGitService()
		fakeApiService := NewFakeApiService()
		coreHandler = NewCoreHandler(coreHandler.accountRepository, fakeApiService, NewComparatorService())

		account := givenAccount()
		account.Domain.ID = "123-out-sync"
		accountCreated, _ := coreHandler.accountRepository.Create(&account)

		// Test
		go coreHandler.StartAccountHandler(accountCreated.ID.Hex(), fakeGitService)

		// Wait for goroutine to process
		time.Sleep(1 * time.Second)

		// Assert
		accountFromDb, _ := coreHandler.accountRepository.FetchByDomainIdEnvironment(accountCreated.Domain.ID, accountCreated.Environment)
		assert.Equal(t, model.StatusSynced, accountFromDb.Domain.Status)
		assert.Contains(t, accountFromDb.Domain.Message, model.MessageSynced)
		assert.Equal(t, "123", accountFromDb.Domain.LastCommit)
		assert.Equal(t, 1, accountFromDb.Domain.Version)
		assert.NotEqual(t, "", accountFromDb.Domain.LastDate)

		tearDown()
	})

	t.Run("Should sync successfully when repository is up to date but not synced", func(t *testing.T) {
		// Given
		fakeGitService := NewFakeGitService()
		fakeGitService.existingData.Content = `{
			"domain": {
				"group": [{
					"name": "Release 1",
					"description": "Showcase configuration",
					"activated": true,
					"config": [{
						"key": "MY_SWITCHER",
						"description": "",
						"activated": true,
						"strategies": [],
						"components": [
							"switcher-playground"
						]
					}]
				}]
			}
		}`
		fakeApiService := NewFakeApiService()
		coreHandler = NewCoreHandler(coreHandler.accountRepository, fakeApiService, NewComparatorService())

		account := givenAccount()
		account.Domain.ID = "123-up-to-date-not-synced"
		accountCreated, _ := coreHandler.accountRepository.Create(&account)

		// Test
		go coreHandler.StartAccountHandler(accountCreated.ID.Hex(), fakeGitService)

		// Wait for goroutine to process
		time.Sleep(1 * time.Second)

		// Assert
		accountFromDb, _ := coreHandler.accountRepository.FetchByDomainIdEnvironment(accountCreated.Domain.ID, accountCreated.Environment)
		assert.Equal(t, model.StatusSynced, accountFromDb.Domain.Status)
		assert.Contains(t, accountFromDb.Domain.Message, model.MessageSynced)
		assert.Equal(t, "123", accountFromDb.Domain.LastCommit)
		assert.Equal(t, 1, accountFromDb.Domain.Version)
		assert.NotEqual(t, "", accountFromDb.Domain.LastDate)

		tearDown()
	})

	t.Run("Should sync successfully when account was just created", func(t *testing.T) {
		// Given
		fakeGitService := NewFakeGitService()
		fakeGitService.errorGetRepoData = "file not found"
		fakeGitService.newData = fakeGitService.existingData
		fakeGitService.newData.Content = `{
			"domain": {
				"group": [{
					"name": "Release 1",
					"description": "Showcase configuration",
					"activated": true,
					"config": [{
						"key": "MY_SWITCHER",
						"description": "",
						"activated": true,
						"strategies": [],
						"components": [
							"switcher-playground"
						]
					}]
				}]
			}
		}`

		fakeApiService := NewFakeApiService()
		coreHandler = NewCoreHandler(coreHandler.accountRepository, fakeApiService, NewComparatorService())

		account := givenAccount()
		account.Domain.ID = "123-just-created"
		accountCreated, _ := coreHandler.accountRepository.Create(&account)

		// Test
		go coreHandler.StartAccountHandler(accountCreated.ID.Hex(), fakeGitService)

		// Wait for goroutine to process
		time.Sleep(1 * time.Second)

		// Assert
		accountFromDb, _ := coreHandler.accountRepository.FetchByDomainIdEnvironment(accountCreated.Domain.ID, accountCreated.Environment)
		assert.Equal(t, model.StatusSynced, accountFromDb.Domain.Status)
		assert.Contains(t, accountFromDb.Domain.Message, model.MessageSynced)
		assert.Equal(t, "123", accountFromDb.Domain.LastCommit)
		assert.Equal(t, 1, accountFromDb.Domain.Version)
		assert.NotEqual(t, "", accountFromDb.Domain.LastDate)

		tearDown()
	})
}

func TestAccountHandlerSyncAPI(t *testing.T) {
	t.Run("Should sync API when repository has new changes", func(t *testing.T) {
		// Given
		fakeGitService := NewFakeGitService()
		fakeApiService := NewFakeApiService()

		coreHandler = NewCoreHandler(coreHandler.accountRepository, fakeApiService, NewComparatorService())

		account := givenAccount()
		account.Domain.ID = "123-sync-api"
		account.Domain.Version = 1
		accountCreated, _ := coreHandler.accountRepository.Create(&account)

		// Test
		go coreHandler.StartAccountHandler(accountCreated.ID.Hex(), fakeGitService)

		// Wait for goroutine to process
		time.Sleep(1 * time.Second)

		// Assert
		accountFromDb, _ := coreHandler.accountRepository.FetchByDomainIdEnvironment(accountCreated.Domain.ID, accountCreated.Environment)
		assert.Equal(t, model.StatusSynced, accountFromDb.Domain.Status)
		assert.Contains(t, accountFromDb.Domain.Message, model.MessageSynced)
		assert.Equal(t, "123", accountFromDb.Domain.LastCommit)
		assert.Equal(t, 100, accountFromDb.Domain.Version)
		assert.NotEqual(t, "", accountFromDb.Domain.LastDate)

		tearDown()
	})

	t.Run("Should sync and prune successfully when API is out of sync", func(t *testing.T) {
		// Given
		fakeGitService := NewFakeGitService()
		fakeGitService.existingData.Content = `{
			"domain": {
				"group": []
			}
		}`
		fakeApiService := NewFakeApiService()
		coreHandler = NewCoreHandler(coreHandler.accountRepository, fakeApiService, NewComparatorService())

		account := givenAccount()
		account.Domain.ID = "123-out-sync-prune"
		account.Domain.Version = 1
		account.Settings.ForcePrune = true
		accountCreated, _ := coreHandler.accountRepository.Create(&account)

		// Test
		go coreHandler.StartAccountHandler(accountCreated.ID.Hex(), fakeGitService)

		// Wait for goroutine to process
		time.Sleep(1 * time.Second)

		// Assert
		accountFromDb, _ := coreHandler.accountRepository.FetchByDomainIdEnvironment(accountCreated.Domain.ID, accountCreated.Environment)
		assert.Equal(t, model.StatusSynced, accountFromDb.Domain.Status)
		assert.Contains(t, accountFromDb.Domain.Message, model.MessageSynced)
		assert.Equal(t, "123", accountFromDb.Domain.LastCommit)
		assert.Equal(t, 100, accountFromDb.Domain.Version)
		assert.NotEqual(t, "", accountFromDb.Domain.LastDate)

		tearDown()
	})

	t.Run("Should sync and not prune when API is out of sync", func(t *testing.T) {
		// Given
		fakeGitService := NewFakeGitService()
		fakeGitService.existingData.Content = `{
			"domain": {
				"group": []
			}
		}`
		fakeApiService := NewFakeApiService()
		coreHandler = NewCoreHandler(coreHandler.accountRepository, fakeApiService, NewComparatorService())

		account := givenAccount()
		account.Domain.ID = "123-out-sync-not-prune"
		account.Domain.Version = 1
		account.Settings.ForcePrune = false
		accountCreated, _ := coreHandler.accountRepository.Create(&account)

		// Test
		go coreHandler.StartAccountHandler(accountCreated.ID.Hex(), fakeGitService)

		// Wait for goroutine to process
		time.Sleep(1 * time.Second)

		// Assert
		accountFromDb, _ := coreHandler.accountRepository.FetchByDomainIdEnvironment(accountCreated.Domain.ID, accountCreated.Environment)
		assert.Equal(t, model.StatusSynced, accountFromDb.Domain.Status)
		assert.Contains(t, accountFromDb.Domain.Message, model.MessageSynced)
		assert.Equal(t, "123", accountFromDb.Domain.LastCommit)
		assert.Equal(t, 100, accountFromDb.Domain.Version)
		assert.NotEqual(t, "", accountFromDb.Domain.LastDate)

		tearDown()
	})
}

func TestAccountHandlerNotSync(t *testing.T) {
	t.Run("Should not sync when account is not active", func(t *testing.T) {
		// Given
		fakeGitService := NewFakeGitService()

		account := givenAccount()
		account.Domain.ID = "123-not-active"
		account.Settings.Active = false
		accountCreated, _ := coreHandler.accountRepository.Create(&account)

		// Test
		go coreHandler.StartAccountHandler(accountCreated.ID.Hex(), fakeGitService)

		time.Sleep(1 * time.Second)

		// Assert
		accountFromDb, _ := coreHandler.accountRepository.FetchByDomainIdEnvironment(accountCreated.Domain.ID, accountCreated.Environment)
		assert.Equal(t, model.StatusPending, accountFromDb.Domain.Status)
		assert.Equal(t, "Account was deactivated", accountFromDb.Domain.Message)
		assert.Equal(t, "", accountFromDb.Domain.LastCommit)

		tearDown()
	})

	t.Run("Should not sync when fetch repository data returns an error", func(t *testing.T) {
		// Given
		fakeGitService := NewFakeGitService()
		fakeGitService.errorGetRepoData = "something went wrong"

		account := givenAccount()
		account.Domain.ID = "123-error-get-repo-data"
		accountCreated, _ := coreHandler.accountRepository.Create(&account)

		// Test
		go coreHandler.StartAccountHandler(accountCreated.ID.Hex(), fakeGitService)

		time.Sleep(1 * time.Second)

		// Assert
		accountFromDb, _ := coreHandler.accountRepository.FetchByDomainIdEnvironment(accountCreated.Domain.ID, accountCreated.Environment)
		assert.Equal(t, model.StatusError, accountFromDb.Domain.Status)
		assert.Contains(t, accountFromDb.Domain.Message, "Failed to fetch repository data")
		assert.Equal(t, "", accountFromDb.Domain.LastCommit)

		tearDown()
	})

	t.Run("Should not sync when fetch repository data returns a malformed JSON content", func(t *testing.T) {
		// Given
		fakeGitService := NewFakeGitService()
		fakeGitService.existingData.Content = `{
			"domain": {
				"group": [{
					"name": "Release 1",
					"description": "Showcase configuration",
					"activated": true
		}`

		account := givenAccount()
		account.Domain.ID = "123-error-malformed-json"
		accountCreated, _ := coreHandler.accountRepository.Create(&account)

		// Test
		go coreHandler.StartAccountHandler(accountCreated.ID.Hex(), fakeGitService)

		time.Sleep(1 * time.Second)

		// Assert
		accountFromDb, _ := coreHandler.accountRepository.FetchByDomainIdEnvironment(accountCreated.Domain.ID, accountCreated.Environment)
		assert.Equal(t, model.StatusError, accountFromDb.Domain.Status)
		assert.Contains(t, accountFromDb.Domain.Message, "Invalid JSON content")
		assert.Equal(t, "", accountFromDb.Domain.LastCommit)

		tearDown()
	})

	t.Run("Should not sync when fetch snapshot version returns an error", func(t *testing.T) {
		// Given
		fakeGitService := NewFakeGitService()
		fakeApiService := NewFakeApiService()
		fakeApiService.throwErrorVersion = true

		coreHandler = NewCoreHandler(coreHandler.accountRepository, fakeApiService, NewComparatorService())

		account := givenAccount()
		account.Domain.ID = "123-error-fetch-snapshot-version"
		accountCreated, _ := coreHandler.accountRepository.Create(&account)

		// Test
		go coreHandler.StartAccountHandler(accountCreated.ID.Hex(), fakeGitService)

		time.Sleep(1 * time.Second)

		// Assert
		accountFromDb, _ := coreHandler.accountRepository.FetchByDomainIdEnvironment(accountCreated.Domain.ID, accountCreated.Environment)
		assert.Equal(t, model.StatusError, accountFromDb.Domain.Status)
		assert.Contains(t, accountFromDb.Domain.Message, "Failed to fetch snapshot version")
		assert.Equal(t, "", accountFromDb.Domain.LastCommit)

		tearDown()
	})

	t.Run("Should not sync after account is deleted", func(t *testing.T) {
		// Given
		fakeGitService := NewFakeGitService()

		account := givenAccount()
		account.Domain.ID = "123-deleted"
		accountCreated, _ := coreHandler.accountRepository.Create(&account)

		// Test
		go coreHandler.StartAccountHandler(accountCreated.ID.Hex(), fakeGitService)
		numGoroutinesBefore := runtime.NumGoroutine()

		// Terminate the goroutine
		coreHandler.accountRepository.DeleteByDomainIdEnvironment(accountCreated.Domain.ID, accountCreated.Environment)
		time.Sleep(1 * time.Second)
		numGoroutinesAfter := runtime.NumGoroutine()

		// Assert
		_, err := coreHandler.accountRepository.FetchByDomainIdEnvironment(accountCreated.Domain.ID, accountCreated.Environment)
		assert.LessOrEqual(t, numGoroutinesAfter, numGoroutinesBefore)
		assert.NotNil(t, err)

		tearDown()
	})

	t.Run("Should not sync when API returns an error", func(t *testing.T) {
		// Given
		fakeGitService := NewFakeGitService()
		fakeApiService := NewFakeApiService()
		fakeApiService.throwErrorSnapshot = true

		coreHandler = NewCoreHandler(coreHandler.accountRepository, fakeApiService, NewComparatorService())

		account := givenAccount()
		account.Domain.ID = "123-api-error"
		accountCreated, _ := coreHandler.accountRepository.Create(&account)

		// Test
		go coreHandler.StartAccountHandler(accountCreated.ID.Hex(), fakeGitService)

		// Wait for goroutine to process
		time.Sleep(1 * time.Second)

		// Assert
		accountFromDb, _ := coreHandler.accountRepository.FetchByDomainIdEnvironment(accountCreated.Domain.ID, accountCreated.Environment)
		assert.Equal(t, model.StatusError, accountFromDb.Domain.Status)
		assert.Contains(t, accountFromDb.Domain.Message, "Failed to check for changes")
		assert.Equal(t, "123", accountFromDb.Domain.LastCommit)
		assert.NotEqual(t, "", accountFromDb.Domain.LastDate)

		tearDown()
	})

	t.Run("Should not sync when comparatorService panics", func(t *testing.T) {
		// Given
		fakeGitService := NewFakeGitService()
		fakeApiService := NewFakeApiService()
		fakeComparatorService := NewFakeComparatorService()
		fakeComparatorService.throwPanicCheckDiff = true

		coreHandler = NewCoreHandler(coreHandler.accountRepository, fakeApiService, fakeComparatorService)
		account := givenAccount()
		account.Domain.ID = "123-comparator-panic"
		accountCreated, _ := coreHandler.accountRepository.Create(&account)

		// Test
		go coreHandler.StartAccountHandler(accountCreated.ID.Hex(), fakeGitService)

		// Wait for goroutine to process
		time.Sleep(1 * time.Second)

		// Assert
		accountFromDb, _ := coreHandler.accountRepository.FetchByDomainIdEnvironment(accountCreated.Domain.ID, accountCreated.Environment)
		assert.Equal(t, model.StatusError, accountFromDb.Domain.Status)
		assert.Contains(t, accountFromDb.Domain.Message, "Panic occurred in CheckSnapshotDiff")
		assert.Equal(t, "123", accountFromDb.Domain.LastCommit)
		assert.NotEqual(t, "", accountFromDb.Domain.LastDate)
		assert.Equal(t, 0, accountFromDb.Domain.Version)

		tearDown()
	})

	t.Run("Should not sync when git token has no permission to push changes", func(t *testing.T) {
		// Given
		fakeGitService := NewFakeGitService()
		fakeGitService.errorPushChanges = "authorization failed"
		fakeApiService := NewFakeApiService()

		coreHandler = NewCoreHandler(coreHandler.accountRepository, fakeApiService, NewComparatorService())

		account := givenAccount()
		account.Domain.ID = "123-no-permission"
		accountCreated, _ := coreHandler.accountRepository.Create(&account)

		// Test
		go coreHandler.StartAccountHandler(accountCreated.ID.Hex(), fakeGitService)

		// Wait for goroutine to process
		time.Sleep(1 * time.Second)

		// Assert
		accountFromDb, _ := coreHandler.accountRepository.FetchByDomainIdEnvironment(accountCreated.Domain.ID, accountCreated.Environment)
		assert.Equal(t, model.StatusError, accountFromDb.Domain.Status)
		assert.Contains(t, accountFromDb.Domain.Message, "authorization failed")
		assert.Contains(t, accountFromDb.Domain.Message, "Failed to push changes [Repository]")
		assert.Equal(t, "123", accountFromDb.Domain.LastCommit)
		assert.NotEqual(t, "", accountFromDb.Domain.LastDate)

		tearDown()
	})

	t.Run("Should not sync when API returns an error while pushing changes", func(t *testing.T) {
		// Given
		fakeGitService := NewFakeGitService()
		fakeApiService := NewFakeApiService()
		fakeApiService.throwErrorPush = true

		coreHandler = NewCoreHandler(coreHandler.accountRepository, fakeApiService, NewComparatorService())

		account := givenAccount()
		account.Domain.ID = "123-api-error-push"
		account.Domain.Version = 1
		accountCreated, _ := coreHandler.accountRepository.Create(&account)

		// Test
		go coreHandler.StartAccountHandler(accountCreated.ID.Hex(), fakeGitService)

		// Wait for goroutine to process
		time.Sleep(1 * time.Second)

		// Assert
		accountFromDb, _ := coreHandler.accountRepository.FetchByDomainIdEnvironment(accountCreated.Domain.ID, accountCreated.Environment)
		assert.Equal(t, model.StatusError, accountFromDb.Domain.Status)
		assert.Contains(t, accountFromDb.Domain.Message, "Failed to push changes")
		assert.Equal(t, "123", accountFromDb.Domain.LastCommit)
		assert.NotEqual(t, "", accountFromDb.Domain.LastDate)

		tearDown()
	})

	t.Run("Should not sync when account was just created - failed to fetch API", func(t *testing.T) {
		// Given
		fakeGitService := NewFakeGitService()
		fakeGitService.errorGetRepoData = "file not found"
		fakeGitService.newData = fakeGitService.existingData

		fakeApiService := NewFakeApiService()
		fakeApiService.throwErrorSnapshot = true
		coreHandler = NewCoreHandler(coreHandler.accountRepository, fakeApiService, NewComparatorService())

		account := givenAccount()
		account.Domain.ID = "123-just-created-error-fetch-api"
		accountCreated, _ := coreHandler.accountRepository.Create(&account)

		// Test
		go coreHandler.StartAccountHandler(accountCreated.ID.Hex(), fakeGitService)

		// Wait for goroutine to process
		time.Sleep(1 * time.Second)

		// Assert
		accountFromDb, _ := coreHandler.accountRepository.FetchByDomainIdEnvironment(accountCreated.Domain.ID, accountCreated.Environment)
		assert.Equal(t, model.StatusError, accountFromDb.Domain.Status)
		assert.Contains(t, accountFromDb.Domain.Message, "Failed to fetch repository data")
		assert.Equal(t, "", accountFromDb.Domain.LastCommit)
		assert.Equal(t, 0, accountFromDb.Domain.Version)
		assert.NotEqual(t, "", accountFromDb.Domain.LastDate)

		tearDown()
	})
}

// Helpers

func tearDown() {
	collection := mongoDb.Collection(model.CollectionName)
	collection.Drop(context.Background())
}

// Fakes

type FakeGitService struct {
	existingData     model.RepositoryData
	newData          model.RepositoryData
	status           string
	errorPushChanges string
	errorGetRepoData string
}

func NewFakeGitService() *FakeGitService {
	return &FakeGitService{
		existingData: model.RepositoryData{
			CommitHash: "123",
			CommitDate: time.Now().Format(time.ANSIC),
			Content: `{
				"domain": {
					"group": [{
						"name": "Release 1",
						"description": "Showcase configuration",
						"activated": true,
						"config": [{
							"key": "MY_SWITCHER",
							"description": "",
							"activated": false,
							"strategies": [],
							"components": [
								"switcher-playground"
							]
						}]
					}]
				}
			}`,
		},
		status: model.StatusOutSync,
	}
}

func (f *FakeGitService) GetRepositoryData(environment string) (*model.RepositoryData, error) {
	if f.errorGetRepoData != "" {
		return nil, errors.New(f.errorGetRepoData)
	}

	return &f.existingData, nil
}

func (f *FakeGitService) CreateRepositoryData(environment string, content string) (*model.RepositoryData, error) {
	return &f.newData, nil
}

func (f *FakeGitService) PushChanges(environment string, content string, message string) (*model.RepositoryData, error) {
	if f.errorPushChanges != "" {
		return nil, errors.New(f.errorPushChanges)
	}

	return &f.existingData, nil
}

func (f *FakeGitService) UpdateRepositorySettings(repository string, token string, branch string, path string) {
	// Do nothing
}

type FakeApiService struct {
	throwErrorVersion  bool
	throwErrorSnapshot bool
	throwErrorPush     bool
	responseVersion    string
	responseSnapshot   string
	pushChanges        PushChangeResponse
}

func NewFakeApiService() *FakeApiService {
	return &FakeApiService{
		throwErrorVersion:  false,
		throwErrorSnapshot: false,
		throwErrorPush:     false,
		pushChanges: PushChangeResponse{
			Message: "Changes applied successfully",
			Version: 100,
		},
		responseVersion: `{
			"data": {
				"domain": {
					"version": 1
				}
			}
		}`,
		responseSnapshot: `{
			"data": {
				"domain": {
					"name": "Switcher GitOps",
					"version": 1,
					"group": [{
						"name": "Release 1",
						"description": "Showcase configuration",
						"activated": true,
						"config": [{
							"key": "MY_SWITCHER",
							"description": "",
							"activated": true,
							"strategies": [],
							"components": [
								"switcher-playground"
							]
						}]
					}]
				}
			}
		}`,
	}
}

func (f *FakeApiService) FetchSnapshotVersion(domainId string, environment string) (string, error) {
	if f.throwErrorVersion {
		return "", errors.New("something went wrong")
	}

	return f.responseVersion, nil
}

func (f *FakeApiService) FetchSnapshot(domainId string, environment string) (string, error) {
	if f.throwErrorSnapshot {
		return "", errors.New("something went wrong")
	}

	return f.responseSnapshot, nil
}

func (f *FakeApiService) PushChanges(domainId string, diff model.DiffResult) (PushChangeResponse, error) {
	if f.throwErrorPush {
		return PushChangeResponse{}, errors.New("something went wrong")
	}

	return f.pushChanges, nil
}

func (f *FakeApiService) NewDataFromJson(jsonData []byte) model.Data {
	var data model.Data
	json.Unmarshal(jsonData, &data)
	return data
}

type FakeComparatorService struct {
	throwPanicCheckDiff bool
}

func NewFakeComparatorService() *FakeComparatorService {
	return &FakeComparatorService{}
}

func (f *FakeComparatorService) CheckSnapshotDiff(left model.Snapshot, right model.Snapshot, diffType DiffType) model.DiffResult {
	if f.throwPanicCheckDiff {
		panic("Panic occurred in CheckSnapshotDiff")
	}
	return model.DiffResult{}
}

func (f *FakeComparatorService) MergeResults(diffResults []model.DiffResult) model.DiffResult {
	return model.DiffResult{}
}

func (f *FakeComparatorService) NewSnapshotFromJson(jsonData []byte) model.Snapshot {
	var snapshot model.Snapshot
	json.Unmarshal(jsonData, &snapshot)
	return snapshot
}

func (f *FakeComparatorService) RemoveDeleted(diffResult model.DiffResult) model.DiffResult {
	return diffResult
}
