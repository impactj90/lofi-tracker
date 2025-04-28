package tracker

import (
	"fmt"
	"time"
)

func FormatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60

	return fmt.Sprintf("%dh %dm", hours, minutes)
}
