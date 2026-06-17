package gates

import (
	"testing"

	"github.com/AnimusHQ/news/internal/shortform"
)

// ----- self-approval gate (§4.6) -----

func TestSelfApprovalGatePasses(t *testing.T) {
	assertPass(t, SelfApprovalGate(SelfApprovalInput{
		CreatedBy: "model:seedance-mock", ApproverID: "human:editor", TargetStatus: shortform.StatusApproved,
	}))
	// Not an approval transition -> not governed.
	assertPass(t, SelfApprovalGate(SelfApprovalInput{
		CreatedBy: "model:x", ApproverID: "model:x", TargetStatus: shortform.StatusInReview,
	}))
}

func TestSelfApprovalGateBlocks(t *testing.T) {
	assertBlocked(t, SelfApprovalGate(SelfApprovalInput{
		CreatedBy: "model:x", ApproverID: "", TargetStatus: shortform.StatusApproved,
	}), "approver_missing")

	assertBlocked(t, SelfApprovalGate(SelfApprovalInput{
		CreatedBy: "human:editor", ApproverID: "human:editor", TargetStatus: shortform.StatusApproved,
	}), "self_approval")

	// A model-created artifact approved by system (non-human).
	assertBlocked(t, SelfApprovalGate(SelfApprovalInput{
		CreatedBy: "model:x", ApproverID: "system", TargetStatus: shortform.StatusApproved,
	}), "non_human_approval")
}

// ----- immutability gate (§4.9) -----

func TestImmutabilityGatePasses(t *testing.T) {
	// Draft may change.
	assertPass(t, ImmutabilityGate(ImmutabilityInput{CurrentStatus: shortform.StatusDraft, OldHash: "a", NewHash: "b"}))
	// Approved unchanged is fine.
	assertPass(t, ImmutabilityGate(ImmutabilityInput{CurrentStatus: shortform.StatusApproved, OldHash: "a", NewHash: "a"}))
}

func TestImmutabilityGateBlocks(t *testing.T) {
	assertBlocked(t, ImmutabilityGate(ImmutabilityInput{CurrentStatus: shortform.StatusApproved, OldHash: "a", NewHash: "b"}), "immutable_mutation")
	assertBlocked(t, ImmutabilityGate(ImmutabilityInput{CurrentStatus: shortform.StatusLocked, OldHash: "a", NewHash: "b"}), "immutable_mutation")
}

// ----- AI disclosure gate (§4.8) -----

func TestAIDisclosureGatePasses(t *testing.T) {
	assertPass(t, AIDisclosureGate(AIDisclosureInput{Required: true, Text: "disclosed", Present: true}))
	assertPass(t, AIDisclosureGate(AIDisclosureInput{Required: false}))
}

func TestAIDisclosureGateBlocks(t *testing.T) {
	assertBlocked(t, AIDisclosureGate(AIDisclosureInput{Required: true, Text: "", Present: true}), "disclosure_text_missing")
	assertBlocked(t, AIDisclosureGate(AIDisclosureInput{Required: true, Text: "disclosed", Present: false}), "disclosure_not_present")
}

// ----- multi-verifier gate (§4.7) -----

func TestMultiVerifierGatePasses(t *testing.T) {
	assertPass(t, MultiVerifierGate(MultiVerifierInput{Verifiers: []string{"human:a", "model:b"}}))
}

func TestMultiVerifierGateBlocks(t *testing.T) {
	assertBlocked(t, MultiVerifierGate(MultiVerifierInput{Verifiers: []string{"human:a"}}), "insufficient_verifiers")
	// Duplicates collapse to one distinct verifier.
	assertBlocked(t, MultiVerifierGate(MultiVerifierInput{Verifiers: []string{"human:a", "human:a"}}), "insufficient_verifiers")
	assertBlocked(t, MultiVerifierGate(MultiVerifierInput{Verifiers: nil}), "insufficient_verifiers")
}

// ----- provenance gate (§4.1, §4.2) -----

func TestProvenanceGatePasses(t *testing.T) {
	assertPass(t, ProvenanceGate(ProvenanceInput{
		ScriptApproved:  true,
		ResearchPackRef: "research_pack.json",
		Claims:          []ClaimRef{{ID: "c1", SourceIDs: []string{"s1"}}},
	}))
}

func TestProvenanceGateBlocks(t *testing.T) {
	assertBlocked(t, ProvenanceGate(ProvenanceInput{ScriptApproved: true, ResearchPackRef: ""}), "research_pack_missing")
	assertBlocked(t, ProvenanceGate(ProvenanceInput{
		ScriptApproved:  true,
		ResearchPackRef: "research_pack.json",
		Claims:          []ClaimRef{{ID: "c1", SourceIDs: nil}},
	}), "claim_without_source")
}
