package rules_test

import (
	"regexp"
	"testing"

	"github.com/ingvar-goryainov/agents-lint/internal/rules"
)

var kebabCase = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

func TestAllRulesHaveUniqueKebabCaseIDs(t *testing.T) {
	t.Parallel()
	seen := map[string]bool{}
	for _, r := range rules.All() {
		id := r.ID()
		if !kebabCase.MatchString(id) {
			t.Errorf("rule ID %q is not kebab-case", id)
		}
		if seen[id] {
			t.Errorf("duplicate rule ID %q", id)
		}
		seen[id] = true
		if r.Description() == "" {
			t.Errorf("rule %q has no description", id)
		}
	}
}

func TestParseSeverity(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in      string
		want    rules.Severity
		wantErr bool
	}{
		{in: "off", want: rules.SeverityOff},
		{in: "info", want: rules.SeverityInfo},
		{in: "warning", want: rules.SeverityWarning},
		{in: "error", want: rules.SeverityError},
		{in: "fatal", wantErr: true},
		{in: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()
			got, err := rules.ParseSeverity(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseSeverity(%q) error = %v, wantErr %v", tt.in, err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseSeverity(%q) = %v, want %v", tt.in, got, tt.want)
			}
			if !tt.wantErr && got.String() != tt.in {
				t.Errorf("Severity(%d).String() = %q, want %q", got, got.String(), tt.in)
			}
		})
	}
}
