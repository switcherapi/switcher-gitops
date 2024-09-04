package core

import (
	"time"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/switcherapi/switcher-gitops/src/config"
	"github.com/switcherapi/switcher-gitops/src/model"
)

type IGitService interface {
	GetRepositoryData(environment string) (*model.RepositoryData, error)
	PushChanges(environment string, content string) (string, error)
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

func (g *GitService) PushChanges(environment string, content string) (string, error) {
	// Create an in-memory file system
	fs := memfs.New()

	// Get the repository & Clone repository using in-memory storage
	r, _ := g.getRepository(fs)

	// Write the content to the in-memory file
	filePath := model.FilePath + environment + ".json"
	file, _ := fs.Create(filePath)
	file.Write([]byte(content))
	file.Close()

	// Get the worktree
	w, _ := r.Worktree()

	// Add the file to the worktree
	w.Add(filePath)

	// Commit the changes
	commit, _ := w.Commit("[switcher-gitops] updated "+environment+".json", g.createCommitOptions())

	// Push the changes
	err := r.Push(&git.PushOptions{
		Force: true,
		Auth:  g.getAuth(),
	})

	if err != nil {
		return "", err
	}

	return commit.String(), nil
}

func (g *GitService) createCommitOptions() *git.CommitOptions {
	return &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Switcher GitOps",
			Email: "switcher.gitops.no-reply@switcherapi.com",
			When:  time.Now(),
		},
	}
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
	r, err := g.getRepository(memfs.New())

	if err != nil {
		return nil, err
	}

	// Get the HEAD reference
	ref, _ := r.Head()

	// Get the commit object from the reference
	return r.CommitObject(ref.Hash())
}

func (g *GitService) getRepository(fs billy.Filesystem) (*git.Repository, error) {
	return git.Clone(memory.NewStorage(), fs, &git.CloneOptions{
		URL:           g.RepoURL,
		ReferenceName: plumbing.NewBranchReferenceName(g.BranchName),
		Auth:          g.getAuth(),
	})
}

func (g *GitService) getAuth() *http.BasicAuth {
	return &http.BasicAuth{
		Username: config.GetEnv("GIT_USER"),
		Password: g.Token,
	}
}
