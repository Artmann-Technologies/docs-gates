package gatecli

import (
	"io"
	"os"
	"sync"
	"time"
)

// CapturedOutput is what a Run-shaped function emitted plus its exit code.
type CapturedOutput struct {
	Stdout string
	Stderr string
	Code   int
}

// CaptureOutput redirects os.Stdout and os.Stderr around fn and returns
// everything it emitted plus its exit code (the gates write via the
// process-level streams; tests capture them the way the original Python
// harness captured subprocess output).
//
// The swap is a PROCESS-GLOBAL mutation: every caller must stay out of
// t.Parallel() — mixing this helper with parallel stdout writers races
// the whole suite's output. The serialization rule lives here so every
// consuming suite inherits it from one place (TODO stdout-capture,
// 2026-10-16: one shared capture instead of one hand-rolled per suite).
func CaptureOutput(fn func() int) CapturedOutput {
	pipes := newCapturedPipes()

	code := fn()

	pipes.restore()
	pipes.closeWrites()

	stdout, stderr := pipes.wait()

	return CapturedOutput{Stdout: stdout, Stderr: stderr, Code: code}
}

// capturedPipes is one end-to-end stream capture: both pipe pairs with
// CONCURRENT drainers. Draining while fn runs is not an optimization —
// it is the 2026-11-05 fleet-wedge fix: a gate writing more than the
// pipe buffer (~64KB — crm's 179-finding attribution dump) parks its
// write on a full pipe while io.ReadAll waits for fn to return first,
// and fn cannot return while its write blocks. Netpoll-parked, 0% CPU,
// forever (the autopsied epoll-thread wedge). The drainers' EOF stays
// deterministic: the write ends close only after fn returns, so no
// partial output can race the verdict — the goroutines only READ.
type capturedPipes struct {
	restore     func()
	closeWrites func()
	wait        func() (stdout, stderr string)
}

func newCapturedPipes() capturedPipes {
	outR, outW, _ := os.Pipe()
	errR, errW, _ := os.Pipe()
	origOut, origErr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = outW, errW

	liveOut.Store(origOut)
	liveErr.Store(origErr)

	var (
		wg       sync.WaitGroup
		outBytes []byte
		errBytes []byte

		// restoreOnce makes the restore single-shot: CaptureOutputWithTimeout's
		// timeout path restores immediately, and the abandoned gate goroutine
		// calls restore AGAIN whenever the straggler finally returns — a late
		// second assignment of the process globals would race whatever
		// capture started in between (the 2026-11-14 race-detector finding).
		restoreOnce sync.Once
	)

	const pipeCount = 2

	wg.Add(pipeCount)

	go func() {
		defer wg.Done()

		outBytes, _ = io.ReadAll(outR)
	}()

	go func() {
		defer wg.Done()

		errBytes, _ = io.ReadAll(errR)
	}()

	return capturedPipes{
		restore: func() {
			restoreOnce.Do(func() {
				liveOut.Store(nil)
				liveErr.Store(nil)

				os.Stdout, os.Stderr = origOut, origErr
			})
		},
		closeWrites: func() { _ = outW.Close(); _ = errW.Close() },
		wait: func() (string, string) {
			wg.Wait()

			_ = outR.Close()
			_ = errR.Close()

			return string(outBytes), string(errBytes)
		},
	}
}

// CaptureOutputWithTimeout runs fn under the captured streams like
// CaptureOutput, but abandons it once the wall-clock budget expires:
// the second return reports the timeout. The gate functions take no
// context, so an expired gate CANNOT be killed — it keeps running as an
// abandoned goroutine whose late writes fall through to the restored
// streams (cosmetic at worst), and os.Exit does not wait for
// goroutines. A timeout <= 0 never fires — the opt-out shape.
func CaptureOutputWithTimeout(fn func() int, timeout time.Duration) (CapturedOutput, bool) {
	if timeout <= 0 {
		return CaptureOutput(fn), false
	}

	pipes := newCapturedPipes()

	done := make(chan CapturedOutput, 1)

	go func() {
		code := fn()

		pipes.restore()
		pipes.closeWrites()

		stdout, stderr := pipes.wait()

		done <- CapturedOutput{Stdout: stdout, Stderr: stderr, Code: code}
	}()

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case captured := <-done:
		return captured, false
	case <-timer.C:
		// Restore what the swap found — NOT devnull: the caller keeps
		// printing (the matrix must render). Closing the write ends EOFs
		// the drainers so nothing leaks past the straggler's lifetime.
		pipes.restore()
		pipes.closeWrites()

		return CapturedOutput{Code: ExitFindings}, true
	}
}
