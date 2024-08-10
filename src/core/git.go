package core

import (
	"time"

	"github.com/go-git/go-billy/v5/memfs"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/switcherapi/switcher-gitops/src/model"
)

type IGitService interface {
	GetRepositoryData(environment string) (*model.RepositoryData, error)
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

func (g *GitService) GetRepositoryData(environment string) (*model.RepositoryData, error) {
	commitHash, commitDate, content, err := g.getLastCommitData(model.FilePath + environment + ".json")

	if err != nil {
		return nil, err
	}

	return &model.RepositoryData{
		CommitHash: commitHash,
		CommitDate: commitDate.Format(time.ANSIC),
		Content:    content,
	}, nil
}

func (g *GitService) getLastCommitData(filePath string) (string, time.Time, string, error) {
	c, err := g.getCommitObject()

	if err != nil {
		return "", time.Time{}, "", err
	}

	// Get the tree from the commit object
	tree, _ := c.Tree()

	// Get the file
	f, err := tree.File(filePath)

	if err != nil {
		return "", time.Time{}, "", err
	}

	// Get the content
	content, _ := f.Contents()

	// Return the date of the commit
	return c.Hash.String(), c.Author.When, content, nil
}

func (g *GitService) getCommitObject() (*object.Commit, error) {
	r, err := g.getRepository()

	if err != nil {
		return nil, err
	}

	// Get the HEAD reference
	ref, _ := r.Head()

	// Get the commit object from the reference
	return r.CommitObject(ref.Hash())
}

func (g *GitService) getRepository() (*git.Repository, error) {
	// Clone repository using in-memory storage
	r, _ := git.Clone(memory.NewStorage(), memfs.New(), &git.CloneOptions{
		URL: g.RepoURL,
		Auth: &http.BasicAuth{
			Username: "git-user",
			Password: g.Token,
		},
	})

	// Checkout branch
	return g.checkoutBranch(*r)
}

func (g *GitService) checkoutBranch(r git.Repository) (*git.Repository, error) {
	// Fetch worktree
	w, err := r.Worktree()

	if err != nil {
		return nil, err
	}

	// Checkout remote branch
	err = w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewRemoteReferenceName("origin", g.BranchName),
	})

	return &r, err
}
