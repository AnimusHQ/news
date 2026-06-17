// Package gates implements the short-form quality and release gates as pure,
// server-side functions: func(input) Result. Each gate returns a decision plus
// machine-readable reasons. Every blocking condition is covered by a
// failing-input test (see *_test.go) and each gate has a passing test.
//
// Gates are the enforcement surface for the §4 invariants and §8 gates of the
// M1 contract. They are deterministic and perform no I/O.
package gates

import "strings"

// Decision is the binary outcome of a gate.
type Decision string

const (
	// DecisionPass means the gate allows progression.
	DecisionPass Decision = "pass"
	// DecisionBlock means the gate halts progression.
	DecisionBlock Decision = "block"
)

// Reason is a machine-readable explanation for a gate decision.
type Reason struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

// Result is the output of a gate evaluation.
type Result struct {
	Gate     string   `json:"gate"`
	Decision Decision `json:"decision"`
	Reasons  []Reason `json:"reasons,omitempty"`
}

// Blocked reports whether the gate blocked progression.
func (r Result) Blocked() bool { return r.Decision == DecisionBlock }

// ProductionQAApproved is the only production QA decision that unblocks release.
const ProductionQAApproved = "approved"

// evaluator accumulates blocking reasons for a single gate.
type evaluator struct {
	gate    string
	reasons []Reason
}

func newEval(gate string) *evaluator { return &evaluator{gate: gate} }

// require records a blocking reason when cond is false.
func (e *evaluator) require(cond bool, code, message, field string) {
	if !cond {
		e.reasons = append(e.reasons, Reason{Code: code, Message: message, Field: field})
	}
}

func (e *evaluator) result() Result {
	if len(e.reasons) == 0 {
		return Result{Gate: e.gate, Decision: DecisionPass}
	}
	return Result{Gate: e.gate, Decision: DecisionBlock, Reasons: e.reasons}
}

func present(s string) bool { return strings.TrimSpace(s) != "" }

func isSHA256(s string) bool {
	return strings.HasPrefix(s, "sha256:") && len(s) == len("sha256:")+64
}

func isHuman(identity string) bool {
	return strings.HasPrefix(identity, "human:")
}
