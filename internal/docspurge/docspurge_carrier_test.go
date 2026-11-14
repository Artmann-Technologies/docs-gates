package docspurge

import (
	"errors"
	"strings"
	"testing"
)

// TestMissingAnchorErrorRoundTrip pins the carrier contract:
// errors.AsType reads the refusing scope, and the message carries the
// changelog, the anchor phrase, and the nearest same-level heading.
func TestMissingAnchorErrorRoundTrip(t *testing.T) {
	t.Parallel()

	err := newMissingAnchorError(anchorChangedSection, "CHANGELOG.md", 7)

	mae, ok := errors.AsType[*missingAnchorError](err)
	if !ok {
		t.Fatalf("errors.AsType[*missingAnchorError] failed on: %v", err)
	}

	if mae.scope != anchorChangedSection {
		t.Errorf("scope = %v, want anchorChangedSection", mae.scope)
	}

	if mae.nearestLine != 7 {
		t.Errorf("nearestLine = %d, want 7", mae.nearestLine)
	}

	if mae.changelog != "CHANGELOG.md" {
		t.Errorf("changelog = %q, want CHANGELOG.md", mae.changelog)
	}
}

// TestMissingAnchorErrorMessageBytes pins the refusal spelling — the
// FATAL line verbatim, with and without a nearest-heading hint.
func TestMissingAnchorErrorMessageBytes(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		err  *missingAnchorError
		want string
	}{
		"unreleased with hint": {
			err:  newMissingAnchorError(anchorUnreleasedHeading, "CHANGELOG.md", 3),
			want: "CHANGELOG.md has no ## [Unreleased] heading — refusing to guess an insertion point (nearest ## heading at line 3)",
		},
		"changed with hint": {
			err:  newMissingAnchorError(anchorChangedSection, "CHANGELOG.md", 5),
			want: "CHANGELOG.md has no ### Changed section under [Unreleased] — refusing to guess an insertion point (nearest ### heading at line 5)",
		},
		"no heading in the file, no hint": {
			err:  newMissingAnchorError(anchorUnreleasedHeading, "CHANGELOG.md", 0),
			want: "CHANGELOG.md has no ## [Unreleased] heading — refusing to guess an insertion point",
		},
		"loud default scope": {
			err:  newMissingAnchorError(anchorScope(99), "CHANGELOG.md", 0),
			want: "CHANGELOG.md has no anchorScope(99) — refusing to guess an insertion point",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := tc.err.Error(); got != tc.want {
				t.Fatalf("Error() = %q, want %q", got, tc.want)
			}
		})
	}
}

// TestInsertBulletsRefusalNamesNearestHeading drives the real scan: a
// missing anchor refuses and names the nearest heading of the anchor's
// own level.
func TestInsertBulletsRefusalNamesNearestHeading(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		changelog     string
		wantInMessage string
	}{
		"no [Unreleased], h2 hint": {
			changelog:     "# Changelog\n\n## v0.1.0\n\n### Fixed\n\n- x\n",
			wantInMessage: "(nearest ## heading at line 3)",
		},
		"no Changed under [Unreleased], h3 hint": {
			changelog:     "# Changelog\n\n## [Unreleased]\n\n### Fixed\n\n- x\n",
			wantInMessage: "(nearest ### heading at line 5)",
		},
		"no heading at all, no hint": {
			changelog:     "# Changelog\n\nsome prose\n",
			wantInMessage: "refusing to guess an insertion point\n",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := insertBullets(tc.changelog, []string{"- bullet"})
			if err == nil {
				t.Fatal("insertBullets succeeded, want a missing-anchor refusal")
			}

			if !strings.Contains(err.Error(), tc.wantInMessage) {
				t.Fatalf("refusal %q must contain %q", err.Error(), tc.wantInMessage)
			}
		})
	}
}
