package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateAccount(t *testing.T) {
	// Given
	account := givenAccount(true)

	// Test
	createdAccount, err := accountRepository.Create(&account)

	// Assert
	assert.Nil(t, err)
	assert.NotNil(t, createdAccount.ID)
}

func TestFetchAccountByDomainId(t *testing.T) {
	// Given
	account := givenAccount(true)
	accountRepository.Create(&account)

	// Test
	fetchedAccount, err := accountRepository.FetchByDomainId(account.Domain.ID)

	// Assert
	assert.Nil(t, err)
	assert.NotNil(t, fetchedAccount.ID)
}

func TestFetchAccountByDomainIdNotFound(t *testing.T) {
	// Test
	_, err := accountRepository.FetchByDomainId("non_existent_domain_id")

	// Assert
	assert.NotNil(t, err)
}

func TestFetchAllActiveAccounts(t *testing.T) {
	// Given
	account1 := givenAccount(true)
	account2 := givenAccount(false)
	accountRepository.Create(&account1)
	accountRepository.Create(&account2)

	// Test
	accounts, err := accountRepository.FetchAllActiveAccounts()

	// Assert
	assert.Nil(t, err)
	assert.Equal(t, 1, len(accounts))
}

func TestUpdateAccount(t *testing.T) {
	// Given
	account := givenAccount(true)
	accountRepository.Create(&account)

	// Test
	account.Branch = "new_branch"
	updatedAccount, err := accountRepository.Update(&account)

	// Assert
	assert.Nil(t, err)
	assert.Equal(t, "new_branch", updatedAccount.Branch)
}

func TestUpdateAccountNotFound(t *testing.T) {
	// Given
	account := givenAccount(true)
	account.Domain.ID = "non_existent_domain_id"

	// Test
	_, err := accountRepository.Update(&account)

	// Assert
	assert.NotNil(t, err)
}

func TestDeleteAccount(t *testing.T) {
	// Given
	account := givenAccount(true)
	accountRepository.Create(&account)

	// Test
	err := accountRepository.DeleteByDomainId(account.Domain.ID)

	// Assert
	assert.Nil(t, err)
}

func TestDeleteAccountNotFound(t *testing.T) {
	// Test
	err := accountRepository.DeleteByDomainId("non_existent_domain_id")

	// Assert
	assert.NotNil(t, err)
}
