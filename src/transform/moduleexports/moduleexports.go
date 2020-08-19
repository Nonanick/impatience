package moduleexports

import "github.com/nonanick/impatience/transform"

// Register Module Exports transformer
func Register() {
	transform.AddFileTransformer(".js", Transform)
}

// Transform tries to modify all module.export syntax
// to ES6 export
func Transform(path string, content []byte) []byte {

	return content
}
