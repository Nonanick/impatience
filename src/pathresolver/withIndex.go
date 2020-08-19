package pathresolver

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/nonanick/impatience/files"
)

// WithIndex Tries to resolve the file by adding index.html
func WithIndex(
	path string,
	public string,
) (string, error) {

	if !strings.HasPrefix(path, public) {
		path = filepath.Join(public, path)
	}

	path = filepath.Join(path, "index.html")

	if files.IsKnown(path) {
		return path, nil
	}

	return "", errors.New("Could not locate path with index.html inside known files")
}
