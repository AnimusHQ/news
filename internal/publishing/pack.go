package publishing

import (
	"fmt"
	"strings"

	"github.com/AnimusHQ/news/internal/artifacts"
)

// PackInput contains source material for publication metadata generation.
type PackInput struct {
	EpisodeID     string
	Title         string
	Summary       string
	Sources       []artifacts.Source
	Chapters      []Chapter
	CTA           string
	Visibility    artifacts.PublishVisibility
	HumanApproved bool
}

// Chapter describes a YouTube-style chapter marker.
type Chapter struct {
	Timestamp string `json:"timestamp"`
	Title     string `json:"title"`
}

// Pack is the safe, reviewable publication metadata bundle.
type Pack struct {
	EpisodeID       string                      `json:"episode_id"`
	TitleCandidates []string                    `json:"title_candidates"`
	Description     string                      `json:"description"`
	PinnedComment   string                      `json:"pinned_comment"`
	CommunityPost   string                      `json:"community_post"`
	Visibility      artifacts.PublishVisibility `json:"visibility"`
	HumanApproved   bool                        `json:"human_release_approval"`
	Warnings        []string                    `json:"warnings,omitempty"`
}

// GeneratePack creates safe publication metadata. It never defaults to public
// visibility and refuses public visibility without explicit human approval.
func GeneratePack(input PackInput) (Pack, error) {
	if input.EpisodeID == "" {
		return Pack{}, fmt.Errorf("episode id is required")
	}
	if strings.TrimSpace(input.Title) == "" {
		return Pack{}, fmt.Errorf("title is required")
	}
	visibility := input.Visibility
	if visibility == "" {
		visibility = artifacts.PublishVisibilityPrivate
	}
	if visibility == artifacts.PublishVisibilityPublic && !input.HumanApproved {
		return Pack{}, fmt.Errorf("public visibility requires explicit human release approval")
	}

	pack := Pack{
		EpisodeID: input.EpisodeID,
		TitleCandidates: []string{
			input.Title,
			input.Title + " — Explained from First Principles",
			"What Really Happens After " + strings.TrimPrefix(input.Title, "What Happens After "),
		},
		Description:   buildDescription(input),
		PinnedComment: buildPinnedComment(input),
		CommunityPost: buildCommunityPost(input),
		Visibility:    visibility,
		HumanApproved: input.HumanApproved,
	}
	if len(input.Sources) == 0 {
		pack.Warnings = append(pack.Warnings, "no sources supplied for publish pack")
	}
	if visibility != artifacts.PublishVisibilityPrivate && visibility != artifacts.PublishVisibilityScheduled && visibility != artifacts.PublishVisibilityPublic {
		return Pack{}, fmt.Errorf("unsupported visibility: %s", visibility)
	}
	return pack, nil
}

func buildDescription(input PackInput) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n\n", input.Summary)
	if len(input.Chapters) > 0 {
		fmt.Fprintf(&b, "Chapters:\n")
		for _, chapter := range input.Chapters {
			fmt.Fprintf(&b, "%s %s\n", chapter.Timestamp, chapter.Title)
		}
		fmt.Fprintf(&b, "\n")
	}
	if len(input.Sources) > 0 {
		fmt.Fprintf(&b, "Sources:\n")
		for _, source := range input.Sources {
			fmt.Fprintf(&b, "- %s — %s\n", source.Title, source.URI)
		}
		fmt.Fprintf(&b, "\n")
	}
	if input.CTA != "" {
		fmt.Fprintf(&b, "%s\n", input.CTA)
	}
	return strings.TrimSpace(b.String())
}

func buildPinnedComment(input PackInput) string {
	if input.CTA == "" {
		return "What should we explain next from the production engineering world?"
	}
	return input.CTA
}

func buildCommunityPost(input PackInput) string {
	return fmt.Sprintf("New Animus News draft: %s. This episode is prepared through a source-grounded, multimodel-reviewed pipeline.", input.Title)
}
