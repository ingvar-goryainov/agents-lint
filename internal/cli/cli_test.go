package cli_test

import (
	"context"
	"os"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"

	"github.com/ingvar-goryainov/agents-lint/internal/cli"
)

// TestMain lets testscript run the real CLI as the "agents-lint" command inside scripts.
func TestMain(m *testing.M) {
	testscript.Main(m, map[string]func(){
		"agents-lint": func() {
			os.Exit(cli.Execute(context.Background(), os.Args[1:], os.Stdout, os.Stderr))
		},
	})
}

// TestScripts runs the CLI end-to-end scripts in testdata/script/*.txtar.
func TestScripts(t *testing.T) {
	t.Parallel()
	testscript.Run(t, testscript.Params{Dir: "testdata/script"})
}
