package docsfreshness

import (
	"regexp"
	"slices"
	"strings"
)

//nolint:gochecknoglobals // static lookup table; rebuilding it per token would be the real waste
var stopwords = map[string]bool{
	"the": true, "and": true, "for": true, "with": true, "from": true,
	"that": true, "this": true, "then": true, "when": true, "into": true,
	"after": true, "before": true, "once": true, "over": true, "under": true,
	"must": true, "they": true, "them": true, "their": true,
	"should": true, "would": true, "could": true, "also": true,
	"more": true, "most": true, "than": true, "very": true,
	"items": true, "item": true, "task": true, "tasks": true,
	"open": true, "still": true, "pending": true, "until": true,
	"about": true, "again": true, "each": true, "every": true,
	"some": true, "what": true, "which": true, "where": true,
	"while": true, "yours": true, "ours": true, "have": true,
	"been": true, "being": true, "does": true, "doing": true,
	"done": true, "todo": true, "list": true, "roadmap": true,
	"report": true, "reports": true, "session": true, "sessions": true,
	"next": true, "things": true, "work": true, "owner": true,
	"user": true, "owners": true, "users": true, "here": true,
	"there": true, "only": true, "just": true, "like": true,
	"make": true, "made": true, "take": true, "taken": true,
	"keep": true, "kept": true, "both": true, "same": true,
	"other": true, "another": true, "such": true, "much": true,
	"many": true, "less": true, "least": true, "first": true,
	"last": true, "second": true, "third": true, "half": true,
	"part": true, "full": true, "wide": true, "long": true,
	"short": true, "small": true, "large": true, "real": true,
	"really": true,
}

const MinTokenLen = 4

var (
	tokenRe  = regexp.MustCompile(`[a-z0-9][a-z0-9._/-]*`)
	hyphenRe = regexp.MustCompile(`[-\s]+`)
)

// Tokens: lowercased raw tokens minus stopwords, trailing punctuation
// stripped, minimum length enforced. A filename's INTERNAL dot survives
// ("backup-restore.md" stays one token).
func Tokens(text string) []string {
	raw := tokenRe.FindAllString(strings.ToLower(text), -1)

	var out []string

	for _, t := range raw {
		t = strings.TrimRight(t, ".,;:!?)](\"'")
		if len(t) >= MinTokenLen && !stopwords[t] {
			out = append(out, t)
		}
	}

	return out
}

// Squash is the hyphen/space-collapsed, lowercased comparison key (TODO 8.4).
func Squash(text string) string {
	return strings.ToLower(hyphenRe.ReplaceAllString(text, ""))
}

// tokenPresent: raw substring match first, then the hyphen/space-collapsed
// key (TODO 8.4) when the token is long enough to carry it.
func tokenPresent(token string, livingLines, livingSquashed []string) bool {
	for _, line := range livingLines {
		if strings.Contains(line, token) {
			return true
		}
	}

	key := Squash(token)
	if len(key) >= MinTokenLen {
		for _, line := range livingSquashed {
			if strings.Contains(line, key) {
				return true
			}
		}
	}

	return false
}

func missingTokens(toks, found []string) []string {
	var missing []string

	for _, t := range toks {
		if !slices.Contains(found, t) {
			missing = append(missing, t)
		}
	}

	return missing
}

func hasDistinctive(found []string, tokenDF func(string) int, genericDF int) bool {
	for _, t := range found {
		if tokenDF(t) < genericDF {
			return true
		}
	}

	return false
}

// squashLines is the batch form of Squash used for the living-docs corpus.
func squashLines(lines []string) []string {
	squashed := make([]string, 0, len(lines))

	for _, l := range lines {
		squashed = append(squashed, Squash(l))
	}

	return squashed
}
