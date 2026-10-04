package rules

// All returns every built-in rule, in a stable order.
//
// Rules are listed explicitly (no init() registration) so the rule set is greppable and deterministic.
// Add a new rule as one line here.
func All() []Rule {
	return []Rule{}
}
