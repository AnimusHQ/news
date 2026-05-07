package storyboard

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/AnimusHQ/news/internal/artifacts"
	"gopkg.in/yaml.v3"
)

const (
	SchemaVersion = "1.0"
	fileStatus    = "draft"
)

// Input contains the approved editorial material needed to generate a storyboard.
// The QA recommendation is required so storyboarding cannot silently bypass the
// human gate.
type Input struct {
	EpisodeID             string
	ArtifactID            string
	ScriptMarkdown        string
	HumanQARecommendation artifacts.HumanDecision
	Claims                []artifacts.Claim
}

// File is the canonical storyboard artifact shape.
type File struct {
	SchemaVersion string  `json:"schema_version" yaml:"schema_version"`
	EpisodeID     string  `json:"episode_id" yaml:"episode_id"`
	ArtifactID    string  `json:"artifact_id" yaml:"artifact_id"`
	Status        string  `json:"status" yaml:"status"`
	Scenes        []Scene `json:"scenes" yaml:"scenes"`
}

// Scene is one deterministic storyboard segment.
type Scene struct {
	SceneID      string     `json:"scene_id" yaml:"scene_id"`
	TimeTarget   string     `json:"time_target" yaml:"time_target"`
	Narration    string     `json:"narration" yaml:"narration"`
	Mascot       MascotPlan `json:"mascot" yaml:"mascot"`
	Visual       VisualPlan `json:"visual" yaml:"visual"`
	OnScreenText string     `json:"on_screen_text" yaml:"on_screen_text"`
	CaptionPlan  string     `json:"caption_plan" yaml:"caption_plan"`
	ClaimIDs     []string   `json:"claim_ids,omitempty" yaml:"claim_ids,omitempty"`
	SourceIDs    []string   `json:"source_ids,omitempty" yaml:"source_ids,omitempty"`
}

// MascotPlan describes deterministic mascot direction for the scene.
type MascotPlan struct {
	Mode    string `json:"mode" yaml:"mode"`
	Emotion string `json:"emotion" yaml:"emotion"`
	Action  string `json:"action" yaml:"action"`
}

// VisualPlan describes the deterministic visual plan for the scene.
type VisualPlan struct {
	Type    string `json:"type" yaml:"type"`
	Content string `json:"content" yaml:"content"`
}

// Generate converts an approved script into a storyboard without model calls.
func Generate(input Input) (File, error) {
	if strings.TrimSpace(input.EpisodeID) == "" {
		return File{}, fmt.Errorf("episode id is required")
	}
	if strings.TrimSpace(input.ScriptMarkdown) == "" {
		return File{}, fmt.Errorf("script markdown is required")
	}
	if !allowsStoryboarding(input.HumanQARecommendation) {
		return File{}, fmt.Errorf("human QA approval is required before storyboarding: %s", input.HumanQARecommendation)
	}

	narration := narrationSegments(input.ScriptMarkdown)
	if len(narration) == 0 {
		return File{}, fmt.Errorf("script contains no storyboardable narration")
	}

	file := File{
		SchemaVersion: SchemaVersion,
		EpisodeID:     input.EpisodeID,
		ArtifactID:    artifactID(input),
		Status:        fileStatus,
		Scenes:        make([]Scene, 0, len(narration)),
	}

	startSecond := 0
	for i, text := range narration {
		duration := estimateDurationSeconds(text, i == 0)
		claimIDs, sourceIDs := referencesFor(text, input.Claims)
		file.Scenes = append(file.Scenes, Scene{
			SceneID:      fmt.Sprintf("scene-%03d", i+1),
			TimeTarget:   timeTarget(startSecond, startSecond+duration),
			Narration:    text,
			Mascot:       mascotPlan(text, i == 0),
			Visual:       visualPlan(text),
			OnScreenText: onScreenText(text),
			CaptionPlan:  "captions_from_narration",
			ClaimIDs:     claimIDs,
			SourceIDs:    sourceIDs,
		})
		startSecond += duration
	}

	if err := ensureTechnicalClaimsCovered(input.Claims, file.Scenes); err != nil {
		return File{}, err
	}
	return file, nil
}

// MarshalYAML renders the storyboard artifact in canonical YAML form.
func MarshalYAML(file File) ([]byte, error) {
	return yaml.Marshal(file)
}

func allowsStoryboarding(decision artifacts.HumanDecision) bool {
	return decision == artifacts.HumanDecisionApprove || decision == artifacts.HumanDecisionApproveWithMinorEdits
}

func artifactID(input Input) string {
	if strings.TrimSpace(input.ArtifactID) != "" {
		return strings.TrimSpace(input.ArtifactID)
	}
	return "storyboard-" + input.EpisodeID + "-generated-v1"
}

func narrationSegments(markdown string) []string {
	var segments []string
	inFence := false
	for _, raw := range strings.Split(markdown, "\n") {
		line := strings.TrimSpace(raw)
		if strings.HasPrefix(line, "```") {
			inFence = !inFence
			continue
		}
		if inFence || line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = stripMarkdownPrefix(line)
		line = strings.TrimSpace(strings.ReplaceAll(line, "`", ""))
		if line == "" {
			continue
		}
		for _, part := range splitSentenceLine(line) {
			if text := normalizeNarration(part); text != "" {
				segments = append(segments, text)
			}
		}
	}
	return dedupe(segments)
}

var listPrefix = regexp.MustCompile(`^([-*]|\d+[.)])\s+`)

func stripMarkdownPrefix(line string) string {
	return listPrefix.ReplaceAllString(line, "")
}

func splitSentenceLine(line string) []string {
	var parts []string
	start := 0
	runes := []rune(line)
	for i, r := range runes {
		if r != '.' && r != '!' && r != '?' {
			continue
		}
		if i+1 < len(runes) && runes[i+1] != ' ' && runes[i+1] != '\t' {
			continue
		}
		parts = append(parts, string(runes[start:i+1]))
		start = i + 1
	}
	if start < len(runes) {
		parts = append(parts, string(runes[start:]))
	}
	if len(parts) == 0 {
		return []string{line}
	}
	return parts
}

func normalizeNarration(text string) string {
	text = strings.TrimSpace(text)
	text = strings.Trim(text, "-* ")
	return strings.TrimSuffix(text, ".")
}

func dedupe(items []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(items))
	for _, item := range items {
		key := strings.ToLower(item)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, item)
	}
	return out
}

func estimateDurationSeconds(text string, first bool) int {
	words := len(strings.Fields(text))
	duration := 5 + (words / 3)
	if first && duration < 8 {
		duration = 8
	}
	if duration < 6 {
		return 6
	}
	if duration > 18 {
		return 18
	}
	return duration
}

func timeTarget(start, end int) string {
	return fmt.Sprintf("%s-%s", clock(start), clock(end))
}

func clock(seconds int) string {
	return fmt.Sprintf("%d:%02d", seconds/60, seconds%60)
}

func referencesFor(text string, claims []artifacts.Claim) ([]string, []string) {
	lower := normalizeReferenceText(text)
	var claimIDs []string
	var sourceIDs []string
	for _, claim := range claims {
		if claim.ID == "" || claim.Text == "" {
			continue
		}
		if !strings.Contains(lower, normalizeReferenceText(claim.Text)) {
			continue
		}
		claimIDs = append(claimIDs, claim.ID)
		sourceIDs = append(sourceIDs, claim.SourceIDs...)
	}
	sort.Strings(claimIDs)
	sort.Strings(sourceIDs)
	return compact(claimIDs), compact(sourceIDs)
}

func normalizeReferenceText(text string) string {
	text = strings.ToLower(strings.TrimSpace(text))
	text = strings.TrimSuffix(text, ".")
	return text
}

func mascotPlan(text string, first bool) MascotPlan {
	lower := strings.ToLower(text)
	if first || strings.Contains(lower, "git push") {
		return MascotPlan{Mode: "Production Mode", Emotion: "curious", Action: "opens terminal"}
	}
	if containsAny(lower, "rollback", "incident", "production") {
		return MascotPlan{Mode: "Reliability Guide", Emotion: "focused", Action: "points at state transition"}
	}
	return MascotPlan{Mode: "Explainer", Emotion: "focused", Action: "points at pipeline diagram"}
}

func visualPlan(text string) VisualPlan {
	lower := strings.ToLower(text)
	switch {
	case containsAny(lower, "git push", "commit", "repository"):
		return VisualPlan{Type: "terminal_animation", Content: "git push origin main"}
	case containsAny(lower, "ci", "automation", "check"):
		return VisualPlan{Type: "pipeline_diagram", Content: "repository event -> CI checks"}
	case containsAny(lower, "artifact", "container", "image", "build"):
		return VisualPlan{Type: "artifact_flow", Content: "build -> artifact -> deployable package"}
	case containsAny(lower, "deploy", "deployment", "production"):
		return VisualPlan{Type: "deployment_diagram", Content: "artifact -> deployment strategy -> production"}
	case containsAny(lower, "rollback"):
		return VisualPlan{Type: "state_transition", Content: "bad release -> rollback -> stable state"}
	default:
		return VisualPlan{Type: "explainer_card", Content: text}
	}
}

func onScreenText(text string) string {
	words := strings.Fields(text)
	if len(words) > 7 {
		words = words[:7]
	}
	return strings.Trim(strings.Join(words, " "), ",:;")
}

func ensureTechnicalClaimsCovered(claims []artifacts.Claim, scenes []Scene) error {
	covered := map[string]bool{}
	for _, scene := range scenes {
		for _, claimID := range scene.ClaimIDs {
			covered[claimID] = true
		}
	}
	var missing []string
	for _, claim := range claims {
		if claim.ID == "" || claim.Type != "technical" {
			continue
		}
		if !covered[claim.ID] {
			missing = append(missing, claim.ID)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return fmt.Errorf("technical claims are not represented in storyboard scenes: %s", strings.Join(missing, ", "))
	}
	return nil
}

func compact(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	out := values[:0]
	var previous string
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || value == previous {
			continue
		}
		out = append(out, value)
		previous = value
	}
	return out
}

func containsAny(text string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(text, needle) {
			return true
		}
	}
	return false
}
