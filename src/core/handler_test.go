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

func TestInitCoreHandlerCoroutine(t *testing.T) {
	// Given
	fakeGitService := NewFakeGitService()
	fakeApiService := NewFakeApiService()
	coreHandler = NewCoreHandler(coreHandler.AccountRepository, fakeGitService, fakeApiService, NewComparatorService())

	account := givenAccount()
	account.Domain.ID = "123-init-core-handler"
	accountCreated, _ := coreHandler.AccountRepository.Create(&account)

	// Test
	status, err := coreHandler.InitCoreHandlerCoroutine()

	// Terminate the goroutine
	coreHandler.AccountRepository.DeleteByDomainId(accountCreated.Domain.ID)
	time.Sleep(1 * time.Second)

	// Assert
	assert.Nil(t, err)
	assert.Equal(t, 1, status)

	tearDown()
}

func TestStartAccountHandler(t *testing.T) {
	t.Run("Should not sync when account is not active", func(t *testing.T) {
		// Given
		account := givenAccount()
		account.Domain.ID = "123-not-active"
		account.Settings.Active = false
		accountCreated, _ := coreHandler.AccountRepository.Create(&account)

		// Test
		go coreHandler.StartAccountHandler(accountCreated.ID.Hex())

		time.Sleep(1 * time.Second)

		// Assert
		accountFromDb, _ := coreHandler.AccountRepository.FetchByDomainId(accountCreated.Domain.ID)
		assert.Equal(t, model.StatusPending, accountFromDb.Domain.Status)
		assert.Equal(t, "Account was deactivated", accountFromDb.Domain.Message)
		assert.Equal(t, "", accountFromDb.Domain.LastCommit)

		tearDown()
	})

	t.Run("Should not sync after account is deleted", func(t *testing.T) {
		// Given
		account := givenAccount()
		account.Domain.ID = "123-deleted"
		accountCreated, _ := coreHandler.AccountRepository.Create(&account)

		// Test
		go coreHandler.StartAccountHandler(accountCreated.ID.Hex())
		numGoroutinesBefore := runtime.NumGoroutine()

		// Terminate the goroutine
		coreHandler.AccountRepository.DeleteByDomainId(accountCreated.Domain.ID)
		time.Sleep(1 * time.Second)
		numGoroutinesAfter := runtime.NumGoroutine()

		// Assert
		_, err := coreHandler.AccountRepository.FetchByDomainId(accountCreated.Domain.ID)
		assert.LessOrEqual(t, numGoroutinesAfter, numGoroutinesBefore)
		assert.NotNil(t, err)

		tearDown()
	})

	t.Run("Should sync successfully when repository is out of sync", func(t *testing.T) {
		// Given
		fakeGitService := NewFakeGitService()
		fakeApiService := NewFakeApiService()
		coreHandler = NewCoreHandler(coreHandler.AccountRepository, fakeGitService, fakeApiService, NewComparatorService())

		account := givenAccount()
		account.Domain.ID = "123-out-sync"
		accountCreated, _ := coreHandler.AccountRepository.Create(&account)

		// Test
		go coreHandler.StartAccountHandler(accountCreated.ID.Hex())

		// Wait for goroutine to process
		time.Sleep(1 * time.Second)

		// Assert
		accountFromDb, _ := coreHandler.AccountRepository.FetchByDomainId(accountCreated.Domain.ID)
		assert.Equal(t, model.StatusSynced, accountFromDb.Domain.Status)
		assert.Contains(t, accountFromDb.Domain.Message, model.MessageSynced)
		assert.Equal(t, "123", accountFromDb.Domain.LastCommit)
		assert.Equal(t, "1", accountFromDb.Domain.Version)
		assert.NotEqual(t, "", accountFromDb.Domain.LastDate)

		tearDown()
	})

	t.Run("Should sync and prune successfully when repository is out of sync", func(t *testing.T) {
		// Given
		fakeGitService := NewFakeGitService()
		fakeGitService.content = `{
			"domain": {
				"group": []
			}
		}`
		fakeApiService := NewFakeApiService()
		coreHandler = NewCoreHandler(coreHandler.AccountRepository, fakeGitService, fakeApiService, NewComparatorService())

		account := givenAccount()
		account.Domain.ID = "123-out-sync-prune"
		account.Domain.Version = "1"
		account.Settings.ForcePrune = true
		accountCreated, _ := coreHandler.AccountRepository.Create(&account)

		// Test
		go coreHandler.StartAccountHandler(accountCreated.ID.Hex())

		// Wait for goroutine to process
		time.Sleep(1 * time.Second)

		// Assert
		accountFromDb, _ := coreHandler.AccountRepository.FetchByDomainId(accountCreated.Domain.ID)
		assert.Equal(t, model.StatusSynced, accountFromDb.Domain.Status)
		assert.Contains(t, accountFromDb.Domain.Message, model.MessageSynced)
		assert.Equal(t, "123", accountFromDb.Domain.LastCommit)
		assert.Equal(t, "2", accountFromDb.Domain.Version)
		assert.NotEqual(t, "", accountFromDb.Domain.LastDate)

		tearDown()
	})

	t.Run("Should sync successfully when API has a newer version", func(t *testing.T) {
		// Given
		fakeGitService := NewFakeGitService()
		fakeGitService.lastCommit = "111"

		fakeApiService := NewFakeApiService()
		coreHandler = NewCoreHandler(coreHandler.AccountRepository, fakeGitService, fakeApiService, NewComparatorService())

		account := givenAccount()
		account.Domain.ID = "123-newer-version"
		accountCreated, _ := coreHandler.AccountRepository.Create(&account)

		// Test
		go coreHandler.StartAccountHandler(accountCreated.ID.Hex())

		// Wait for goroutine to process
		time.Sleep(1 * time.Second)

		// Assert
		accountFromDb, _ := coreHandler.AccountRepository.FetchByAccountId(string(accountCreated.ID.Hex()))
		assert.Equal(t, model.StatusSynced, accountFromDb.Domain.Status)
		assert.Contains(t, accountFromDb.Domain.Message, model.MessageSynced)
		assert.Equal(t, "111", accountFromDb.Domain.LastCommit)
		assert.Equal(t, "1", accountFromDb.Domain.Version)
		assert.NotEqual(t, "", accountFromDb.Domain.LastDate)

		tearDown()
	})

	t.Run("Should not sync when API returns an error", func(t *testing.T) {
		// Given
		fakeGitService := NewFakeGitService()
		fakeApiService := NewFakeApiService()
		fakeApiService.throwError = true

		coreHandler = NewCoreHandler(coreHandler.AccountRepository, fakeGitService, fakeApiService, NewComparatorService())

		account := givenAccount()
		account.Domain.ID = "123-api-error"
		accountCreated, _ := coreHandler.AccountRepository.Create(&account)

		// Test
		go coreHandler.StartAccountHandler(accountCreated.ID.Hex())

		// Wait for goroutine to process
		time.Sleep(1 * time.Second)

		// Assert
		accountFromDb, _ := coreHandler.AccountRepository.FetchByDomainId(accountCreated.Domain.ID)
		assert.Equal(t, model.StatusError, accountFromDb.Domain.Status)
		assert.Contains(t, accountFromDb.Domain.Message, "Failed to check for changes")
		assert.Equal(t, "123", accountFromDb.Domain.LastCommit)
		assert.NotEqual(t, "", accountFromDb.Domain.LastDate)

		tearDown()
	})

	t.Run("Should not sync when git token has no permission to push changes", func(t *testing.T) {
		// Given
		fakeGitService := NewFakeGitService()
		fakeGitService.errorPushChanges = "authorization failed"
		fakeApiService := NewFakeApiService()

		coreHandler = NewCoreHandler(coreHandler.AccountRepository, fakeGitService, fakeApiService, NewComparatorService())

		account := givenAccount()
		account.Domain.ID = "123-no-permission"
		accountCreated, _ := coreHandler.AccountRepository.Create(&account)

		// Test
		go coreHandler.StartAccountHandler(accountCreated.ID.Hex())

		// Wait for goroutine to process
		time.Sleep(1 * time.Second)

		// Assert
		accountFromDb, _ := coreHandler.AccountRepository.FetchByDomainId(accountCreated.Domain.ID)
		assert.Equal(t, model.StatusError, accountFromDb.Domain.Status)
		assert.Contains(t, accountFromDb.Domain.Message, "authorization failed")
		assert.Contains(t, accountFromDb.Domain.Message, "Failed to apply changes [Repository]")
		assert.Equal(t, "", accountFromDb.Domain.LastCommit)
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
	lastCommit       string
	date             string
	content          string
	status           string
	errorPushChanges string
}

func NewFakeGitService() *FakeGitService {
	return &FakeGitService{
		lastCommit: "123",
		date:       time.Now().Format(time.ANSIC),
		content: `{
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
		status: model.StatusOutSync,
	}
}

func (f *FakeGitService) GetRepositoryData(environment string) (*model.RepositoryData, error) {
	return &model.RepositoryData{
		CommitHash: f.lastCommit,
		CommitDate: f.date,
		Content:    f.content,
	}, nil
}

func (f *FakeGitService) PushChanges(environment string, content string) (string, error) {
	if f.errorPushChanges != "" {
		return "", errors.New(f.errorPushChanges)
	}

	return f.lastCommit, nil
}

type FakeApiService struct {
	throwError bool
	response   string
}

func NewFakeApiService() *FakeApiService {
	return &FakeApiService{
		throwError: false,
		response: `{
			"data": {
				"domain": {
					"name": "Switcher GitOps",
					"version": "1",
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

func (f *FakeApiService) FetchSnapshot(domainId string, environment string) (string, error) {
	if f.throwError {
		return "", assert.AnError
	}

	return f.response, nil
}

func (f *FakeApiService) ApplyChangesToAPI(domainId string, environment string, diff model.DiffResult) (ApplyChangeResponse, error) {
	return ApplyChangeResponse{}, nil
}

func (f *FakeApiService) NewDataFromJson(jsonData []byte) model.Data {
	var data model.Data
	json.Unmarshal(jsonData, &data)
	return data
}
