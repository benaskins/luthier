// Package report formats and writes the final orchestration status summary.
package report

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/benaskins/maestro/internal/orchestrate"
)

// Print writes a human-readable summary of the orchestration result to w.
func Print(w io.Writer, result *orchestrate.Result) {
	if result == nil {
		return
	}

	fmt.Fprintln(w)

	// Per-step table
	titleWidth := 0
	for _, s := range result.Steps {
		if len(s.Title) > titleWidth {
			titleWidth = len(s.Title)
		}
	}
	if titleWidth < 12 {
		titleWidth = 12
	}

	for _, s := range result.Steps {
		icon := stepIcon(s.Status)
		attempts := ""
		if s.Status != orchestrate.StatusSkipped {
			if s.Attempts == 1 {
				attempts = "1 attempt "
			} else {
				attempts = fmt.Sprintf("%d attempts", s.Attempts)
			}
		}
		dur := ""
		if s.Duration > 0 {
			dur = formatDuration(s.Duration)
		}
		fmt.Fprintf(w, "  %s [%d] %-*s  %-9s  %-10s  %s\n",
			icon, s.Number,
			titleWidth, s.Title,
			string(s.Status),
			attempts,
			dur,
		)
	}

	// Summary block
	fmt.Fprintln(w)
	fmt.Fprintf(w, "  Total:     %d step%s\n", result.Total, plural(result.Total))
	fmt.Fprintf(w, "  Completed: %d step%s\n", result.Completed, plural(result.Completed))
	if result.Skipped > 0 {
		fmt.Fprintf(w, "  Skipped:   %d step%s\n", result.Skipped, plural(result.Skipped))
	}
	if result.Failed > 0 {
		fmt.Fprintf(w, "  Failed:    %d step%s", result.Failed, plural(result.Failed))
		if result.FailedAt != nil {
			fmt.Fprintf(w, " (step %d: %s)", result.FailedAt.Number, result.FailedAt.Title)
		}
		fmt.Fprintln(w)
	}

	// Retry statistics
	retried, totalRetries := retryStats(result.Steps)
	executed := result.Completed + result.Failed
	if executed > 0 && retried > 0 {
		fmt.Fprintln(w)
		fmt.Fprintf(w, "  Retries:   %d step%s needed retries (%d extra attempt%s)\n",
			retried, plural(retried), totalRetries, plural(totalRetries))
	}

	// Total time
	if result.Duration > 0 {
		fmt.Fprintf(w, "  Time:      %s\n", formatDuration(result.Duration))
	}
	fmt.Fprintln(w)
}

// WriteFile writes the summary to path, creating or truncating the file.
func WriteFile(path string, result *orchestrate.Result) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create summary file: %w", err)
	}
	defer f.Close()
	Print(f, result)
	return nil
}

func stepIcon(status orchestrate.StepStatus) string {
	switch status {
	case orchestrate.StatusCompleted:
		return "+"
	case orchestrate.StatusFailed:
		return "!"
	case orchestrate.StatusSkipped:
		return "-"
	default:
		return "?"
	}
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

// retryStats returns how many steps needed retries and the total extra attempts.
func retryStats(steps []orchestrate.StepResult) (retried, totalExtra int) {
	for _, s := range steps {
		if s.Status == orchestrate.StatusSkipped {
			continue
		}
		if s.Attempts > 1 {
			retried++
			totalExtra += s.Attempts - 1
		}
	}
	return
}

// formatDuration renders a duration in a compact, human-readable form.
func formatDuration(d time.Duration) string {
	d = d.Round(time.Millisecond)
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	mins := int(d.Minutes())
	secs := int(d.Seconds()) % 60
	if secs == 0 {
		return fmt.Sprintf("%dm", mins)
	}
	return fmt.Sprintf("%dm%ds", mins, secs)
}
