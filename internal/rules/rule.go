// Package rules defines the lint rule contract, findings and the explicit rule registry.
package rules

import (
	"fmt"

	"github.com/ingvar-goryainov/agents-lint/internal/document"
)

// Severity is how serious a finding is. Off disables a rule.
type Severity int

// Severities, ordered from least to most severe.
const (
	SeverityOff Severity = iota
	SeverityInfo
	SeverityWarning
	SeverityError
)

var severityNames = [...]string{"off", "info", "warning", "error"}

func (s Severity) String() string {
	if s < SeverityOff || int(s) >= len(severityNames) {
		return fmt.Sprintf("Severity(%d)", int(s))
	}
	return severityNames[s]
}

// ParseSeverity converts a config value (off|info|warning|error) into a Severity.
func ParseSeverity(v string) (Severity, error) {
	for i, name := range severityNames {
		if name == v {
			return Severity(i), nil
		}
	}
	return SeverityOff, fmt.Errorf("unknown severity %q (want off|info|warning|error)", v)
}

// Finding is one lint result. Findings are data, not errors.
type Finding struct {
	RuleID   string
	Severity Severity
	Path     string
	Line     int
	Col      int
	Message  string
}

// Rule is a single lint check. Rules are pure: they see only the parsed Document.
type Rule interface {
	// ID is the stable kebab-case rule identifier used in config; never renamed.
	ID() string
	Description() string
	DefaultSeverity() Severity
	Check(doc *document.Document) []Finding
}
