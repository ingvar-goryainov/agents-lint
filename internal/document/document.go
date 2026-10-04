// Package document holds the typed model of a parsed agent instruction file.
//
// The parser and the remaining fields (Sections, positions) arrive with the first parsing feature;
// the scaffold only fixes the type the rules package depends on.
package document

// Kind classifies an agent instruction file (agents-md, claude-md, skill, agent, command, ...).
type Kind string

// Document is one parsed agent instruction file.
type Document struct {
	Path        string
	Kind        Kind
	Frontmatter map[string]any
	Body        string
}
