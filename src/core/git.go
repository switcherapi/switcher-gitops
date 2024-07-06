package core

import (
	"time"

	"github.com/switcherapi/switcher-gitops/src/model"
)

type IGitService interface {
	GetRepositoryData() (string, string, string)
	CheckForChanges(account model.Account, lastCommit string, date string, content string) (status string, message string)
}

type GitService struct {
	RepoURL    string
	Token      string
	BranchName string
}

func NewGitService(repoURL string, token string, branchName string) *GitService {
	return &GitService{
		RepoURL:    repoURL,
		Token:      token,
		BranchName: branchName,
	}
}

func (g *GitService) GetRepositoryData() (string, string, string) {
	lastCommit := "123"
	date := time.Now().Format(time.ANSIC)
	content := "Content"

	return lastCommit, date, content
}

func (g *GitService) CheckForChanges(account model.Account, lastCommit string,
	date string, content string) (status string, message string) {
	return model.StatusSynced, "Synced successfully"
}
