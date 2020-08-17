package pathresolver

import (
	"errors"
	"path/filepath"
)

// Relative Path resolver that check if path is relative to public folder and is known by impatience server
func Relative(
	path string,
	public string,
	knownFiles []string,
) (string, error) {

	absPath := filepath.Join(public, path)

	if HasKnownPath(absPath, knownFiles) {
		return absPath, nil
	}

	return "", errors.New("Could not locate path as realtive path inside known files")
}
