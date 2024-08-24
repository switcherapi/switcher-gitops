package core

import (
	"testing"
	"time"

	"github.com/go-git/go-billy/v5/memfs"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/stretchr/testify/assert"
	appConfig "github.com/switcherapi/switcher-gitops/src/config"
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
		appConfig.GetEnv("GIT_REPO_URL"),
		appConfig.GetEnv("GIT_TOKEN"),
		appConfig.GetEnv("GIT_BRANCH"))

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
		appConfig.GetEnv("GIT_REPO_URL"),
		appConfig.GetEnv("GIT_TOKEN"),
		appConfig.GetEnv("GIT_BRANCH"))

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
		appConfig.GetEnv("GIT_REPO_URL"),
		"invalid",
		appConfig.GetEnv("GIT_BRANCH"))

	// Test
	repositoryData, err := gitService.GetRepositoryData("default")

	// Assert
	assert.NotNil(t, err)
	assert.Nil(t, repositoryData)
}

func TestPushChanges(t *testing.T) {
	if !canRunIntegratedTests() {
		t.Skip(SkipMessage)
	}

	t.Run("Should push changes to repository", func(t *testing.T) {
		// Given
		branchName := "test-branch-" + time.Now().Format("20060102150405")
		createBranch(branchName)
		gitService := NewGitService(
			appConfig.GetEnv("GIT_REPO_URL"),
			appConfig.GetEnv("GIT_TOKEN"),
			branchName)

		// Test
		commitHash, err := gitService.PushChanges("default", "content")

		// Assert
		assert.Nil(t, err)
		assert.NotEmpty(t, commitHash)

		data, _ := gitService.GetRepositoryData("default")
		assert.Equal(t, commitHash, data.CommitHash)
		assert.Equal(t, "content", data.Content)

		// Clean up
		deleteBranch(branchName)
	})

	t.Run("Should fail to push changes to repository - no write access", func(t *testing.T) {
		// Given
		gitService := NewGitService(
			appConfig.GetEnv("GIT_REPO_URL"),
			appConfig.GetEnv("GIT_TOKEN_READ_ONLY"),
			appConfig.GetEnv("GIT_BRANCH"))

		// Test
		commitHash, err := gitService.PushChanges("default", "content")

		// Assert
		assert.NotNil(t, err)
		assert.Equal(t, "authorization failed", err.Error())
		assert.Empty(t, commitHash)
	})
}

// Helpers

func createBranch(branchName string) {
	repo, _ := git.Clone(memory.NewStorage(), memfs.New(), &git.CloneOptions{
		URL:  appConfig.GetEnv("GIT_REPO_URL"),
		Auth: getAuth(),
	})

	wt, _ := repo.Worktree()

	head, _ := repo.Head()

	wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branchName),
		Hash:   head.Hash(),
		Create: true,
	})

	repo.Push(&git.PushOptions{
		RemoteName: "origin",
		Auth:       getAuth(),
		RefSpecs:   []config.RefSpec{config.RefSpec("refs/heads/" + branchName + ":refs/heads/" + branchName)},
	})
}

func deleteBranch(branchName string) {
	r, _ := git.Clone(memory.NewStorage(), memfs.New(), &git.CloneOptions{
		URL:  appConfig.GetEnv("GIT_REPO_URL"),
		Auth: getAuth(),
	})

	r.Push(&git.PushOptions{
		Auth:     getAuth(),
		RefSpecs: []config.RefSpec{config.RefSpec(":refs/heads/" + branchName)},
	})
}

func getAuth() *http.BasicAuth {
	return &http.BasicAuth{
		Username: "git-user",
		Password: appConfig.GetEnv("GIT_TOKEN"),
	}
}
