package repository

import (
	"context"
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
	t.Run("Should fetch an account by ID", func(t *testing.T) {
		// Given
		account := givenAccount(true)
		account.Domain.ID = "123-fetch-account-by-id"
		accountCreated, _ := accountRepository.Create(&account)

		// Test
		fetchedAccount, err := accountRepository.FetchByAccountId(accountCreated.ID.Hex())

		// // Assert
		assert.Nil(t, err)
		assert.NotNil(t, fetchedAccount.ID)
	})

	t.Run("Should not fetch an account by ID - not found", func(t *testing.T) {
		// Test
		_, err := accountRepository.FetchByAccountId("non_existent_id")

		// Assert
		assert.NotNil(t, err)
	})

	t.Run("Should fetch an account by domain ID", func(t *testing.T) {
		// Given
		account := givenAccount(true)
		account.Domain.ID = "123-fetch-account-by-domain-id"
		accountCreated, _ := accountRepository.Create(&account)

		// Test
		fetchedAccount, err := accountRepository.FetchByDomainId(accountCreated.Domain.ID)

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
		// Drop collection
		mongoDb.Collection("accounts").Drop(context.Background())

		// Given
		account1 := givenAccount(true)
		account1.Domain.ID = "123-fetch-all-active-accounts"
		account2 := givenAccount(false)
		account2.Domain.ID = "123-fetch-all-active-accounts-2"

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
		account.Domain.ID = "123-update-account"
		accountCreated, _ := accountRepository.Create(&account)

		// Test
		accountCreated.Branch = "new_branch"
		updatedAccount, err := accountRepository.Update(accountCreated)

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
	t.Run("Should delete an account by account ID", func(t *testing.T) {
		// Given
		account := givenAccount(true)
		account.Domain.ID = "123-delete-account-by-account-id"
		accountCreated, _ := accountRepository.Create(&account)

		// Test
		err := accountRepository.DeleteByAccountId(accountCreated.ID.Hex())

		// Assert
		assert.Nil(t, err)
	})

	t.Run("Should not delete an account by account ID - not found", func(t *testing.T) {
		// Test
		err := accountRepository.DeleteByAccountId("non_existent_id")

		// Assert
		assert.NotNil(t, err)
		assert.Equal(t, "Account not found for id: non_existent_id", err.Error())
	})

	t.Run("Should delete an account by domain ID", func(t *testing.T) {
		// Given
		account := givenAccount(true)
		account.Domain.ID = "123-delete-account-by-domain-id"
		accountCreated, _ := accountRepository.Create(&account)

		// Test
		err := accountRepository.DeleteByDomainId(accountCreated.Domain.ID)

		// Assert
		assert.Nil(t, err)
	})

	t.Run("Should not delete an account by domain ID - not found", func(t *testing.T) {
		// Test
		err := accountRepository.DeleteByDomainId("non_existent_domain_id")

		// Assert
		assert.NotNil(t, err)
		assert.Equal(t, "Account not found for domain.id: non_existent_domain_id", err.Error())
	})
}
