package analyzer

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"time"

	"docscan/internal/config"
)

type Variables struct {
	File  string
	Tools string
}

const defaultTimeout = 60 * time.Second

func Execute(
	ctx context.Context,
	step config.Analyzer,
	tool config.Tool,
	file string,
	toolsDir string,
) Result {

	args := expandArgs(
		step.Args,
		Variables{
			File:  file,
			Tools: toolsDir,
		},
	)

	result := Result{
		Analyzer: step.Name,
		Tool:     step.Tool,
		Command:  tool.Command,
		Args:     args,
		ExitCode: -1,
	}

	timeout := defaultTimeout

	if tool.Timeout > 0 {
		timeout = time.Duration(tool.Timeout) * time.Second
	}

	execCtx, cancel := context.WithTimeout(
		ctx,
		timeout,
	)

	defer cancel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.CommandContext(
		execCtx,
		tool.Command,
		args...,
	)

	cmd.Env = append(
		os.Environ(),
		"PYTHONUTF8=1",
		"PYTHONIOENCODING=utf-8",
	)

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result.Stdout = stdout.String()
	result.Stderr = stderr.String()

	if err == nil {
		result.ExitCode = 0
		return result
	}

	if errors.Is(execCtx.Err(), context.DeadlineExceeded) {
		result.TimedOut = true
		result.Error = "analysis timeout"

		return result
	}

	if exitError, ok := err.(*exec.ExitError); ok {
		result.ExitCode = exitError.ExitCode()
	}

	result.Error = err.Error()

	return result
}

func expandArgs(
	args []string,
	vars Variables,
) []string {

	result := make([]string, len(args))

	for i, arg := range args {
		arg = strings.ReplaceAll(
			arg,
			"{file}",
			vars.File,
		)

		arg = strings.ReplaceAll(
			arg,
			"{tools}",
			vars.Tools,
		)

		result[i] = arg
	}

	return result
}
