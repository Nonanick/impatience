package nodemodules

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/nonanick/impatience/files"
	"github.com/nonanick/impatience/options"
	"github.com/nonanick/impatience/transform"
	"github.com/nonanick/impatience/transform/require"
)

// NodePublicRoot preffiz is used on transformed js/html files to indicate
// that a library dependency
var NodePublicRoot = "/_impatience/node/"

var registeredNodeLibraries map[string]bool = map[string]bool{}

var registeredDependencies map[string]bool = map[string]bool{}

var knownLibs = map[string]bool{}
var loadedFiles = map[string]bool{}

// AddNodeFile exposes a file present in node_modules
func AddNodeFile(file string) {

	libName := file
	targetFile := file

	// If node file has "/" its targeting a file inside of a package, the lib
	// name will be the first portion
	if strings.Index(file, "/") > 0 {
		libName = strings.Split(file, "/")[0]
	} else
	// If the file is targeting the root of the lib it necessary to check
	// its package.json for the "main" directive
	if knownLibs[libName] != true {
		mainFile, err := DiscoverLibMainFile(libName)
		if err != nil {
			fmt.Println("NodeModules cannot resolve package main file of library ", libName)
			return
		}

		targetFile = mainFile
	}

	// target file already loaded ? NOOP
	if loadedFiles[targetFile] == true {
		return
	}

	createdFile, createErr := files.Create(targetFile)
	if createErr != nil {
		fmt.Println("NodeModules could not read library file: ", file)
		return
	}
	createdFile.PublicPath = filepath.Join(options.PublicRoot, NodePublicRoot, file)

	files.AddDefinition(createdFile)
	loadedFiles[targetFile] = true

}

// DiscoverLibMainFile will check package.json to find the main file
// of the imported lib
func DiscoverLibMainFile(libName string) (string, error) {

	libPath := filepath.Join(options.PublicRoot, options.NodeModulesRoot, libName)
	// Does not know ? start discovery from entry point
	packageJSON, err := ioutil.ReadFile(filepath.Join(libPath, "package.json"))
	if err != nil {
		return "", errors.New("Could not discover package JSON inside node_modules/" + libName)
	}
	var result map[string]string
	json.Unmarshal([]byte(packageJSON), &result)

	targetFile := filepath.Join(string(libPath), string(result["main"]))

	return targetFile, nil

}

// Register add node transformers for known extensions
func Register() {
	require.Register()
	transform.AddFileTransformer(".js", NodeTransform)
}

var importMatcher = regexp.MustCompile(
	"import\\s*(?P<name>.*)\\s*from.*(?P<path>\".*\"|'.*')(?:;?)",
)

var exportMatcher = regexp.MustCompile("module.exports\\s*=\\s*(?P<name>.*?);")

// NodeTransform Transform a node module
var NodeTransform transform.FileTransformer = func(
	path string,
	content []byte,
) []byte {

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
					path = NodePublicRoot + path
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

// NodeLib node library
type NodeLib struct {
	Name       string
	PublicPath string
	FileRoot   string
}
