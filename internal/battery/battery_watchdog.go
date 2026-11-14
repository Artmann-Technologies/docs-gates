package battery

import (
	"time"

	"github.com/Artmann-Technologies/docs-gates/internal/gatecli"
)

// halfBudgetDivisor splits the wall-clock budget for the early warning:
// the watcher fires once the gate burns half of its cap.
const halfBudgetDivisor = 2

// startHalfBudgetWatchdog arms the ONE live mid-gate line: once a gate
// burns past half its wall-clock budget, the watcher names the cap
// LIVE — on the REAL stdout via gatecli.Livef (human face), because
// during a capture the swapped streams would park the warning inside
// the gate's own buffer, killing exactly the early-warning value it
// exists for. The JSON face hands in gatecli.LiveErrf: its stdout is
// contractually the machine document alone, and the real-stream
// snapshot is what keeps the goroutine's write from racing the swap
// (the 2026-11-14 race-detector finding). The CV wedge sat
// invisible in band for hours; a watcher names the budget it burns.
// timeout <= 0 (the config opt-out) arms nothing; the returned func is
// always safe to call.
func startHalfBudgetWatchdog(gateName string, timeout time.Duration, warnf func(string, ...any)) func() {
	if timeout <= 0 {
		return func() {}
	}

	stop := make(chan struct{})

	go func() {
		timer := time.NewTimer(timeout / halfBudgetDivisor)
		defer timer.Stop()

		select {
		case <-timer.C:
			warnf(
				"docs-gates: gate %s over half its %s budget — abandoned at the cap\n",
				gateName,
				gatecli.FormatDuration(timeout),
			)
		case <-stop:
		}
	}()

	return func() { close(stop) }
}
