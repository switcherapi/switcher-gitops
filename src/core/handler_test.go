package core

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/switcherapi/switcher-gitops/src/model"
)

func TestInitCoreHandlerCoroutine(t *testing.T) {
	// Given
	fakeGitService := NewFakeGitService()
	fakeApiService := NewFakeApiService(false)
	coreHandler = NewCoreHandler(coreHandler.AccountRepository, fakeGitService, fakeApiService, NewComparatorService())

	account := givenAccount()
	coreHandler.AccountRepository.Create(&account)

	// Test
	status, err := coreHandler.InitCoreHandlerCoroutine()

	// Assert
	assert.Nil(t, err)
	assert.Equal(t, 1, status)

	tearDown()
}

func TestStartAccountHandler(t *testing.T) {
	t.Run("Should not sync when account is not active", func(t *testing.T) {
		// Given
		account := givenAccount()
		account.Settings.Active = false
		coreHandler.AccountRepository.Create(&account)

		// Prepare goroutine signals
		var wg sync.WaitGroup
		wg.Add(1)

		// Test
		go coreHandler.StartAccountHandler(account, make(chan bool), &wg)

		// Wait for the goroutine to run and terminate
		time.Sleep(1 * time.Second)
		wg.Wait()

		// Assert
		accountFromDb, _ := coreHandler.AccountRepository.FetchByDomainId(account.Domain.ID)
		assert.Equal(t, model.StatusOutSync, accountFromDb.Domain.Status)
		assert.Equal(t, "", accountFromDb.Domain.Message)
		assert.Equal(t, "", accountFromDb.Domain.LastCommit)

		tearDown()
	})

	t.Run("Should sync successfully when repository is out of sync", func(t *testing.T) {
		// Given
		fakeGitService := NewFakeGitService()
		fakeApiService := NewFakeApiService(false)
		coreHandler = NewCoreHandler(coreHandler.AccountRepository, fakeGitService, fakeApiService, NewComparatorService())

		account := givenAccount()
		coreHandler.AccountRepository.Create(&account)

		// Prepare goroutine signals
		var wg sync.WaitGroup
		quit := make(chan bool)
		wg.Add(1)

		// Test
		go coreHandler.StartAccountHandler(account, quit, &wg)

		// Wait for the goroutine to run and terminate
		time.Sleep(1 * time.Second)
		quit <- true
		wg.Wait()

		// Assert
		accountFromDb, _ := coreHandler.AccountRepository.FetchByDomainId(account.Domain.ID)
		assert.Equal(t, model.StatusSynced, accountFromDb.Domain.Status)
		assert.Equal(t, "Synced successfully", accountFromDb.Domain.Message)
		assert.Equal(t, "123", accountFromDb.Domain.LastCommit)
		assert.NotEqual(t, "", accountFromDb.Domain.LastDate)

		tearDown()
	})

	t.Run("Should not sync when API returns an error", func(t *testing.T) {
		// Given
		fakeGitService := NewFakeGitService()
		fakeApiService := NewFakeApiService(true)
		coreHandler = NewCoreHandler(coreHandler.AccountRepository, fakeGitService, fakeApiService, NewComparatorService())

		account := givenAccount()
		coreHandler.AccountRepository.Create(&account)

		// Prepare goroutine signals
		var wg sync.WaitGroup
		quit := make(chan bool)
		wg.Add(1)

		// Test
		go coreHandler.StartAccountHandler(account, quit, &wg)

		// Wait for the goroutine to run and terminate
		time.Sleep(1 * time.Second)
		quit <- true
		wg.Wait()

		// Assert
		accountFromDb, _ := coreHandler.AccountRepository.FetchByDomainId(account.Domain.ID)
		assert.Equal(t, model.StatusError, accountFromDb.Domain.Status)
		assert.Equal(t, "Error syncing up", accountFromDb.Domain.Message)
		assert.Equal(t, "123", accountFromDb.Domain.LastCommit)
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
	lastCommit string
	date       string
	content    string
	status     string
	message    string
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

func (f *FakeGitService) CheckForChanges(account model.Account, lastCommit string,
	date string, content string) (status string, message string) {
	return f.status, f.message
}

type FakeApiService struct {
	throwError bool
	response   string
}

func NewFakeApiService(throwError bool) *FakeApiService {
	return &FakeApiService{
		throwError: throwError,
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

func (f *FakeApiService) NewDataFromJson(jsonData []byte) model.Data {
	var data model.Data
	json.Unmarshal(jsonData, &data)
	return data
}
