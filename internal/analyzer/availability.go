package analyzer

import (
	"os/exec"
	"sort"

	"docscan/internal/config"
)

type ToolStatus struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Command string `json:"command"`

	Available bool   `json:"available"`
	Path      string `json:"path,omitempty"`
	Error     string `json:"error,omitempty"`
}

func CheckTools(cfg *config.Config) []ToolStatus {
	names := make([]string, 0, len(cfg.Tools))

	for name := range cfg.Tools {
		names = append(names, name)
	}

	sort.Strings(names)

	results := make(
		[]ToolStatus,
		0,
		len(names),
	)

	for _, name := range names {
		results = append(
			results,
			CheckTool(name, cfg.Tools[name]),
		)
	}

	return results
}

func CheckTool(
	name string,
	tool config.Tool,
) ToolStatus {

	status := ToolStatus{
		Name:    name,
		Type:    tool.Type,
		Command: tool.Command,
	}

	switch tool.Type {
	case "command":
		path, err := exec.LookPath(tool.Command)

		if err != nil {
			status.Error = err.Error()
			return status
		}

		status.Available = true
		status.Path = path

		return status

	default:
		status.Error = "unsupported tool type"
		return status
	}
}
