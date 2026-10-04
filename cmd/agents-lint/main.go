// Command agents-lint lints AI-agent instruction files (AGENTS.md, CLAUDE.md, skills, agents).
package main

import (
	"context"
	"os"

	"github.com/ingvar-goryainov/agents-lint/internal/cli"
)

func main() {
	os.Exit(cli.Execute(context.Background(), os.Args[1:], os.Stdout, os.Stderr))
}
