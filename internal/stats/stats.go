package stats

import (
	"fmt"
	"time"
)

type NetworkStats struct {
	Target      string
	TotalChecks int
	UpCount     int
	DownCount   int
	LastCheck   time.Time
}

func (n *NetworkStats) RecordCheck(isUp bool) {
	n.TotalChecks++
	if isUp {
		n.UpCount++
	} else {
		n.DownCount++
	}
	n.LastCheck = time.Now()
}

func (n NetworkStats) SuccessRate() float64 {
	if n.TotalChecks == 0 {
		return 0
	}

	return float64(n.UpCount) / float64(n.TotalChecks) * 100
}

func (n NetworkStats) LastCheckAgo() string {
	if n.LastCheck.IsZero() {
		return "never"
	}

	return fmt.Sprintf("%s", time.Since(n.LastCheck).Round(time.Second))
}

func (n NetworkStats) Report() string {
	return fmt.Sprintf("[NET] \t%s | Success Rate: %.2f%% | Last Check: %s",
		n.Target, n.SuccessRate(), n.LastCheckAgo())
}
