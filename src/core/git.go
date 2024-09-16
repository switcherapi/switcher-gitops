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
	"github.com/switcherapi/switcher-gitops/src/utils"
)

type IGitService interface {
	GetRepositoryData(environment string) (*model.RepositoryData, error)
	PushChanges(environment string, content string) (string, error)
	UpdateRepositorySettings(repository string, encryptedToken string, branch string)
}

type GitService struct {
	repoURL        string
	encryptedToken string
	branchName     string
}

func NewGitService(repoURL string, encryptedToken string, branchName string) *GitService {
	return &GitService{
		repoURL:        repoURL,
		encryptedToken: encryptedToken,
		branchName:     branchName,
	}
}

func (g *GitService) UpdateRepositorySettings(repository string, encryptedToken string, branch string) {
	g.repoURL = repository
	g.encryptedToken = encryptedToken
	g.branchName = branch
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
		URL:           g.repoURL,
		ReferenceName: plumbing.NewBranchReferenceName(g.branchName),
		Auth:          g.getAuth(),
	})
}

func (g *GitService) getAuth() *http.BasicAuth {
	decryptedToken, err := utils.Decrypt(g.encryptedToken, config.GetEnv("GIT_TOKEN_PRIVATE_KEY"))

	if err != nil || decryptedToken == "" {
		return nil
	}

	return &http.BasicAuth{
		Username: config.GetEnv("GIT_USER"),
		Password: decryptedToken,
	}
}
