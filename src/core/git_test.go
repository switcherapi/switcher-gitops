package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/switcherapi/switcher-gitops/src/config"
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
	gitService := NewGitService(
		config.GetEnv("GIT_REPO_URL"),
		config.GetEnv("GIT_TOKEN"),
		config.GetEnv("GIT_BRANCH"))

	// Test
	lastCommit, date, content, err := gitService.GetRepositoryData("default")

	// Assert
	assert.Nil(t, err)
	assert.NotEmpty(t, lastCommit)
	assert.NotEmpty(t, date)
	assert.NotEmpty(t, content)
}

func TestGetRepositoryDataErrorInvalidEnvironment(t *testing.T) {
	// Given
	gitService := NewGitService(
		config.GetEnv("GIT_REPO_URL"),
		config.GetEnv("GIT_TOKEN"),
		config.GetEnv("GIT_BRANCH"))

	// Test
	lastCommit, date, content, err := gitService.GetRepositoryData("invalid")

	// Assert
	assert.NotNil(t, err)
	assert.Empty(t, lastCommit)
	assert.Empty(t, date)
	assert.Empty(t, content)
}

func TestGetRepositoryDataErrorInvalidToken(t *testing.T) {
	// Given
	gitService := NewGitService(
		config.GetEnv("GIT_REPO_URL"),
		"invalid",
		config.GetEnv("GIT_BRANCH"))

	// Test
	lastCommit, date, content, err := gitService.GetRepositoryData("default")

	// Assert
	assert.NotNil(t, err)
	assert.Empty(t, lastCommit)
	assert.Empty(t, date)
	assert.Empty(t, content)
}

func TestCheckForChanges(t *testing.T) {
	// Given
	gitService := NewGitService(
		config.GetEnv("GIT_REPO_URL"),
		config.GetEnv("GIT_TOKEN"),
		config.GetEnv("GIT_BRANCH"))

	account := givenAccount()
	lastCommit, date, content, _ := gitService.GetRepositoryData(account.Environment)

	// Test
	status, message := gitService.CheckForChanges(account, lastCommit, date, content)

	// Assert
	assert.Equal(t, model.StatusSynced, status)
	assert.Equal(t, "Synced successfully", message)
}
