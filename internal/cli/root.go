// Package cli wires the agents-lint command tree and maps results to exit codes.
package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

// Execute runs the agents-lint command line with args and returns the process exit code.
func Execute(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	root := newRootCmd(stdout, stderr)
	root.SetArgs(args)
	if err := root.ExecuteContext(ctx); err != nil {
		_, _ = fmt.Fprintf(stderr, "agents-lint: %v\n", err)
		return ExitError
	}
	return ExitOK
}

func newRootCmd(stdout, stderr io.Writer) *cobra.Command {
	root := &cobra.Command{
		Use:           "agents-lint",
		Short:         "Lint AI-agent instruction files",
		Long:          "agents-lint checks AGENTS.md, CLAUDE.md, skills, agents and commands for common problems.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.AddCommand(newVersionCmd())
	return root
}
