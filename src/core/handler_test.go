package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitCoreHandlerCoroutine(t *testing.T) {
	// Given
	account1 := givenAccount()
	coreHandler.AccountRepository.Create(&account1)

	// Test
	status, err := coreHandler.InitCoreHandlerCoroutine()

	// Assert
	assert.Nil(t, err)
	assert.Equal(t, 1, status)
}
