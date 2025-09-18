package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/impactj90/lofi-tracker/cmd/internal/tracker"
	"github.com/spf13/cobra"
)

var (
	exportFrom   string
	exportTo     string
	exportOutput string
	initTracker  = tracker.Init
)

func init() {
	exportCmd.Flags().StringVar(&exportFrom, "from", "", "Start date (YYYY-MM-DD). Defaults to first day of current month")
	exportCmd.Flags().StringVar(&exportTo, "to", "", "End date (YYYY-MM-DD). Defaults to today")
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "File path for CSV output (stdout when omitted)")
	rootCmd.AddCommand(exportCmd)
}

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export tracked time to CSV",
	RunE: func(cmd *cobra.Command, args []string) error {
		now := time.Now().UTC()
		fromDate, toDate, err := resolveDateRange(now, exportFrom, exportTo)
		if err != nil {
			return err
		}

		tr, _, err := initTracker()
		if err != nil {
			return fmt.Errorf("failed to initialize tracker: %w", err)
		}
		defer tr.Close()

		days, err := tr.DailySummaries(fromDate, toDate)
		if err != nil {
			return fmt.Errorf("failed to collect summaries: %w", err)
		}

		writer, closeFn, err := resolveWriter(exportOutput)
		if err != nil {
			return err
		}
		if closeFn != nil {
			defer closeFn()
		}

		if err := writeCSV(days, writer); err != nil {
			return err
		}

		if exportOutput != "" {
			fmt.Printf("✅ Exported %d rows to %s\n", len(days), exportOutput)
		}

		return nil
	},
}

func resolveDateRange(now time.Time, fromStr, toStr string) (time.Time, time.Time, error) {
	const layout = "2006-01-02"

	from := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	if fromStr != "" {
		parsed, err := time.Parse(layout, fromStr)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid --from date %q: %w", fromStr, err)
		}
		from = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.UTC)
	}

	to := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	if toStr != "" {
		parsed, err := time.Parse(layout, toStr)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid --to date %q: %w", toStr, err)
		}
		to = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.UTC)
	}

	return from, to, nil
}

func resolveWriter(path string) (io.Writer, func() error, error) {
	if path == "" {
		return os.Stdout, nil, nil
	}

	file, err := os.Create(path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create output file: %w", err)
	}

	return file, file.Close, nil
}

func writeCSV(rows []tracker.DaySummary, w io.Writer) error {
	csvWriter := csv.NewWriter(w)
	if err := csvWriter.Write([]string{"date", "ticket", "minutes"}); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	for _, row := range rows {
		minutes := roundedMinutes(row.Active)
		record := []string{
			row.Date.Format("2006-01-02"),
			row.Branch,
			strconv.Itoa(minutes),
		}
		if err := csvWriter.Write(record); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}

	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return fmt.Errorf("failed to flush csv writer: %w", err)
	}

	return nil
}

func roundedMinutes(d time.Duration) int {
	if d <= 0 {
		return 0
	}

	return int((d + time.Minute/2) / time.Minute)
}
