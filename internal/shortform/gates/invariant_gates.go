package gates

import "github.com/AnimusHQ/news/internal/shortform"

// SelfApprovalGate enforces §4.6: an artifact whose creator is a model or the
// system cannot transition itself to approved. Approval requires a human (or, by
// policy, a distinct verifier identity) different from the creator.
type SelfApprovalInput struct {
	CreatedBy    string
	ApproverID   string
	TargetStatus string
	RequireHuman bool // when true, only human:* may approve
}

func SelfApprovalGate(in SelfApprovalInput) Result {
	e := newEval("self_approval")
	if in.TargetStatus != shortform.StatusApproved {
		// The gate only governs transitions into approved.
		return e.result()
	}
	e.require(present(in.ApproverID), "approver_missing", "approval requires a recorded approver identity", "approver")
	e.require(in.ApproverID != in.CreatedBy, "self_approval", "an artifact may not approve itself (approver equals creator)", "approver")
	requireHuman := in.RequireHuman || isMachineIdentity(in.CreatedBy)
	if requireHuman {
		e.require(isHuman(in.ApproverID), "non_human_approval", "generated output requires a human approver", "approver")
	}
	return e.result()
}

func isMachineIdentity(identity string) bool {
	if identity == "system" {
		return true
	}
	return hasPrefix(identity, "model:") || hasPrefix(identity, "system:")
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

// ImmutabilityGate enforces §4.9 / §7: approved or locked artifacts are immutable.
// A mutation is any content-hash change. Producing a new versioned artifact is
// the only legal path (callers create a fresh artifact, which has a new
// artifact_id and is not governed by this gate).
type ImmutabilityInput struct {
	CurrentStatus string
	OldHash       string
	NewHash       string
}

func ImmutabilityGate(in ImmutabilityInput) Result {
	e := newEval("immutability")
	frozen := in.CurrentStatus == shortform.StatusApproved || in.CurrentStatus == shortform.StatusLocked
	mutated := in.OldHash != in.NewHash
	e.require(!(frozen && mutated), "immutable_mutation", "approved or locked artifacts may not be mutated in place", "content_hash")
	return e.result()
}

// AIDisclosureGate enforces §4.8: if AI disclosure is required, release fails
// unless the disclosure is set correctly. It is reused by ReleaseGate.
type AIDisclosureInput struct {
	Required bool
	Text     string
	Present  bool
}

func AIDisclosureGate(in AIDisclosureInput) Result {
	e := newEval("ai_disclosure")
	if !in.Required {
		return e.result()
	}
	e.require(present(in.Text), "disclosure_text_missing", "AI disclosure text is required when disclosure is required", "ai_disclosure")
	e.require(in.Present, "disclosure_not_present", "AI disclosure must be marked present when required", "ai_disclosure_present")
	return e.result()
}

// MultiVerifierGate enforces §4.7: final sign-off requires at least N distinct
// verifier identities (default 2). No single model is the final authority.
type MultiVerifierInput struct {
	Verifiers []string
	Required  int
}

func MultiVerifierGate(in MultiVerifierInput) Result {
	e := newEval("multi_verifier")
	required := in.Required
	if required <= 0 {
		required = 2
	}
	distinct := map[string]bool{}
	for _, v := range in.Verifiers {
		if present(v) {
			distinct[v] = true
		}
	}
	e.require(len(distinct) >= required, "insufficient_verifiers", "final sign-off requires at least the configured number of distinct verifiers", "verifiers")
	return e.result()
}

// ClaimRef is a minimal claim provenance reference for the provenance gate.
type ClaimRef struct {
	ID        string
	SourceIDs []string
}

// ProvenanceGate enforces §4.1 and §4.2: no claim without a source, and no script
// reaching downstream use without a linked research pack. This guards the
// boundary where the short-form pipeline consumes approved long-form content.
type ProvenanceInput struct {
	ScriptApproved  bool
	ResearchPackRef string
	Claims          []ClaimRef
}

func ProvenanceGate(in ProvenanceInput) Result {
	e := newEval("provenance")
	if in.ScriptApproved {
		e.require(present(in.ResearchPackRef), "research_pack_missing", "an approved script requires a linked research pack", "research_pack_ref")
	}
	for _, claim := range in.Claims {
		e.require(len(claim.SourceIDs) > 0, "claim_without_source", "every claim must reference at least one source", claim.ID)
	}
	return e.result()
}
