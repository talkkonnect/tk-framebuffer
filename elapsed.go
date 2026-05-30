package main

import (
	"fmt"
	"time"
)

// elapsedTracker drives the "Elapsed :" display for active TX/RX sessions.
// It uses only timestamps and edge flags — no time.Timer, time.Ticker, or goroutines.
type elapsedTracker struct {
	txStart time.Time
	rxStart time.Time
	wasTX   bool
	wasRX   bool
}

func (t *elapsedTracker) update(now time.Time, transmitting, receiving bool) string {
	if transmitting && !t.wasTX {
		t.txStart = now
	}
	if receiving && !t.wasRX {
		t.rxStart = now
	}
	if t.wasRX && !receiving {
		t.rxStart = time.Time{}
	}
	if t.wasTX && !transmitting {
		t.txStart = time.Time{}
	}

	t.wasTX = transmitting
	t.wasRX = receiving

	switch {
	case transmitting && !t.txStart.IsZero():
		return formatElapsed(now.Sub(t.txStart))
	case receiving && !t.rxStart.IsZero():
		return formatElapsed(now.Sub(t.rxStart))
	default:
		return "00s"
	}
}

func formatElapsed(d time.Duration) string {
	sec := int(d.Seconds())
	if sec < 0 {
		sec = 0
	}
	return fmt.Sprintf("%02ds", sec)
}
