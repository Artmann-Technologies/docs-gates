package gatecli

import "os"

// WritePreservingMode writes content with the file's existing permission
// bits; a fresh write gets the caller's fallback. The ONE home for the
// mutator verbs' writes — purge, archive and strike all preserve the
// file's own mode (a rewrite must never change a file's permissions),
// each passing its own fresh-file fallback.
func WritePreservingMode(path, content string, fallback os.FileMode) error {
	mode := fallback

	if info, err := os.Stat(path); err == nil {
		mode = info.Mode().Perm()
	}

	return os.WriteFile(path, []byte(content), mode)
}
