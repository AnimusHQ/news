package publishing

import (
	"fmt"
	"strings"

	"github.com/AnimusHQ/news/internal/artifacts"
	"github.com/AnimusHQ/news/internal/productionqa"
	"github.com/AnimusHQ/news/internal/render"
	"github.com/AnimusHQ/news/internal/storyboard"
)

const (
	manifestSchemaVersion = "1.0"
	manifestStatus        = "draft"
	defaultCTA            = "Join the Animus open-source community and follow the source-backed production path."
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

// ReleasePackInput contains the approved production artifacts needed for a
// reviewable release metadata bundle.
type ReleasePackInput struct {
	EpisodeID                   string
	Title                       string
	Summary                     string
	Sources                     []artifacts.Source
	Claims                      []artifacts.Claim
	Storyboard                  storyboard.File
	RenderManifest              render.RenderManifest
	ProductionQA                productionqa.Report
	CTA                         string
	Visibility                  artifacts.PublishVisibility
	HumanApproved               bool
	Platform                    string
	DescriptionPath             string
	ThumbnailPath               string
	ScheduledAt                 string
	SyntheticDisclosureRequired bool
	SyntheticDisclosure         string
}

// Chapter describes a YouTube-style chapter marker.
type Chapter struct {
	Timestamp string `json:"timestamp"`
	Title     string `json:"title"`
}

// ReleasePack is the generated release bundle for human review.
type ReleasePack struct {
	Pack     Pack                 `json:"pack"`
	Manifest PublishManifestDraft `json:"publish_manifest"`
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

// PublishManifestDraft is the canonical publish_manifest.json draft shape.
type PublishManifestDraft struct {
	SchemaVersion               string                      `json:"schema_version"`
	EpisodeID                   string                      `json:"episode_id"`
	ArtifactID                  string                      `json:"artifact_id"`
	Status                      string                      `json:"status"`
	Platform                    string                      `json:"platform"`
	Visibility                  artifacts.PublishVisibility `json:"visibility"`
	Title                       string                      `json:"title"`
	DescriptionPath             string                      `json:"description_path"`
	ThumbnailPath               string                      `json:"thumbnail_path,omitempty"`
	ScheduledAt                 string                      `json:"scheduled_at,omitempty"`
	HumanReleaseApproval        bool                        `json:"human_release_approval"`
	SyntheticDisclosureRequired bool                        `json:"synthetic_disclosure_required,omitempty"`
	SyntheticDisclosure         string                      `json:"synthetic_disclosure,omitempty"`
}

// GenerateReleasePack creates a safe publish pack and manifest from approved
// production artifacts. It does not upload anything.
func GenerateReleasePack(input ReleasePackInput) (ReleasePack, error) {
	if input.ProductionQA.Decision != productionqa.DecisionApproved {
		return ReleasePack{}, fmt.Errorf("approved production QA is required before publish pack generation")
	}
	if len(input.Claims) > 0 && len(input.Sources) == 0 {
		return ReleasePack{}, fmt.Errorf("claims require a visible source list in the publish pack")
	}
	if len(input.Storyboard.Scenes) == 0 {
		return ReleasePack{}, fmt.Errorf("storyboard scenes are required for chapter generation")
	}
	if len(input.RenderManifest.Outputs) == 0 {
		return ReleasePack{}, fmt.Errorf("render manifest outputs are required")
	}
	if input.SyntheticDisclosureRequired && strings.TrimSpace(input.SyntheticDisclosure) == "" {
		return ReleasePack{}, fmt.Errorf("synthetic disclosure text is required")
	}

	visibility := input.Visibility
	if visibility == "" {
		visibility = artifacts.PublishVisibilityPrivate
	}
	cta := input.CTA
	if strings.TrimSpace(cta) == "" {
		cta = defaultCTA
	}
	pack, err := GeneratePack(PackInput{
		EpisodeID:     input.EpisodeID,
		Title:         input.Title,
		Summary:       input.Summary,
		Sources:       input.Sources,
		Chapters:      ChaptersFromStoryboard(input.Storyboard),
		CTA:           cta,
		Visibility:    visibility,
		HumanApproved: input.HumanApproved,
	})
	if err != nil {
		return ReleasePack{}, err
	}

	manifest := PublishManifestDraft{
		SchemaVersion:               manifestSchemaVersion,
		EpisodeID:                   input.EpisodeID,
		ArtifactID:                  "publish-manifest-" + input.EpisodeID + "-draft-v1",
		Status:                      manifestStatus,
		Platform:                    defaultText(input.Platform, "youtube"),
		Visibility:                  visibility,
		Title:                       pack.TitleCandidates[0],
		DescriptionPath:             defaultText(input.DescriptionPath, "dist/"+input.EpisodeID+"-description.md"),
		ThumbnailPath:               defaultText(input.ThumbnailPath, "dist/"+input.EpisodeID+"-thumbnail.png"),
		ScheduledAt:                 strings.TrimSpace(input.ScheduledAt),
		HumanReleaseApproval:        input.HumanApproved,
		SyntheticDisclosureRequired: input.SyntheticDisclosureRequired,
		SyntheticDisclosure:         strings.TrimSpace(input.SyntheticDisclosure),
	}
	return ReleasePack{Pack: pack, Manifest: manifest}, nil
}

// ChaptersFromStoryboard converts scene timing into deterministic chapter data.
func ChaptersFromStoryboard(file storyboard.File) []Chapter {
	chapters := make([]Chapter, 0, len(file.Scenes))
	for _, scene := range file.Scenes {
		timestamp := chapterTimestamp(scene.TimeTarget)
		title := strings.TrimSpace(scene.OnScreenText)
		if title == "" {
			title = strings.TrimSpace(scene.Visual.Type)
		}
		if timestamp == "" || title == "" {
			continue
		}
		chapters = append(chapters, Chapter{Timestamp: timestamp, Title: title})
	}
	return chapters
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
			input.Title + " - Explained from First Principles",
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
			fmt.Fprintf(&b, "- %s - %s\n", source.Title, source.URI)
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

func chapterTimestamp(timeTarget string) string {
	timeTarget = strings.TrimSpace(timeTarget)
	if timeTarget == "" {
		return ""
	}
	parts := strings.Split(timeTarget, "-")
	return strings.TrimSpace(parts[0])
}

func defaultText(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}
