package pathresolver

import (
	"errors"
	"path/filepath"
	"strings"
)

// ResolveWithExtensions list of all extensions that will be tried
var ResolveWithExtensions []string = []string{".js", ".css", ".html", ".ts"}

// WithExtension Tries to resolve the file by adding known extensions
func WithExtension(
	path string,
	public string,
	knownFiles []string,
) (string, error) {

	if !strings.HasPrefix(path, public) {
		path = filepath.Join(public, path)
	}

	for _, ext := range ResolveWithExtensions {
		withExt := path + ext

		if HasKnownPath(withExt, knownFiles) {
			return withExt, nil
		}
	}

	return "", errors.New("Could not locate path with extensions inside known files")
}
