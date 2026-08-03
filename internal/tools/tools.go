package tools

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

const defaultTimeoutSecs = 30

// ShellInput is the input schema for the shell execution tool.
type ShellInput struct {
	// Command is the shell command, CLI invocation, or Python one-liner to run.
	Command string `json:"command"`
	// TimeoutSecs is the maximum seconds to wait before killing the process.
	// Defaults to 30 when zero or negative.
	TimeoutSecs int `json:"timeout_secs,omitempty"`
}

// ShellOutput holds the result of a shell command execution.
type ShellOutput struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exit_code"`
	// TimedOut is true when the command was killed for exceeding TimeoutSecs.
	TimedOut bool `json:"timed_out"`
}

// RegisterShellTool defines and registers a genkit tool that runs arbitrary
// shell commands (including CLI tools and Python one-liners) via /bin/sh -c,
// captures stdout and stderr, and returns them along with the exit code.
//
// WARNING: commands execute with the same OS privileges as the running process.
// Only expose this tool inside trusted, sandboxed environments.
func RegisterShellTool(g *genkit.Genkit) *ai.ToolDef[ShellInput, ShellOutput] {
	return genkit.DefineTool(
		g,
		"shell_exec",
		"Execute a shell command, CLI tool invocation, common linux commands or "+
			"scripts and commands usually run with a shell interface such as "+
			"inline python code or one-liners, "+
			"generic python scripts saved on the filesystem "+
			"Commands run via /bin/sh -c. "+
			"Returns stdout, stderr, exit_code, and timed_out.",
		func(tctx *ai.ToolContext, input ShellInput) (ShellOutput, error) {
			timeout := input.TimeoutSecs
			if timeout <= 0 {
				timeout = defaultTimeoutSecs
			}

			ctx, cancel := context.WithTimeout(tctx, time.Duration(timeout)*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, "/bin/sh", "-c", input.Command)

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()

			out := ShellOutput{
				Stdout: stdout.String(),
				Stderr: stderr.String(),
			}

			if ctx.Err() != nil {
				out.TimedOut = true
				out.ExitCode = -1
				return out, nil
			}

			if err != nil {
				var exitErr *exec.ExitError
				if errors.As(err, &exitErr) {
					out.ExitCode = exitErr.ExitCode()
					return out, nil
				}
				return ShellOutput{}, fmt.Errorf("shell_exec: %w", err)
			}

			return out, nil
		},
	)
}
