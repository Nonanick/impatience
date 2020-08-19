package nodemodules

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/kr/pretty"
	"github.com/nonanick/impatience/analyzer/javascript"
	"github.com/nonanick/impatience/files"
	"github.com/nonanick/impatience/options"
	"github.com/nonanick/impatience/transform"
	"github.com/nonanick/impatience/transform/require"
)

// NodeLibraryPreffix preffiz is used on transformed js/html files to indicate
// that a library dependency
var NodeLibraryPreffix = "/_impatience/node/"

var registeredNodeLibraries map[string]bool = map[string]bool{}

var registeredDependencies map[string]bool = map[string]bool{}

// AddNodeLibrary adds a node library
func AddNodeLibrary(name string) {

}

// Register add node transformers for known extensions
func Register() {
	require.Register()
	transform.AddFileTransformer(".js", NodeTransform)
}

// InjectLibraries Add the libraries inside the 'known' files
func InjectLibraries() {

	for lib := range registeredNodeLibraries {
		libPath := filepath.Join(options.PublicRoot, options.NodeModulesRoot, lib)
		packageJSON, err := ioutil.ReadFile(filepath.Join(libPath, "package.json"))

		if err != nil {
			pretty.Println("[NodeModules]: Did not find package JSON, did you install it? ", string(packageJSON))
			continue
		}

		var result map[string]string
		json.Unmarshal([]byte(packageJSON), &result)
		libTarget := filepath.Join(string(libPath), string(result["main"]))

		// Try to load and transform file

		targetFile, addErr := files.Create(libTarget)
		if addErr != nil {
			fmt.Println("Failed to create file definition for node library!", addErr.Error())
			continue
		}

		// Remap path
		targetFile.Path = filepath.Join(options.PublicRoot, NodeLibraryPreffix, lib)
		dir, name := filepath.Split(targetFile.Path)
		targetFile.Dir = dir
		targetFile.Name = name

		files.AddDefinition(targetFile)
	}
}

// NodeTransform Transform a node module
var NodeTransform transform.FileTransformer = func(
	path string,
	content []byte,
) []byte {

	importMatcher := javascript.ImportRegExp
	exportMatcher := regexp.MustCompile("module.exports\\s*=\\s*(?P<name>.*?);")

	// Find all submatches using import matcher
	importMatches := importMatcher.FindAllSubmatch(content, -1)
	importMatchesIndex := importMatcher.FindAllIndex(content, -1)
	subNames := importMatcher.SubexpNames()

	// all transformation will be done in this
	newContent := []byte{}
	lastIndex := 0

	// Find all import matches -- First one is the whole match!
	if importMatches != nil {
		for ioSubMatch, subMatches := range importMatches {

			if len(subMatches) > 0 {
				name := subMatches[require.IndexOf("name", subNames)]

				path := strings.ReplaceAll(
					strings.ReplaceAll(
						string(subMatches[require.IndexOf("path", subNames)]),
						"'",
						"",
					),
					"\"",
					"",
				)

				startOfMatch := importMatchesIndex[ioSubMatch][0]
				endOfMatch := importMatchesIndex[ioSubMatch][1]

				if !strings.HasPrefix(path, "./") && !strings.HasPrefix(path, "/") {

					pretty.Println(
						"Probably found a node lib with import name of", string(name), "and path", string(path),
					)

					registeredNodeLibraries[path] = true
					path = NodeLibraryPreffix + path

					newContent = append(newContent, content[lastIndex:startOfMatch]...)
					newContent = append(newContent, []byte("import "+string(name)+" from '"+string(path)+"';")...)
					lastIndex = endOfMatch

				}
			}
		}
	}
	newContent = append(newContent, content[lastIndex:]...)

	// wrap up import, try to find exports
	content = newContent
	newContent = []byte{}
	lastIndex = 0
	exportName := []byte{}
	// Try exports
	exportMatches := exportMatcher.FindAllSubmatch(content, -1)
	exportMatchesIndex := exportMatcher.FindAllIndex(content, -1)
	eSubNames := exportMatcher.SubexpNames()
	if exportMatches != nil {
		for ioSubMatch, subMatches := range exportMatches {

			if len(subMatches) > 0 {
				name := subMatches[require.IndexOf("name", eSubNames)]
				exportName = name
				startOfMatch := exportMatchesIndex[ioSubMatch][0]
				endOfMatch := exportMatchesIndex[ioSubMatch][1]

				pretty.Println(
					"Probably found a node lib with module.exports", string(name),
				)

				newContent = append(newContent, content[lastIndex:startOfMatch]...)
				newContent = append(newContent, []byte("export { "+string(name)+" };")...)

				lastIndex = endOfMatch
			}
		}
	}

	newContent = append(newContent, content[lastIndex:]...)
	newContent = bytes.ReplaceAll(newContent, []byte("module.exports"), exportName)

	return newContent
}
