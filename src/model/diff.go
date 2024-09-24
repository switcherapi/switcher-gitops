package model

type DiffResult struct {
	Environment string        `json:"environment,omitempty"`
	Changes     []DiffDetails `json:"changes"`
}

type DiffDetails struct {
	Action  string   `json:"action"`
	Diff    string   `json:"diff"`
	Path    []string `json:"path"`
	Content any      `json:"content"`
}
