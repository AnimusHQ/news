package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/AnimusHQ/news/internal/artifacts"
	claimextractor "github.com/AnimusHQ/news/internal/claims"
	"github.com/AnimusHQ/news/internal/pipeline"
	"github.com/AnimusHQ/news/internal/security"
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
  animus-news validate [--json] <path>
  animus-news validate-episode <episode-dir>
  animus-news extract-claims <episode-dir>
  animus-news dry-run <episode-dir>
  animus-news scan-secrets <path>
  animus-news worker
  animus-news start-workflow <episode-id> <episode-dir>
  animus-news signal-human-qa <workflow-id> <approve|approve_with_minor_edits|request_revision|block>
  animus-news signal-release <workflow-id> <approve|block>
  animus-news query-state <workflow-id>

This CLI is intentionally safe-by-default: no direct public publishing. The worker and workflow commands require a local or configured Temporal service.`)
}
