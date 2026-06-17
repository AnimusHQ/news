package shortform

import (
	"fmt"
	"time"
)

// Approval transforms move draft artifacts to approved/locked states and re-stamp
// the content hash. They represent operator/human approval applied as a pipeline
// step. They never set created_by to the approver (the creator is preserved so
// the self-approval gate can reason about provenance).

func rfc3339(now time.Time) string {
	if now.IsZero() {
		now = time.Unix(0, 0).UTC()
	}
	return now.UTC().Format(time.RFC3339)
}

// ApproveStoryboardImages marks every image approved with visual review passed,
// records the approver, and re-stamps. The creator (created_by) is unchanged.
func ApproveStoryboardImages(m *StoryboardImageManifest, approver string, now time.Time) error {
	if m == nil {
		return fmt.Errorf("storyboard image manifest is required")
	}
	at := rfc3339(now)
	for i := range m.Images {
		m.Images[i].Status = StatusApproved
		m.Images[i].VisualReviewPassed = true
		m.Images[i].ApprovedBy = approver
		m.Images[i].ApprovedAt = at
	}
	m.Status = StatusApproved
	return Stamp(m)
}

// ApproveVisualShots records operator approval on every shot and re-stamps.
func ApproveVisualShots(m *VisualShotManifest, now time.Time) error {
	if m == nil {
		return fmt.Errorf("visual shot manifest is required")
	}
	for i := range m.Shots {
		m.Shots[i].OperatorApproval = true
	}
	m.Status = StatusApproved
	return Stamp(m)
}

// ApproveVoiceover records operator approval and re-stamps.
func ApproveVoiceover(m *VoiceoverManifest, now time.Time) error {
	if m == nil {
		return fmt.Errorf("voiceover manifest is required")
	}
	m.OperatorApproval = true
	m.Status = StatusApproved
	return Stamp(m)
}

// ApproveSubtitles records operator approval and re-stamps.
func ApproveSubtitles(m *SubtitleManifest, now time.Time) error {
	if m == nil {
		return fmt.Errorf("subtitle manifest is required")
	}
	m.OperatorApproval = true
	m.Status = StatusApproved
	return Stamp(m)
}

// ApproveRenderOutputs marks render outputs approved and re-stamps.
func ApproveRenderOutputs(m *ShortRenderManifest, now time.Time) error {
	if m == nil {
		return fmt.Errorf("short render manifest is required")
	}
	for i := range m.Outputs {
		m.Outputs[i].Status = StatusApproved
	}
	m.Status = StatusApproved
	return Stamp(m)
}

// ComponentRef is a concrete, serializable descriptor of an approved artifact to
// be pinned into a production candidate. It is used at the activity boundary so
// no interface-typed value crosses Temporal's data converter.
type ComponentRef struct {
	ArtifactID  string `json:"artifact_id"`
	Kind        string `json:"kind"`
	ContentHash string `json:"content_hash"`
	Status      string `json:"status"`
}

// ComponentRefOf builds a ComponentRef from a stamped artifact.
func ComponentRefOf(a Artifact) ComponentRef {
	env := a.EnvelopeRef()
	return ComponentRef{ArtifactID: env.ArtifactID, Kind: a.Kind(), ContentHash: env.ContentHash, Status: env.Status}
}

// AssembleProductionCandidate builds a locked, immutable bundle pointer from the
// approved upstream artifacts. The candidate is locked (immutable) on creation.
func AssembleProductionCandidate(episodeID, candidateID string, now time.Time, components []ComponentRef) (*ProductionCandidate, error) {
	if len(components) == 0 {
		return nil, fmt.Errorf("production candidate requires at least one component")
	}
	pinned := make([]CandidateComponent, 0, len(components))
	sources := make([]string, 0, len(components))
	for _, c := range components {
		if c.ContentHash == "" {
			return nil, fmt.Errorf("component %s must be stamped before bundling", c.Kind)
		}
		if c.Status != StatusApproved && c.Status != StatusLocked {
			return nil, fmt.Errorf("component %s must be approved before bundling (status %q)", c.Kind, c.Status)
		}
		pinned = append(pinned, CandidateComponent{ArtifactID: c.ArtifactID, Kind: c.Kind, ContentHash: c.ContentHash})
		sources = append(sources, c.ArtifactID)
	}
	candidate := &ProductionCandidate{
		Envelope: Envelope{
			SchemaVersion:   SchemaVersion,
			EpisodeID:       episodeID,
			ArtifactID:      fmt.Sprintf("%s-%s-v1", KindProductionCandidate, episodeID),
			CreatedAt:       rfc3339(now),
			CreatedBy:       "system",
			SourceArtifacts: sources,
			Status:          StatusLocked,
		},
		CandidateID: candidateID,
		Immutable:   true,
		Components:  pinned,
	}
	return candidate, Stamp(candidate)
}

// BuildReleaseApprovalInput carries the human release decision.
type BuildReleaseApprovalInput struct {
	EpisodeID            string
	CandidateID          string
	Approver             string
	Now                  time.Time
	Platforms            []string
	Visibility           string
	ScheduledAt          string
	AIDisclosureRequired bool
	AIDisclosure         string
	ProductionQARef      string
}

// BuildReleaseApproval produces an approved release_approval artifact authored by
// the human approver. It does not itself enforce gates (the ReleaseGate does).
func BuildReleaseApproval(in BuildReleaseApprovalInput) (*ReleaseApproval, error) {
	if !isHumanIdentity(in.Approver) {
		return nil, fmt.Errorf("release approval requires a human approver, got %q", in.Approver)
	}
	visibility := in.Visibility
	if visibility == "" {
		visibility = "private"
	}
	ra := &ReleaseApproval{
		Envelope: Envelope{
			SchemaVersion: SchemaVersion,
			EpisodeID:     in.EpisodeID,
			ArtifactID:    fmt.Sprintf("%s-%s-v1", KindReleaseApproval, in.EpisodeID),
			CreatedAt:     rfc3339(in.Now),
			CreatedBy:     in.Approver,
			Status:        StatusApproved,
		},
		CandidateID:          in.CandidateID,
		Platforms:            in.Platforms,
		Visibility:           visibility,
		ScheduledAt:          in.ScheduledAt,
		AIDisclosureRequired: in.AIDisclosureRequired,
		AIDisclosure:         in.AIDisclosure,
		HumanReleaseApproval: true,
		ApprovedBy:           in.Approver,
		ApprovedAt:           rfc3339(in.Now),
		ProductionQARef:      in.ProductionQARef,
		RiskAcceptance: RiskAcceptance{
			AIGeneratedVisuals:  true,
			AIDisclosurePresent: in.AIDisclosure != "",
			BrandSafetyChecked:  true,
		},
	}
	return ra, Stamp(ra)
}

func isHumanIdentity(identity string) bool {
	return len(identity) > len("human:") && identity[:len("human:")] == "human:"
}
