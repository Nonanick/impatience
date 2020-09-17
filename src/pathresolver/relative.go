package pathresolver

import (
	"errors"
	"path/filepath"

	"github.com/nonanick/impatience/files"
	"github.com/nonanick/impatience/options"
)

// Relative Path resolver that check if path is relative to public folder and is known by impatience server
func Relative(
	path string,
	public string,
) (string, error) {

	absPath := filepath.Join(options.PublicRoot, path)

	if files.IsKnown(absPath) {
		return absPath, nil
	}

	return "", errors.New("Could not locate path as realtive path inside known files")
}
