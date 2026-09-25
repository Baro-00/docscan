package analyzer

type Result struct {
	Analyzer string `json:"analyzer"`
	Tool     string `json:"tool"`

	Command string   `json:"command"`
	Args    []string `json:"args"`

	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`

	ExitCode int  `json:"exitCode"`
	TimedOut bool `json:"timedOut"`

	Error string `json:"error,omitempty"`
}
