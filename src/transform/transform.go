package transform

import (
	"io/ioutil"
	"path/filepath"

	"github.com/kr/pretty"
)

var registeredTransformers = map[string][]FileTransformer{}

// HasFileTransformer Check if the file has an associated transformer
// the file extension is used to determine if the file actually has
// a transformer associated with it
func HasFileTransformer(filePath string) bool {
	extension := filepath.Ext(filePath)
	return len(registeredTransformers[extension]) > 0
}

// Transform a file applying all transformations inside it
func Transform(file string) []byte {

	ext := filepath.Ext(file)
	content, err := ioutil.ReadFile(file)

	if err != nil {
		pretty.Println("Failed to transform file", file, " impatience could not read bytes from the original file!")

		return []byte{}
	}

	return Apply(ext, file, content)

}

// Apply apply all transformers associated with an extension
func Apply(
	extension string,
	path string,
	bytes []byte,
) []byte {

	if len(registeredTransformers[extension]) > 0 {
		var newBytes = bytes

		for _, transformer := range registeredTransformers[extension] {
			newBytes = transformer(path, newBytes)
		}

		bytes = newBytes
	}

	return bytes

}

// AddFileTransformer adds a file transformer to an extension
func AddFileTransformer(
	extension string,
	transformer FileTransformer,
) {
	registeredTransformers[extension] = append(registeredTransformers[extension], transformer)
}

// FileTransformer Function that "transforms" a file bytes
// it should modify the bytes and return
type FileTransformer = func(path string, content []byte) []byte
