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

func TestFetchAccount(t *testing.T) {
	t.Run("Should fetch an account by domain ID", func(t *testing.T) {
		// Given
		account := givenAccount(true)
		accountRepository.Create(&account)

		// Test
		fetchedAccount, err := accountRepository.FetchByDomainId(account.Domain.ID)

		// Assert
		assert.Nil(t, err)
		assert.NotNil(t, fetchedAccount.ID)
	})

	t.Run("Should not fetch an account by domain ID - not found", func(t *testing.T) {
		// Test
		_, err := accountRepository.FetchByDomainId("non_existent_domain_id")

		// Assert
		assert.NotNil(t, err)
	})

	t.Run("Should fetch all active accounts", func(t *testing.T) {
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
	})
}

func TestUpdateAccount(t *testing.T) {
	t.Run("Should update an account", func(t *testing.T) {
		// Given
		account := givenAccount(true)
		accountRepository.Create(&account)

		// Test
		account.Branch = "new_branch"
		updatedAccount, err := accountRepository.Update(&account)

		// Assert
		assert.Nil(t, err)
		assert.Equal(t, "new_branch", updatedAccount.Branch)
	})

	t.Run("Should not update an account - not found", func(t *testing.T) {
		// Given
		account := givenAccount(true)
		account.Domain.ID = "non_existent_domain_id"

		// Test
		_, err := accountRepository.Update(&account)

		// Assert
		assert.NotNil(t, err)
	})
}

func TestDeleteAccount(t *testing.T) {
	t.Run("Should delete an account", func(t *testing.T) {
		// Given
		account := givenAccount(true)
		accountRepository.Create(&account)

		// Test
		err := accountRepository.DeleteByDomainId(account.Domain.ID)

		// Assert
		assert.Nil(t, err)
	})

	t.Run("Should not delete an account - not found", func(t *testing.T) {
		// Test
		err := accountRepository.DeleteByDomainId("non_existent_domain_id")

		// Assert
		assert.NotNil(t, err)
	})
}
