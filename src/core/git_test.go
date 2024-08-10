package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/switcherapi/switcher-gitops/src/config"
)

const (
	SkipMessage = "Skipping tests because environment variables are not set"
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
	if !canRunIntegratedTests() {
		t.Skip(SkipMessage)
	}

	// Given
	gitService := NewGitService(
		config.GetEnv("GIT_REPO_URL"),
		config.GetEnv("GIT_TOKEN"),
		config.GetEnv("GIT_BRANCH"))

	// Test
	repositoryData, err := gitService.GetRepositoryData("default")

	// Assert
	assert.Nil(t, err)
	assert.NotEmpty(t, repositoryData.CommitHash)
	assert.NotEmpty(t, repositoryData.CommitDate)
	assert.NotEmpty(t, repositoryData.Content)
}

func TestGetRepositoryDataErrorInvalidEnvironment(t *testing.T) {
	if !canRunIntegratedTests() {
		t.Skip(SkipMessage)
	}

	// Given
	gitService := NewGitService(
		config.GetEnv("GIT_REPO_URL"),
		config.GetEnv("GIT_TOKEN"),
		config.GetEnv("GIT_BRANCH"))

	// Test
	repositoryData, err := gitService.GetRepositoryData("invalid")

	// Assert
	assert.NotNil(t, err)
	assert.Nil(t, repositoryData)
}

func TestGetRepositoryDataErrorInvalidToken(t *testing.T) {
	if !canRunIntegratedTests() {
		t.Skip(SkipMessage)
	}

	// Given
	gitService := NewGitService(
		config.GetEnv("GIT_REPO_URL"),
		"invalid",
		config.GetEnv("GIT_BRANCH"))

	// Test
	repositoryData, err := gitService.GetRepositoryData("default")

	// Assert
	assert.NotNil(t, err)
	assert.Nil(t, repositoryData)
}
