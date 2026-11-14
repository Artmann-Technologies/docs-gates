package gatecli

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

// TestCaptureOutputSurvivesLargeOutput is the 2026-11-05 fleet-wedge
// regression: a gate writing past the ~64KB pipe buffer into its own
// capture used to deadlock (the write parked on the full pipe while
// io.ReadAll waited for fn to return first — netpoll-parked at 0% CPU
// forever; crm's 179-finding attribution dump was the live trigger).
// With concurrent drainers the capture must complete and lose nothing.
//
//nolint:paralleltest // process-global capture swap, serialized on purpose
func TestCaptureOutputSurvivesLargeOutput(t *testing.T) {
	const lines = 40000 // ~4MB per stream, far past the pipe buffer

	start := time.Now()

	out := CaptureOutput(func() int {
		for range lines {
			fmt.Fprintln(os.Stdout, strings.Repeat("x", 99))
			fmt.Fprintln(os.Stderr, strings.Repeat("y", 99))
		}

		return ExitClean
	})

	if elapsed := time.Since(start); elapsed > 30*time.Second {
		t.Fatalf("capture deadlocked on large output: %v", elapsed)
	}

	if out.Code != ExitClean {
		t.Fatalf("code = %d, want clean", out.Code)
	}

	if want := lines * 100; len(out.Stdout) != want || len(out.Stderr) != want {
		t.Fatalf("capture lost bytes: stdout %d/%d, stderr %d/%d", len(out.Stdout), want, len(out.Stderr), want)
	}
}

func TestFormatDurationHonestPrecisionTiers(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   time.Duration
		want string
	}{
		{"sub-microsecond rounds to a floor, not a lie", 400, "<1µs"},
		{"exact microsecond tier renders µs", 800 * time.Microsecond, "800µs"},
		{"sub-millisecond keeps one decimal", 1500 * time.Microsecond, "1.5ms"},
		{"sub-10ms rounds to 100µs steps", 3140 * time.Microsecond, "3.1ms"},
		{"the old <10ms bucket is gone", 9 * time.Millisecond, "9ms"},
		{"sub-second keeps 10ms steps", 123 * time.Millisecond, "120ms"},
		{"sub-minute keeps 100ms steps", 4*time.Second + 567*time.Millisecond, "4.6s"},
		{"minutes round whole seconds", 65 * time.Second, "1m5s"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := FormatDuration(tc.in); got != tc.want {
				t.Fatalf("FormatDuration(%v) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
