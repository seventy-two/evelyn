package shitpost

import (
	"fmt"
	"time"
)

func brexitCountdown() string {
	timeBrexit, _ := time.Parse(time.RFC3339, "2020-01-31T23:00:00Z") // it literally never errors
	d := time.Since(timeBrexit)
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	lines := []string{
		fmt.Sprintf("%d days, %d hours, %d minutes and %d seconds since the GBP was made essentially worthless", days, hours, minutes, seconds),
	}
	return lines[0]
}
