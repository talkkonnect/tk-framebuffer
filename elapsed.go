package main

import (
	"fmt"
	"time"
)

// elapsedTracker drives Elapsed (live TX/RX duration) and ActivityEndTime (wall clock at last end).
// Uses only timestamps and edge flags — no time.Timer, time.Ticker, or goroutines.
type elapsedTracker struct {
	txStart   time.Time
	rxStart   time.Time
	txEndedAt time.Time
	rxEndedAt time.Time
	wasTX     bool
	wasRX     bool
}

func (t *elapsedTracker) update(now time.Time, transmitting, receiving bool) (elapsed, activityEndTime string) {
	if transmitting && !t.wasTX {
		t.txStart = now
	}
	if receiving && !t.wasRX {
		t.rxStart = now
	}
	if t.wasRX && !receiving {
		t.rxEndedAt = now
		t.rxStart = time.Time{}
	}
	if t.wasTX && !transmitting {
		t.txEndedAt = now
		t.txStart = time.Time{}
	}

	t.wasTX = transmitting
	t.wasRX = receiving

	switch {
	case transmitting && !t.txStart.IsZero():
		elapsed = formatElapsed(now.Sub(t.txStart))
	case receiving && !t.rxStart.IsZero():
		elapsed = formatElapsed(now.Sub(t.rxStart))
	default:
		elapsed = "00s"
	}

	activityEndTime = formatActivityEndTime(t.txEndedAt, t.rxEndedAt)
	return elapsed, activityEndTime
}

func formatActivityEndTime(txEnded, rxEnded time.Time) string {
	endAt := latestTime(txEnded, rxEnded)
	if endAt.IsZero() {
		return "—"
	}
	return endAt.Format("15:04:05")
}

func latestTime(a, b time.Time) time.Time {
	switch {
	case a.IsZero():
		return b
	case b.IsZero():
		return a
	case a.After(b):
		return a
	default:
		return b
	}
}

func formatElapsed(d time.Duration) string {
	sec := int(d.Seconds())
	if sec < 0 {
		sec = 0
	}
	return fmt.Sprintf("%02ds", sec)
}
