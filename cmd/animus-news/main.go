package main

import (
	"fmt"
	"os"

	"github.com/AnimusHQ/news/internal/artifacts"
	"github.com/AnimusHQ/news/internal/pipeline"
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

	switch args[1] {
	case "validate-episode":
		if len(args) != 3 {
			return fmt.Errorf("usage: animus-news validate-episode <episode-dir>")
		}
		if err := artifacts.ValidateEpisodeDirectory(args[2]); err != nil {
			return err
		}
		fmt.Printf("episode valid: %s\n", args[2])
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
	case "help", "-h", "--help":
		printUsage()
		return nil
	default:
		return fmt.Errorf("unknown command %q", args[1])
	}
}

func printUsage() {
	fmt.Println(`Animus News CLI

Usage:
  animus-news validate-episode <episode-dir>
  animus-news dry-run <episode-dir>

This CLI is intentionally safe-by-default: no model provider calls, no uploads, and no public publishing.`)
}
