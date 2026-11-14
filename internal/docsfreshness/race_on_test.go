//go:build race

package docsfreshness

// raceDetector reports the race-detector build; timing budgets scale up
// under instrumentation (see TestReplaySpeedGuard).
const raceDetector = true
