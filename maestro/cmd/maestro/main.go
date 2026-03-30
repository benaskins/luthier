package main

import (
	"fmt"
	"os"

	"github.com/benaskins/maestro/internal/agent"
	"github.com/benaskins/maestro/internal/logger"
	"github.com/benaskins/maestro/internal/orchestrate"
	"github.com/benaskins/maestro/internal/report"
	"github.com/spf13/cobra"
)

func main() {
	var (
		dryRun      bool
		verbose     bool
		resumeFrom  string
		coder       string
		summaryFile string
	)

	cmd := &cobra.Command{
		Use:   "maestro <project-dir>",
		Short: "Orchestrate coding agent execution of a luthier scaffold plan",
		Long: `maestro reads a plan from the given project directory and delegates
each step to a coding agent, verifying and committing as it goes.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectDir := args[0]

			info, err := os.Stat(projectDir)
			if err != nil {
				return fmt.Errorf("cannot access project directory %q: %w", projectDir, err)
			}
			if !info.IsDir() {
				return fmt.Errorf("%q is not a directory", projectDir)
			}

			a, err := agent.New(coder, verbose)
			if err != nil {
				return fmt.Errorf("coder: %w", err)
			}

			level := logger.LevelInfo
			if verbose {
				level = logger.LevelDebug
			}
			log := logger.New(level, os.Stderr)

			log.Info("maestro: conducting %s\n", projectDir)

			result, runErr := orchestrate.Run(orchestrate.Config{
				ProjectDir: projectDir,
				Agent:      a,
				DryRun:     dryRun,
				Verbose:    verbose,
				ResumeFrom: resumeFrom,
				Logger:     log,
			})

			if result != nil {
				report.Print(os.Stderr, result)
				if summaryFile != "" {
					if writeErr := report.WriteFile(summaryFile, result); writeErr != nil {
						log.Warn("could not write summary file: %v\n", writeErr)
					}
				}
			}

			return runErr
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show what would be done without executing")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "show coding agent output and debug messages")
	cmd.Flags().StringVar(&resumeFrom, "resume-from", "", "resume from step title or number")
	cmd.Flags().StringVar(&coder, "coder", "claude", "coding agent: claude, noop")
	cmd.Flags().StringVar(&summaryFile, "summary-file", "", "write final status summary to this file")

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
