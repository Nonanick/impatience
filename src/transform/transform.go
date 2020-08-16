package transform

import (
	"bytes"
	"mime"
	"path/filepath"
)

var registeredTransformers map[string][]FileTransformer

// File structure containing a "transformed" file
type File struct {
	Bytes    *bytes.Buffer
	MimeType string
	Path     string
}

// HasFileTransformer Check if the file has an associated transformer
// the file extension is used to determine if the file actually has
// a transformer associated with it
func HasFileTransformer(filePath string) bool {
	extension := filepath.Ext(filePath)
	return len(registeredTransformers[extension]) > 0
}

// ApplyTransformers apply all transformers associated with an extension
// Path is an empty string and is added just to facilitate setting its
// value after function call
func ApplyTransformers(extension string, bytes *bytes.Buffer) File {

	if len(registeredTransformers[extension]) > 0 {
		var newBytes = bytes

		for _, transformer := range registeredTransformers[extension] {
			newBytes = transformer(newBytes)
		}
	}

	mimeType := mime.TypeByExtension(extension)

	return File{
		MimeType: mimeType,
		Bytes:    bytes,
		Path:     "",
	}
}

// FileTransformer Function that "transforms" a file bytes
// it should modify the bytes and return
type FileTransformer = func(bytes *bytes.Buffer) *bytes.Buffer
