package core

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/switcherapi/switcher-gitops/src/model"
)

func TestInitCoreHandlerCoroutine(t *testing.T) {
	// Given
	fakeGitService := NewFakeGitService()
	coreHandler = NewCoreHandler(coreHandler.AccountRepository, fakeGitService, NewComparatorService())

	account := givenAccount()
	coreHandler.AccountRepository.Create(&account)

	// Test
	status, err := coreHandler.InitCoreHandlerCoroutine()

	// Assert
	assert.Nil(t, err)
	assert.Equal(t, 1, status)

	tearDown()
}

func TestStartAccountHandlerInactiveAccount(t *testing.T) {
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
	assert.Equal(t, "", accountFromDb.Domain.Message)
	assert.Equal(t, "", accountFromDb.Domain.LastCommit)

	tearDown()
}

func TestStartAccountHandler(t *testing.T) {
	// Given
	fakeGitService := NewFakeGitService()
	fakeGitService.status = model.StatusSynced
	fakeGitService.message = "Synced successfully"
	coreHandler = NewCoreHandler(coreHandler.AccountRepository, fakeGitService, NewComparatorService())

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
		content:    "Content",
		status:     model.StatusOutSync,
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
