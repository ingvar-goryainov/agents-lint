package cli

// Process exit codes. These are a public contract for CI usage.
const (
	// ExitOK means the run completed with no findings at or above --fail-on.
	ExitOK = 0
	// ExitFindings means at least one finding is at or above --fail-on severity.
	ExitFindings = 1
	// ExitError means a usage, config or runtime error stopped the run.
	ExitError = 2
)
