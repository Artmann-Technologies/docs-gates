package docsannotate

import (
	"context"
	"errors"
	"os/exec"
	"strings"
)

// git's "command not found" status; sandboxes (nix flake check) ship
// no git, and the hash/tag rules warn-skip on it.
const rcNoGit = 127

type gitRunner struct {
	root    string
	missing bool
}

func (g gitRunner) run(args ...string) (string, int) {
	if g.missing {
		return "", rcNoGit
	}

	//nolint:gosec // git -C this repo; every subcommand is a literal from this package
	cmd := exec.CommandContext(context.Background(), "git", append([]string{"-C", g.root}, args...)...)

	out, err := cmd.Output()
	if err != nil {
		if ee, ok := errors.AsType[*exec.ExitError](err); ok {
			return string(out), ee.ExitCode()
		}
		// Sandboxes (nix flake check) ship no git; treat as "not a
		// checkout" so the hash/tag rules warn-skip instead of crashing.
		return "", rcNoGit
	}

	return string(out), 0
}

func isCheckout(root string) (gitRunner, bool) {
	g := gitRunner{root: root}

	out, rc := g.run("rev-parse", "--git-dir")
	if rc == rcNoGit {
		g.missing = true

		return g, false
	}

	return g, rc == 0 && strings.TrimSpace(out) != ""
}
