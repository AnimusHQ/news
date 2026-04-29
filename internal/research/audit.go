package research

import (
	"fmt"

	"github.com/AnimusHQ/news/internal/artifacts"
	"github.com/AnimusHQ/news/internal/sources"
)

// Pack is the subset of research_pack.json needed for deterministic audits.
type Pack struct {
	CoreQuestion              string             `json:"core_question" yaml:"core_question"`
	Sources                   []artifacts.Source `json:"sources" yaml:"sources"`
	LearningObjectives        []string           `json:"learning_objectives" yaml:"learning_objectives"`
	ForbiddenSimplifications  []string           `json:"forbidden_simplifications" yaml:"forbidden_simplifications"`
	VisualOpportunities       []string           `json:"visual_opportunities" yaml:"visual_opportunities"`
}

// AuditResult summarizes deterministic research quality checks.
type AuditResult struct {
	Valid    bool     `json:"valid"`
	Warnings []string `json:"warnings,omitempty"`
	Blockers []string `json:"blockers,omitempty"`
}

// AuditPack checks research pack completeness and source authority.
func AuditPack(pack Pack, claims []artifacts.Claim) (AuditResult, error) {
	result := AuditResult{Valid: true}
	if pack.CoreQuestion == "" {
		result.Valid = false
		result.Blockers = append(result.Blockers, "core question is required")
	}
	if len(pack.Sources) == 0 {
		result.Valid = false
		result.Blockers = append(result.Blockers, "at least one source is required")
		return result, nil
	}
	registry, err := sources.NewRegistry(pack.Sources)
	if err != nil {
		return AuditResult{}, fmt.Errorf("source registry audit failed: %w", err)
	}
	if len(pack.LearningObjectives) == 0 {
		result.Warnings = append(result.Warnings, "no learning objectives defined")
	}
	if len(pack.ForbiddenSimplifications) == 0 {
		result.Warnings = append(result.Warnings, "no forbidden simplifications defined")
	}
	if len(pack.VisualOpportunities) == 0 {
		result.Warnings = append(result.Warnings, "no visual opportunities defined")
	}

	for _, claim := range claims {
		if claim.RiskLevel == artifacts.ClaimRiskHigh || claim.RiskLevel == artifacts.ClaimRiskCritical {
			if !registry.SatisfiesHighRiskAuthority(claim.SourceIDs) {
				result.Valid = false
				result.Blockers = append(result.Blockers, fmt.Sprintf("high-risk claim %s lacks primary source authority", claim.ID))
			}
		}
	}
	return result, nil
}
