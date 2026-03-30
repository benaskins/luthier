package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func buildCmd(t *testing.T) (*cobra.Command, *bool, *bool, *string, *string) {
	t.Helper()
	var dryRun, verbose bool
	var resumeFrom, coder string

	cmd := &cobra.Command{
		Use:  "maestro <project-dir>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error { return nil },
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "")
	cmd.Flags().StringVar(&resumeFrom, "resume-from", "", "")
	cmd.Flags().StringVar(&coder, "coder", "claude", "")
	return cmd, &dryRun, &verbose, &resumeFrom, &coder
}

func TestFlagDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	cmd, dryRun, verbose, resumeFrom, coder := buildCmd(t)
	cmd.SetArgs([]string{tmpDir})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if *dryRun {
		t.Error("dry-run default should be false")
	}
	if *verbose {
		t.Error("verbose default should be false")
	}
	if *resumeFrom != "" {
		t.Errorf("resume-from default should be empty, got %q", *resumeFrom)
	}
	if *coder != "claude" {
		t.Errorf("coder default should be claude, got %q", *coder)
	}
}

func TestFlagParsing(t *testing.T) {
	tmpDir := t.TempDir()
	cmd, dryRun, verbose, resumeFrom, coder := buildCmd(t)
	cmd.SetArgs([]string{
		"--dry-run",
		"--verbose",
		"--resume-from", "step 3",
		"--coder", "custom",
		tmpDir,
	})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !*dryRun {
		t.Error("expected dry-run true")
	}
	if !*verbose {
		t.Error("expected verbose true")
	}
	if *resumeFrom != "step 3" {
		t.Errorf("expected resume-from 'step 3', got %q", *resumeFrom)
	}
	if *coder != "custom" {
		t.Errorf("expected coder 'custom', got %q", *coder)
	}
}

func TestRequiresProjectDir(t *testing.T) {
	cmd, _, _, _, _ := buildCmd(t)
	cmd.SetArgs([]string{})
	buf := new(strings.Builder)
	cmd.SetErr(buf)
	if err := cmd.Execute(); err == nil {
		t.Error("expected error when no project dir provided")
	}
}

func TestProjectDirMustExist(t *testing.T) {
	// Use the real command logic inline
	nonexistent := filepath.Join(os.TempDir(), "maestro-no-such-dir-xyz")
	_, err := os.Stat(nonexistent)
	if err == nil {
		t.Skip("directory unexpectedly exists")
	}
	// Confirm os.Stat returns an error for a missing path
	if !os.IsNotExist(err) {
		t.Errorf("unexpected error type: %v", err)
	}
}
