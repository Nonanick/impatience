package transform

import (
	"mime"
	"path/filepath"
)

var registeredTransformers = map[string][]FileTransformer{}

var transformedFiles = map[string]File{}

// File structure containing a "transformed" file
type File struct {
	Bytes    *[]byte
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
func ApplyTransformers(extension string, bytes *[]byte) File {

	if len(registeredTransformers[extension]) > 0 {
		var newBytes = *bytes

		for _, transformer := range registeredTransformers[extension] {
			newBytes = transformer(&newBytes)
		}

		bytes = &newBytes
	}

	mimeType := mime.TypeByExtension(extension)

	return File{
		MimeType: mimeType,
		Bytes:    bytes,
		Path:     "",
	}
}

// AddFileTransformer adds a file transformer to an extension
func AddFileTransformer(extension string, transformer FileTransformer) {
	registeredTransformers[extension] = append(registeredTransformers[extension], transformer)
}

// IsFileTransformed return if a path has a tranaformed file associates with it
func IsFileTransformed(path string) bool {
	return len(*transformedFiles[path].Bytes) > 0
}

// AddTransformedFile Add a File struct as an transformed file;
// file.Path will be used for mapping!
func AddTransformedFile(file File) {
	transformedFiles[file.Path] = file
}

// GetTransformedFile return the transformed file associated with the
// given path
func GetTransformedFile(path string) File {
	return transformedFiles[path]
}

// FileTransformer Function that "transforms" a file bytes
// it should modify the bytes and return
type FileTransformer = func(bytes *[]byte) []byte
