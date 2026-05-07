package claims

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/AnimusHQ/news/internal/artifacts"
)

const (
	TypeTechnical        = "technical"
	TypeHistorical       = "historical"
	TypeProduct          = "product"
	TypeSafety           = "safety"
	TypeEditorialOpinion = "editorial/opinion"
	TypeCTACommunity     = "CTA/community"
)

// Input contains the deterministic data needed to extract claim candidates.
type Input struct {
	EpisodeID      string
	ArtifactID     string
	ScriptMarkdown string
	ResearchPack   artifacts.ResearchPackFile
}

// Candidate records all claim-like statements considered by the extractor.
type Candidate struct {
	Text       string
	Type       string
	RiskLevel  artifacts.ClaimRisk
	SourceIDs  []string
	Included   bool
	SkipReason string
}

// Result contains the canonical claims artifact plus extraction diagnostics.
type Result struct {
	ClaimsFile       artifacts.ClaimsFile
	Candidates       []Candidate
	Warnings         []string
	UnlinkedClaimIDs []string
}

// ExtractEpisode loads script.md and research_pack.json from an episode directory
// and extracts a deterministic claims artifact without mutating files.
func ExtractEpisode(episodeDir string) (Result, error) {
	if episodeDir == "" {
		return Result{}, fmt.Errorf("episode directory is required")
	}
	researchPack, err := artifacts.LoadResearchPackFile(filepath.Join(episodeDir, "research_pack.json"))
	if err != nil {
		return Result{}, err
	}
	script, err := os.ReadFile(filepath.Join(episodeDir, "script.md"))
	if err != nil {
		return Result{}, fmt.Errorf("read script artifact: %w", err)
	}
	return Extract(Input{
		EpisodeID:      researchPack.EpisodeID,
		ArtifactID:     "claims-" + researchPack.EpisodeID + "-extracted-v1",
		ScriptMarkdown: string(script),
		ResearchPack:   researchPack,
	})
}

// Extract converts script markdown into a canonical claims artifact. It is
// intentionally conservative: it links source IDs when inferable, but it never
// fabricates evidence locators or marks extracted claims as supported.
func Extract(input Input) (Result, error) {
	if strings.TrimSpace(input.EpisodeID) == "" {
		return Result{}, fmt.Errorf("episode id is required")
	}
	if strings.TrimSpace(input.ArtifactID) == "" {
		return Result{}, fmt.Errorf("artifact id is required")
	}
	if strings.TrimSpace(input.ScriptMarkdown) == "" {
		return Result{}, fmt.Errorf("script markdown is required")
	}
	if len(input.ResearchPack.Sources) == 0 {
		return Result{}, fmt.Errorf("research pack must include sources")
	}

	sentences := extractSentences(input.ScriptMarkdown)
	result := Result{
		ClaimsFile: artifacts.ClaimsFile{
			SchemaVersion: "1.0",
			EpisodeID:     input.EpisodeID,
			ArtifactID:    input.ArtifactID,
			Status:        string(artifacts.ArtifactStatusDraft),
			Claims:        []artifacts.Claim{},
		},
	}

	for _, sentence := range sentences {
		candidate := classify(sentence, input.ResearchPack.Sources)
		if !candidate.Included {
			result.Candidates = append(result.Candidates, candidate)
			continue
		}
		claimID := fmt.Sprintf("claim-%03d", len(result.ClaimsFile.Claims)+1)
		if len(candidate.SourceIDs) == 0 && (candidate.Type == TypeTechnical || candidate.RiskLevel == artifacts.ClaimRiskHigh || candidate.RiskLevel == artifacts.ClaimRiskCritical) {
			result.UnlinkedClaimIDs = append(result.UnlinkedClaimIDs, claimID)
			result.Warnings = append(result.Warnings, fmt.Sprintf("%s is %s/%s but has no inferred source link", claimID, candidate.Type, candidate.RiskLevel))
		}
		result.ClaimsFile.Claims = append(result.ClaimsFile.Claims, artifacts.Claim{
			ID:               claimID,
			Text:             candidate.Text,
			Type:             candidate.Type,
			RiskLevel:        candidate.RiskLevel,
			SourceIDs:        candidate.SourceIDs,
			EvidenceLocators: nil,
			Status:           artifacts.ClaimStatusNeedsHumanReview,
		})
		result.Candidates = append(result.Candidates, candidate)
	}

	if len(result.ClaimsFile.Claims) == 0 {
		return Result{}, fmt.Errorf("no factual claims extracted")
	}
	return result, nil
}

func extractSentences(markdown string) []string {
	var sentences []string
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
			if normalized := normalizeSentence(part); normalized != "" {
				sentences = append(sentences, normalized)
			}
		}
	}
	return dedupe(sentences)
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

func normalizeSentence(text string) string {
	text = strings.TrimSpace(text)
	text = strings.Trim(text, "-* ")
	if text == "" {
		return ""
	}
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

func classify(sentence string, sources []artifacts.Source) Candidate {
	text := strings.TrimSpace(sentence)
	lower := strings.ToLower(text)
	claimType := classifyType(lower)
	risk := classifyRisk(lower, claimType)
	sourceIDs := inferSourceIDs(lower, sources)

	candidate := Candidate{
		Text:      text,
		Type:      claimType,
		RiskLevel: risk,
		SourceIDs: sourceIDs,
		Included:  true,
	}

	if isViewerHook(lower) {
		candidate.Included = false
		candidate.SkipReason = "viewer-facing hook is not a stable factual claim"
	}
	if claimType == TypeCTACommunity {
		candidate.Included = false
		candidate.SkipReason = "CTA/community sentence is not a factual claim"
	}
	if claimType == TypeEditorialOpinion && !looksFactualOpinion(lower) {
		candidate.Included = false
		candidate.SkipReason = "editorial opinion is not phrased as a verifiable fact"
	}
	if !containsFactualCue(lower) && claimType != TypeSafety {
		candidate.Included = false
		candidate.SkipReason = "no deterministic factual cue"
	}
	return candidate
}

func isViewerHook(lower string) bool {
	return strings.HasPrefix(lower, "you ") || strings.HasPrefix(lower, "your ") || strings.Contains(lower, "your code")
}

func classifyType(lower string) string {
	if containsAny(lower, "if you want", "follow ", "join ", "cta", "subscribe", "comment") {
		return TypeCTACommunity
	}
	if containsAny(lower, "i think", "we believe", "probably", "maybe", "should feel", "is better") {
		return TypeEditorialOpinion
	}
	if containsAny(lower, "credential", "secret", "private data", "unsafe", "policy", "security") {
		return TypeSafety
	}
	if containsAny(lower, "github", "docker", "kubernetes", "platform") {
		return TypeProduct
	}
	if containsAny(lower, "history", "historical", "first released", "created in") {
		return TypeHistorical
	}
	return TypeTechnical
}

func classifyRisk(lower string, claimType string) artifacts.ClaimRisk {
	if containsAny(lower, "credential", "secret", "private data", "malware", "exploit") {
		return artifacts.ClaimRiskCritical
	}
	if claimType == TypeSafety || containsAny(lower, "security", "incident", "data exposure") {
		return artifacts.ClaimRiskHigh
	}
	if claimType == TypeCTACommunity || claimType == TypeEditorialOpinion {
		return artifacts.ClaimRiskLow
	}
	return artifacts.ClaimRiskMedium
}

func inferSourceIDs(lower string, sources []artifacts.Source) []string {
	var ids []string
	for _, source := range sources {
		if source.ID == "" {
			continue
		}
		if sourceMatches(lower, source) {
			ids = append(ids, source.ID)
		}
	}
	sort.Strings(ids)
	return ids
}

func sourceMatches(lower string, source artifacts.Source) bool {
	key := strings.ToLower(source.ID + " " + source.Title + " " + source.URI)
	terms := sourceTerms(key)
	words := wordSet(lower)
	for _, term := range terms {
		if strings.Contains(term, " ") && strings.Contains(lower, term) {
			return true
		}
		if !strings.Contains(term, " ") && words[term] {
			return true
		}
	}
	return false
}

func wordSet(text string) map[string]bool {
	words := map[string]bool{}
	for _, token := range strings.FieldsFunc(text, func(r rune) bool {
		return r < 'a' || r > 'z'
	}) {
		if token != "" {
			words[token] = true
		}
	}
	return words
}

func sourceTerms(sourceKey string) []string {
	terms := []string{}
	if containsAny(sourceKey, "git-scm", "git-docs", "git documentation") {
		terms = append(terms, "git", "push", "commit", "remote repository", "repository")
	}
	if containsAny(sourceKey, "github", "actions") {
		terms = append(terms, "ci", "check", "checks", "automation", "workflow", "repository event", "pull request")
	}
	if containsAny(sourceKey, "docker") {
		terms = append(terms, "docker", "container", "container image", "image", "artifact", "build")
	}
	if containsAny(sourceKey, "kubernetes") {
		terms = append(terms, "kubernetes", "deployment", "deploy", "rollback", "workload", "production")
	}
	for _, token := range strings.FieldsFunc(sourceKey, func(r rune) bool {
		return r == '-' || r == '_' || r == '/' || r == '.' || r == ':' || r == ' '
	}) {
		if len(token) >= 4 && !isStopTerm(token) {
			terms = append(terms, token)
		}
	}
	slices.Sort(terms)
	return slices.Compact(terms)
}

func isStopTerm(token string) bool {
	switch token {
	case "https", "http", "docs", "documentation", "official", "source":
		return true
	default:
		return false
	}
}

func looksFactualOpinion(lower string) bool {
	return containsAny(lower, "is ", "are ", "has ", "have ", "uses ", "requires ")
}

func containsFactualCue(lower string) bool {
	return containsAny(lower,
		" is ", " are ", " has ", " have ", " uses ", " requires ", " sends ", " triggers ",
		" validates ", " produce ", " produces ", " may be ", " exists ", " tells ", " moves ",
		" can ", " leak ", " running ", " deployment", " repository", " commit", " ci", " artifact",
		" rollback", " observability", " production")
}

func containsAny(text string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(text, needle) {
			return true
		}
	}
	return false
}
