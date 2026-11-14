package gatecli

import (
	"fmt"
	"os"
	"sync/atomic"
	"unicode/utf8"
)

// Truncate returns s unchanged when it fits within n bytes, else its first
// n bytes (finding snippets are display-only). The cut never splits a
// UTF-8 sequence: when byte n would land mid-rune, the cut backs off to
// the rune boundary, so the truncated snippet stays valid UTF-8.
func Truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}

	cut := n
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}

	return s[:cut]
}

// Outf prints a formatted verdict line to stdout; a stdout write failure (a
// closed pipe, a full disk) is deliberately not a gate finding — the exit
// code carries the verdict. One home for that policy so the gates cannot
// drift into unchecked fmt.Fprint* calls (errcheck) or forbidden fmt.Print*
// (forbidigo). Outln, Errf and Errln are the newline and stderr variants.
func Outf(format string, args ...any) {
	_, _ = fmt.Fprintf(
		os.Stdout,
		format,
		args...) //nolint:erraudit // policy above: a stdout write failure is not a gate finding
}

func Outln(args ...any) {
	_, _ = fmt.Fprintln(os.Stdout, args...) //nolint:erraudit // policy above
}

func Errf(format string, args ...any) {
	_, _ = fmt.Fprintf(
		os.Stderr,
		format,
		args...) //nolint:erraudit // policy above: a stderr write failure is not a gate finding
}

func Errln(args ...any) {
	_, _ = fmt.Fprintln(os.Stderr, args...) //nolint:erraudit // policy above
}

// liveOut is the process's REAL stdout while a capture swap is active
// (nil otherwise). The battery's mid-gate heartbeat must reach the
// terminal, never the gate's captured buffer — a warning parked in the
// capture only surfaces after the gate returns, which is exactly when
// it stops being an early warning.
//
//nolint:gochecknoglobals // process-global by nature: the stream swap itself is process-global
var liveOut atomic.Pointer[os.File]

// liveErr is the process's REAL stderr while a capture swap is active
// (nil otherwise): the stderr twin of liveOut, for the JSON face's
// mid-gate warnings. The watchdog goroutine fires while the swap is
// live, and reading the swapped global from another goroutine races the
// swap (the 2026-11-14 race-detector finding): snapshot, never the
// global.
//
//nolint:gochecknoglobals // process-global by nature: the stream swap itself is process-global
var liveErr atomic.Pointer[os.File]

// Livef writes a formatted progress line to the REAL stdout even while
// CaptureOutput/CaptureOutputWithTimeout have the process streams
// swapped: the half-budget early warning fires while a wedged gate
// holds the capture, and must be visible live. Outside a capture it is
// plain Outf.
func Livef(format string, args ...any) {
	if f := liveOut.Load(); f != nil {
		_, _ = fmt.Fprintf(
			f,
			format,
			args...) //nolint:erraudit // policy above: a progress write failure is not a finding

		return
	}

	Outf(format, args...)
}

// LiveErrf is Livef's stderr twin: a formatted line to the REAL stderr
// even while the capture swap is active. Outside a capture it is plain
// Errf.
func LiveErrf(format string, args ...any) {
	if f := liveErr.Load(); f != nil {
		_, _ = fmt.Fprintf(
			f,
			format,
			args...) //nolint:erraudit // policy above

		return
	}

	Errf(format, args...)
}
