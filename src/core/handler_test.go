package core

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInitCoreHandlerCoroutine(t *testing.T) {
	// Given
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
	assert.Equal(t, "Synced successfully", accountFromDb.Domain.Message)
	assert.Equal(t, "123", accountFromDb.Domain.LastCommit)

	tearDown()
}

// Helpers

func tearDown() {
	collection := mongoDb.Collection("accounts")
	collection.Drop(context.Background())
}
