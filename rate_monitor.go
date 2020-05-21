package logtop

import (
	"time"
)

type RateMonitor struct {
	lastSnapshotExists bool
	lastSnapshotAt     time.Time

	counts map[string]uint64
}

func (mon *RateMonitor) Record(val string) {
	if current, ok := mon.counts[val]; ok {
		mon.counts[val] = current + 1
	} else {
		mon.counts[val] = 1
	}
}

// returns a set of rates since the last snapshot
func (mon *RateMonitor) Snapshot() map[string]float64 {
	currentTime := time.Now()

	// first time around, no rates to return
	if !mon.lastSnapshotExists {
		mon.lastSnapshotExists = true
		mon.lastSnapshotAt = currentTime
		mon.counts = make(map[string]uint64)
		return make(map[string]float64)
	}

	interval := currentTime.Sub(mon.lastSnapshotAt)

	rates := make(map[string]float64)
	for key, count := range mon.counts {
		rates[key] = float64(count) / interval.Seconds()
	}

	mon.lastSnapshotAt = currentTime
	mon.counts = make(map[string]uint64)

	return rates
}

func NewRateMonitor() *RateMonitor {
	return &RateMonitor{
		counts: make(map[string]uint64),
	}
}
