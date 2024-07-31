package model

type DiffResult struct {
	Changes []DiffDetails `json:"changes"`
}

type DiffDetails struct {
	Action  string   `json:"action"`
	Diff    string   `json:"diff"`
	Path    []string `json:"path"`
	Content any      `json:"content"`
}
