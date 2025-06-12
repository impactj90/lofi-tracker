package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/impactj90/lofi-tracker/cmd/internal/db"
	"github.com/impactj90/lofi-tracker/cmd/internal/tracker"
	"github.com/spf13/cobra"
)

var (
	summaryPeriod string
	summaryBranch string
	summaryDays   int
)

func init() {
	summaryCmd.Flags().StringVarP(&summaryPeriod, "period", "p", "today", "Period to summarize (today, yesterday, week, month, or YYYY-MM-DD)")
	summaryCmd.Flags().StringVarP(&summaryBranch, "branch", "b", "", "Show summary for specific branch")
	summaryCmd.Flags().IntVarP(&summaryDays, "days", "d", 7, "Number of days to include (used with --branch)")

	rootCmd.AddCommand(summaryCmd)
}

var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Show work time summary",
	Long: `Show work time summary for different periods.

Examples:
  lofi-tracker summary                    # Today's summary
  lofi-tracker summary -p yesterday       # Yesterday's summary  
  lofi-tracker summary -p week            # This week's summary
  lofi-tracker summary -p month           # This month's summary
  lofi-tracker summary -p 2024-01-15      # Specific date
  lofi-tracker summary -b feature/ABC-123 # Branch summary (last 7 days)
  lofi-tracker summary -b main -d 30      # Branch summary (last 30 days)`,
	Run: func(cmd *cobra.Command, args []string) {
		tr, currentBranch, err := tracker.Init()
		if err != nil {
			fmt.Printf("❌ Failed to initialize tracker: %v\n", err)
			return
		}
		defer tr.Close()

		// Handle branch-specific summary
		if summaryBranch != "" {
			showBranchSummary(tr, summaryBranch, summaryDays)
			return
		}

		// Handle period-based summary
		showPeriodSummary(tr, summaryPeriod, currentBranch)
	},
}

func showBranchSummary(tr tracker.Tracker, branch string, days int) {
	summary, err := tr.GetBranchSummary(branch, days)
	if err != nil {
		fmt.Printf("❌ Failed to get branch summary: %v\n", err)
		return
	}

	fmt.Printf("📊 Summary for branch '%s' (last %d days)\n\n", branch, days)

	if summary.SessionCount == 0 {
		fmt.Println("No work sessions found for this branch in the specified period.")
		return
	}

	fmt.Printf("🕒 Total Time:    %s\n", tracker.FormatSummaryDuration(summary.TotalTime))
	fmt.Printf("▶️  Active Time:   %s\n", tracker.FormatSummaryDuration(summary.ActiveTime))
	fmt.Printf("⏸️  Pause Time:    %s\n", tracker.FormatSummaryDuration(summary.PauseTime))
	fmt.Printf("📈 Efficiency:    %s\n", tracker.FormatEfficiency(summary.ActiveTime, summary.TotalTime))
	fmt.Printf("🔢 Sessions:      %d\n", summary.SessionCount)

	if !summary.StartDate.IsZero() {
		fmt.Printf("📅 Period:        %s to %s\n",
			summary.StartDate.Format("2006-01-02"),
			summary.EndDate.Format("2006-01-02"))
	}
}

func showPeriodSummary(tr tracker.Tracker, period, currentBranch string) {
	var summaries []db.SummaryData
	var err error
	var periodTitle string

	switch strings.ToLower(period) {
	case "today":
		summaries, err = tr.GetDailySummary(time.Now())
		periodTitle = "Today"

	case "yesterday":
		yesterday := time.Now().AddDate(0, 0, -1)
		summaries, err = tr.GetDailySummary(yesterday)
		periodTitle = "Yesterday"

	case "week":
		// Get start of current week (Monday)
		now := time.Now()
		weekday := int(now.Weekday())
		if weekday == 0 { // Sunday
			weekday = 7
		}
		startOfWeek := now.AddDate(0, 0, -(weekday - 1))
		startOfWeek = time.Date(startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day(), 0, 0, 0, 0, time.Local)

		summaries, err = tr.GetWeeklySummary(startOfWeek)
		periodTitle = "This Week"

	case "month":
		now := time.Now()
		summaries, err = tr.GetMonthlySummary(now.Year(), now.Month())
		periodTitle = "This Month"

	default:
		// Try to parse as date (YYYY-MM-DD)
		date, parseErr := time.Parse("2006-01-02", period)
		if parseErr != nil {
			fmt.Printf("❌ Invalid period '%s'. Use 'today', 'yesterday', 'week', 'month', or YYYY-MM-DD format\n", period)
			return
		}
		summaries, err = tr.GetDailySummary(date)
		periodTitle = date.Format("January 2, 2006")
	}

	if err != nil {
		fmt.Printf("❌ Failed to get summary: %v\n", err)
		return
	}

	fmt.Printf("📊 Work Summary - %s\n\n", periodTitle)

	if len(summaries) == 0 {
		fmt.Println("No work sessions found for this period.")
		return
	}

	// Create table output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Branch\tTotal\tActive\tPauses\tEfficiency\tSessions")
	fmt.Fprintln(w, "------\t-----\t------\t------\t----------\t--------")

	var totalTime, totalActive, totalPauses time.Duration
	var totalSessions int

	for _, summary := range summaries {
		efficiency := tracker.FormatEfficiency(summary.ActiveTime, summary.TotalTime)

		// Highlight current branch
		branchDisplay := summary.Branch
		if summary.Branch == currentBranch {
			branchDisplay = summary.Branch + " *"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%d\n",
			branchDisplay,
			tracker.FormatSummaryDuration(summary.TotalTime),
			tracker.FormatSummaryDuration(summary.ActiveTime),
			tracker.FormatSummaryDuration(summary.PauseTime),
			efficiency,
			summary.SessionCount,
		)

		totalTime += summary.TotalTime
		totalActive += summary.ActiveTime
		totalPauses += summary.PauseTime
		totalSessions += summary.SessionCount
	}

	// Add totals row
	fmt.Fprintln(w, "------\t-----\t------\t------\t----------\t--------")
	fmt.Fprintf(w, "TOTAL\t%s\t%s\t%s\t%s\t%d\n",
		tracker.FormatSummaryDuration(totalTime),
		tracker.FormatSummaryDuration(totalActive),
		tracker.FormatSummaryDuration(totalPauses),
		tracker.FormatEfficiency(totalActive, totalTime),
		totalSessions,
	)

	w.Flush()

	if len(summaries) > 1 {
		fmt.Printf("\n* = current branch\n")
	}
}
