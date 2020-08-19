package pathresolver

import (
	"errors"
	"path/filepath"

	"github.com/nonanick/impatience/files"
)

// Absolute Path resolver that check if path is absolute and is known by impatience server
func Absolute(
	path string,
	public string,
) (string, error) {

	if filepath.IsAbs(path) {
		if files.IsKnown(path) {
			return path, nil
		}
	}

	return "", errors.New("Could not locate path as absolute path inside known files")
}
