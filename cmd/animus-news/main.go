package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AnimusHQ/news/internal/artifacts"
	claimextractor "github.com/AnimusHQ/news/internal/claims"
	"github.com/AnimusHQ/news/internal/pipeline"
	"github.com/AnimusHQ/news/internal/security"
	"github.com/AnimusHQ/news/internal/shortform"
	"github.com/AnimusHQ/news/internal/shortform/pilot"
	"github.com/AnimusHQ/news/internal/shortform/providers/capabilities"
	"github.com/AnimusHQ/news/internal/shortform/runner"
	"github.com/AnimusHQ/news/internal/temporalops"
	"github.com/AnimusHQ/news/internal/worker"
)

func main() {
	if err := run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) < 2 {
		printUsage()
		return nil
	}

	ctx := context.Background()
	switch args[1] {
	case "validate":
		jsonOutput, path, err := parseValidateArgs(args[2:])
		if err != nil {
			return err
		}
		report := artifacts.ValidatePath(path)
		if jsonOutput {
			encoded, err := json.MarshalIndent(report, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(encoded))
		} else if report.Valid {
			fmt.Printf("valid: %s\n", path)
		}
		if !report.Valid {
			return artifacts.ValidateReport(report)
		}
		return nil
	case "demo":
		return runDemo(ctx, args[2:])
	case "validate-shortform":
		if len(args) != 3 {
			return fmt.Errorf("usage: animus-news validate-shortform <artifact-file>")
		}
		issues := shortform.ValidateFile(args[2])
		if len(issues) == 0 {
			fmt.Printf("valid short-form artifact: %s\n", args[2])
			return nil
		}
		for _, issue := range issues {
			fmt.Fprintf(os.Stderr, "invalid: %s\n", issue)
		}
		return fmt.Errorf("short-form artifact validation failed: %d issue(s)", len(issues))
	case "validate-episode":
		if len(args) != 3 {
			return fmt.Errorf("usage: animus-news validate-episode <episode-dir>")
		}
		if err := artifacts.ValidateEpisodeDirectory(args[2]); err != nil {
			return err
		}
		fmt.Printf("episode valid: %s\n", args[2])
		return nil
	case "extract-claims":
		if len(args) != 3 {
			return fmt.Errorf("usage: animus-news extract-claims <episode-dir>")
		}
		result, err := claimextractor.ExtractEpisode(args[2])
		if err != nil {
			return err
		}
		for _, warning := range result.Warnings {
			fmt.Fprintf(os.Stderr, "warning: %s\n", warning)
		}
		encoded, err := json.MarshalIndent(result.ClaimsFile, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(encoded))
		return nil
	case "dry-run":
		if len(args) != 3 {
			return fmt.Errorf("usage: animus-news dry-run <episode-dir>")
		}
		report, err := pipeline.DryRun(args[2])
		if err != nil {
			return err
		}
		fmt.Println(report.String())
		return nil
	case "scan-secrets":
		if len(args) != 3 {
			return fmt.Errorf("usage: animus-news scan-secrets <path>")
		}
		summary, err := security.ScanPath(args[2])
		if err != nil {
			return err
		}
		encoded, err := json.MarshalIndent(summary, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(encoded))
		if summary.HasHighRiskFindings() {
			return fmt.Errorf("high-risk secret findings detected")
		}
		return nil
	case "provider-capabilities":
		if len(args) != 2 {
			return fmt.Errorf("usage: animus-news provider-capabilities")
		}
		registry := capabilities.DefaultRegistry()
		if err := registry.Validate(); err != nil {
			return err
		}
		encoded, err := json.MarshalIndent(registry.List(), "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(encoded))
		return nil
	case "pilot":
		return runPilot(ctx, args[2:])
	case "worker":
		return worker.Run(ctx, worker.Config{})
	case "start-workflow":
		if len(args) != 4 {
			return fmt.Errorf("usage: animus-news start-workflow <episode-id> <episode-dir>")
		}
		run, err := temporalops.StartEpisode(ctx, temporalops.Config{}, args[2], args[3])
		if err != nil {
			return err
		}
		fmt.Printf("workflow started: workflow_id=%s run_id=%s\n", run.GetID(), run.GetRunID())
		return nil
	case "signal-human-qa":
		if len(args) != 4 {
			return fmt.Errorf("usage: animus-news signal-human-qa <workflow-id> <approve|approve_with_minor_edits|request_revision|block>")
		}
		return temporalops.SignalHumanQA(ctx, temporalops.Config{}, args[2], args[3])
	case "signal-release":
		if len(args) != 4 {
			return fmt.Errorf("usage: animus-news signal-release <workflow-id> <approve|block>")
		}
		return temporalops.SignalRelease(ctx, temporalops.Config{}, args[2], args[3])
	case "query-state":
		if len(args) != 3 {
			return fmt.Errorf("usage: animus-news query-state <workflow-id>")
		}
		state, err := temporalops.QueryEpisodeState(ctx, temporalops.Config{}, args[2])
		if err != nil {
			return err
		}
		encoded, err := json.MarshalIndent(state, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(encoded))
		return nil
	case "help", "-h", "--help":
		printUsage()
		return nil
	default:
		return fmt.Errorf("unknown command %q", args[1])
	}
}

func runPilot(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: animus-news pilot <generate-real|resume|status|validate|import-claude-review|import-visual-shot|import-voice>")
	}
	switch args[0] {
	case "generate-real":
		req, err := parsePilotGenerateArgs(args[1:])
		if err != nil {
			return err
		}
		res, err := pilot.GenerateReal(ctx, req)
		if err != nil {
			return err
		}
		fmt.Println(res.String())
		return nil
	case "resume":
		episodeDir, err := parseEpisodeDirFlag("pilot resume", args[1:])
		if err != nil {
			return err
		}
		res, err := pilot.Resume(ctx, episodeDir)
		if err != nil {
			return err
		}
		fmt.Println(res.String())
		return nil
	case "status":
		episodeDir, err := parseEpisodeDirFlag("pilot status", args[1:])
		if err != nil {
			return err
		}
		res, err := pilot.Status(episodeDir)
		if err != nil {
			return err
		}
		fmt.Println(res.String())
		return nil
	case "validate":
		episodeDir, err := parseEpisodeDirFlag("pilot validate", args[1:])
		if err != nil {
			return err
		}
		report, err := pilot.Validate(episodeDir)
		if err != nil {
			return err
		}
		encoded, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(encoded))
		if !report.Valid {
			return fmt.Errorf("pilot episode is not release-candidate ready")
		}
		return nil
	case "import-claude-review":
		req, err := parseImportClaudeReviewArgs(args[1:])
		if err != nil {
			return err
		}
		return pilot.ImportClaudeReview(req)
	case "import-visual-shot":
		req, err := parseImportAssetArgs("pilot import-visual-shot", args[1:], true)
		if err != nil {
			return err
		}
		return pilot.ImportVisualShot(req)
	case "import-voice":
		req, err := parseImportAssetArgs("pilot import-voice", args[1:], false)
		if err != nil {
			return err
		}
		return pilot.ImportVoice(req)
	default:
		return fmt.Errorf("unknown pilot command %q", args[0])
	}
}

func parsePilotGenerateArgs(args []string) (pilot.GenerateRequest, error) {
	fs := flag.NewFlagSet("pilot generate-real", flag.ContinueOnError)
	episodeID := fs.String("episode-id", "", "episode id")
	prompt := fs.String("prompt", "", "source prompt")
	language := fs.String("language", "", "language code")
	duration := fs.String("duration", "", "duration such as 45s")
	platforms := fs.String("platforms", "", "comma-separated platforms")
	visualProvider := fs.String("visual-provider", "", "visual provider")
	voiceProvider := fs.String("voice-provider", "", "voice provider")
	subtitleProvider := fs.String("subtitle-provider", "", "subtitle provider")
	renderProvider := fs.String("render-provider", "", "render provider")
	claudeReview := fs.String("claude-review", "", "Claude review mode: manual or api")
	out := fs.String("out", "", "episode output directory")
	if err := fs.Parse(args); err != nil {
		return pilot.GenerateRequest{}, err
	}
	return pilot.GenerateRequest{
		EpisodeID:        *episodeID,
		Prompt:           *prompt,
		Language:         *language,
		Duration:         *duration,
		Platforms:        strings.Split(*platforms, ","),
		VisualProvider:   *visualProvider,
		VoiceProvider:    *voiceProvider,
		SubtitleProvider: *subtitleProvider,
		RenderProvider:   *renderProvider,
		ClaudeReview:     *claudeReview,
		OutDir:           *out,
	}, nil
}

func parseEpisodeDirFlag(command string, args []string) (string, error) {
	fs := flag.NewFlagSet(command, flag.ContinueOnError)
	episodeDir := fs.String("episode-dir", "", "episode directory")
	if err := fs.Parse(args); err != nil {
		return "", err
	}
	if *episodeDir == "" {
		return "", fmt.Errorf("usage: animus-news %s --episode-dir <dir>", command)
	}
	return *episodeDir, nil
}

func parseImportClaudeReviewArgs(args []string) (pilot.ImportClaudeReviewRequest, error) {
	fs := flag.NewFlagSet("pilot import-claude-review", flag.ContinueOnError)
	episodeDir := fs.String("episode-dir", "", "episode directory")
	kind := fs.String("kind", "", "script or final")
	file := fs.String("file", "", "Claude JSON response file")
	if err := fs.Parse(args); err != nil {
		return pilot.ImportClaudeReviewRequest{}, err
	}
	if *episodeDir == "" || *kind == "" || *file == "" {
		return pilot.ImportClaudeReviewRequest{}, fmt.Errorf("usage: animus-news pilot import-claude-review --episode-dir <dir> --kind <script|final> --file <json>")
	}
	return pilot.ImportClaudeReviewRequest{EpisodeDir: *episodeDir, Kind: *kind, File: *file}, nil
}

func parseImportAssetArgs(command string, args []string, requireShot bool) (pilot.ImportAssetRequest, error) {
	fs := flag.NewFlagSet(command, flag.ContinueOnError)
	episodeDir := fs.String("episode-dir", "", "episode directory")
	shotID := fs.String("shot-id", "", "shot id")
	file := fs.String("file", "", "asset file")
	if err := fs.Parse(args); err != nil {
		return pilot.ImportAssetRequest{}, err
	}
	if *episodeDir == "" || *file == "" || requireShot && *shotID == "" {
		if requireShot {
			return pilot.ImportAssetRequest{}, fmt.Errorf("usage: animus-news %s --episode-dir <dir> --shot-id <shot-id> --file <path>", command)
		}
		return pilot.ImportAssetRequest{}, fmt.Errorf("usage: animus-news %s --episode-dir <dir> --file <path>", command)
	}
	return pilot.ImportAssetRequest{EpisodeDir: *episodeDir, ShotID: *shotID, File: *file}, nil
}

// runDemo drives the short-form pipeline end-to-end on mock providers and writes
// all artifacts, gate decisions, and an audit log under the run directory. With
// --expect it returns a single pass/fail signal (used by `make verify`).
func runDemo(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("demo", flag.ContinueOnError)
	episode := fs.String("episode", "episode-0001", "episode id to run")
	inject := fs.String("inject", "none", "failure injection: none|unapproved_storyboard|render_no_audio|release_denied")
	out := fs.String("out", filepath.Join("dist", "demo"), "output base directory")
	expect := fs.String("expect", "", "assertion: terminal | blocked:<gate> (empty = no assertion)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	injection := map[string]runner.Injection{
		"none":                  runner.InjectNone,
		"unapproved_storyboard": runner.InjectUnapprovedStoryboard,
		"render_no_audio":       runner.InjectRenderNoAudio,
		"release_denied":        runner.InjectReleaseDenied,
	}[*inject]

	res, err := runner.Run(ctx, runner.Config{EpisodeID: *episode, OutputDir: *out, Inject: injection})
	if err != nil {
		return err
	}

	fmt.Printf("episode:     %s\n", res.EpisodeID)
	fmt.Printf("run dir:     %s\n", res.RunDir)
	fmt.Printf("state:       %s\n", res.State)
	fmt.Printf("blocked:     %v\n", res.Blocked)
	if res.Blocked {
		fmt.Printf("block reason: %s\n", res.BlockReason)
	}
	fmt.Printf("artifacts:   %d  gates evaluated: %d\n", len(res.Artifacts), len(res.GateResults))

	if *expect == "" {
		return nil
	}
	if err := assertExpectation(res, *expect); err != nil {
		return err
	}
	fmt.Printf("expectation met: %s\n", *expect)
	return nil
}

func assertExpectation(res runner.Result, expect string) error {
	if expect == "terminal" {
		if res.Blocked || res.State != "published_dry_run_complete" {
			return fmt.Errorf("expected terminal success, got state=%s blocked=%v", res.State, res.Blocked)
		}
		return nil
	}
	if gate, ok := strings.CutPrefix(expect, "blocked:"); ok {
		if !res.Blocked {
			return fmt.Errorf("expected a block at %s, but the run completed", gate)
		}
		if len(res.GateResults) > 0 {
			last := res.GateResults[len(res.GateResults)-1]
			if last.Gate == gate || res.State == gate {
				return nil
			}
		}
		if res.State == gate {
			return nil
		}
		return fmt.Errorf("expected block at %q, got state=%s reason=%s", gate, res.State, res.BlockReason)
	}
	return fmt.Errorf("unknown --expect value: %s", expect)
}

func parseValidateArgs(args []string) (bool, string, error) {
	if len(args) == 0 || args[0] == "--json" && len(args) != 2 {
		return false, "", fmt.Errorf("usage: animus-news validate [--json] <path>")
	}
	if len(args) == 1 {
		return false, args[0], nil
	}
	if len(args) == 2 && args[0] == "--json" {
		return true, args[1], nil
	}
	return false, "", fmt.Errorf("usage: animus-news validate [--json] <path>")
}

func printUsage() {
	fmt.Println(`Animus News CLI

Usage:
  animus-news demo [--episode <id>] [--inject <mode>] [--out <dir>] [--expect <terminal|blocked:<gate>>]
  animus-news validate [--json] <path>
  animus-news validate-shortform <artifact-file>
  animus-news validate-episode <episode-dir>
  animus-news extract-claims <episode-dir>
  animus-news dry-run <episode-dir>
  animus-news scan-secrets <path>
  animus-news provider-capabilities
  animus-news pilot generate-real --episode-id <id> --prompt <text> --language <lang> --duration <seconds> --platforms <list> --visual-provider external-command --voice-provider external-command --subtitle-provider <faster-whisper|script-timing> --render-provider ffmpeg --claude-review <manual|api> --out <episode-dir>
  animus-news pilot resume --episode-dir <episode-dir>
  animus-news pilot status --episode-dir <episode-dir>
  animus-news pilot validate --episode-dir <episode-dir>
  animus-news pilot import-claude-review --episode-dir <episode-dir> --kind <script|final> --file <json>
  animus-news pilot import-visual-shot --episode-dir <episode-dir> --shot-id <shot-id> --file <mp4>
  animus-news pilot import-voice --episode-dir <episode-dir> --file <wav>
  animus-news worker
  animus-news start-workflow <episode-id> <episode-dir>
  animus-news signal-human-qa <workflow-id> <approve|approve_with_minor_edits|request_revision|block>
  animus-news signal-release <workflow-id> <approve|block>
  animus-news query-state <workflow-id>

This CLI is intentionally safe-by-default: no direct public publishing. The worker and workflow commands require a local or configured Temporal service.`)
}
