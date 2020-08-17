package pathresolver

import (
	"errors"
	"path/filepath"
)

// Absolute Path resolver that check if path is absolute and is known by impatience server
func Absolute(
	path string,
	public string,
	knownFiles []string,
) (string, error) {

	if filepath.IsAbs(path) {
		if HasKnownPath(path, knownFiles) {
			return path, nil
		}
	}

	return "", errors.New("Could not locate path as absolute path inside known files")
}
