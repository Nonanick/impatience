package pathresolver

import (
	"errors"
	"path/filepath"
	"strings"
)

// WithIndex Tries to resolve the file by adding index.html
func WithIndex(
	path string,
	public string,
	knownFiles []string,
) (string, error) {

	if !strings.HasPrefix(path, public) {
		path = filepath.Join(public, path)
	}

	path = filepath.Join(path, "index.html")

	if HasKnownPath(path, knownFiles) {
		return path, nil
	}

	return "", errors.New("Could not locate path with index.html inside known files")
}
