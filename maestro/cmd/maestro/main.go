package main

import (
	"fmt"
	"os"

	"github.com/benaskins/maestro/internal/agent"
	"github.com/benaskins/maestro/internal/orchestrate"
	"github.com/spf13/cobra"
)

func main() {
	var (
		dryRun     bool
		verbose    bool
		resumeFrom string
		coder      string
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
			if err != nil || !info.IsDir() {
				return fmt.Errorf("%s is not a directory", projectDir)
			}

			var a agent.Agent
			switch coder {
			case "claude":
				a = &agent.Claude{Verbose: verbose}
			default:
				return fmt.Errorf("unknown coder %q", coder)
			}

			fmt.Fprintf(os.Stderr, "maestro: conducting %s\n", projectDir)

			result, err := orchestrate.Run(orchestrate.Config{
				ProjectDir: projectDir,
				Agent:      a,
				DryRun:     dryRun,
				Verbose:    verbose,
				ResumeFrom: resumeFrom,
			})

			if result != nil {
				fmt.Fprintf(os.Stderr, "\nmaestro: %d/%d steps completed, %d skipped, %d failed\n",
					result.Completed, result.Total, result.Skipped, result.Failed)
			}

			return err
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show what would be done without executing")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "show coding agent output")
	cmd.Flags().StringVar(&resumeFrom, "resume-from", "", "resume from step title or number")
	cmd.Flags().StringVar(&coder, "coder", "claude", "coding agent: claude")

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
