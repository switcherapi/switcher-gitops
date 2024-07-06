package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/switcherapi/switcher-gitops/src/model"
)

func TestNewGitService(t *testing.T) {
	// Given
	repoURL := "repoURL"
	token := "token"
	branchName := "main"

	// Test
	gitService := NewGitService(repoURL, token, branchName)

	// Assert
	assert.Equal(t, repoURL, gitService.RepoURL)
	assert.Equal(t, token, gitService.Token)
	assert.Equal(t, branchName, gitService.BranchName)
}

func TestGetRepositoryData(t *testing.T) {
	// Given
	gitService := NewGitService("repoURL", "token", "main")

	// Test
	lastCommit, date, content := gitService.GetRepositoryData()

	// Assert
	assert.NotEmpty(t, lastCommit)
	assert.NotEmpty(t, date)
	assert.NotEmpty(t, content)
}

func TestCheckForChanges(t *testing.T) {
	// Given
	gitService := NewGitService("repoURL", "token", "main")
	account := givenAccount()
	lastCommit, date, content := gitService.GetRepositoryData()

	// Test
	status, message := gitService.CheckForChanges(account, lastCommit, date, content)

	// Assert
	assert.Equal(t, model.StatusSynced, status)
	assert.Equal(t, "Synced successfully", message)
}
