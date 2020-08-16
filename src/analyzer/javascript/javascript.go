package javascript

import (
	"bytes"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/nonanick/impatience/transform/nodeModules"

	"github.com/nonanick/impatience/analyzer"
)

// useNodeResolution inidicate to server if it should use node resolution
// (by searching inside node_modules for libraries!)
var useNodeResolution = false

var defaultImporter, rgErr = regexp.Compile("import.*from.*(?P<path>\".*\"|'.*')(?:;?)")

var dependeciesMatcher = []*regexp.Regexp{
	defaultImporter,
}

var requiredNodeLibraries = map[string]bytes.Buffer{}

// UseNodeModules intructs the javascript analyzer to use NodeModules transformer
func UseNodeModules() {
	useNodeResolution = true
}

// JsAnalyzer - Open and analyzes a JS file searching for its dependencies
var JsAnalyzer = func(file string) []string {

	allDependencies := make([]string, 0)

	jsFile, fileErr := os.Open(file)
	if fileErr != nil {
		log.Fatal("Failed to open JS file!", file, fileErr)
	}
	defer jsFile.Close()

	readBuffer := new(bytes.Buffer)
	readBuffer.ReadFrom(jsFile)

	// Iterate though all RegExp 'macthers'
	for _, matcher := range dependeciesMatcher {
		allDependencies = append(allDependencies, analyzer.FindCaptureGroupMatches("path", readBuffer, matcher)...)
	}

	strippedDeps := []string{}

	// Remove double/single quotes from string
	for _, deps := range allDependencies {
		removeSingleQuotes := strings.ReplaceAll(deps, "'", "")
		strippedPath := strings.ReplaceAll(removeSingleQuotes, "\"", "")

		// File does not end with '.js' and is a relative "." or absolute path "/"
		if !strings.HasSuffix(strippedPath, ".js") && (strings.HasPrefix(strippedPath, ".") || strings.HasPrefix(strippedPath, "/")) {
			strippedDeps = append(strippedDeps, strippedPath+".js")
		} else
		// File does not end with '.js' and is NOT a relative or absolute path
		if !(strings.HasPrefix(removeSingleQuotes, ".") || strings.HasPrefix(strippedPath, "/")) {
			// Use node resolution?
			if useNodeResolution {
				strippedDeps = append(strippedDeps, nodeModules.NodeLibraryPreffix+strippedPath)
			} else {
				strippedDeps = append(strippedDeps, nodeModules.NodeLibraryPreffix+strippedPath)
			}
		} else {
			strippedDeps = append(strippedDeps, strippedPath)
		}
	}

	return strippedDeps
}

var javascriptAnalyzer = analyzer.ExtensionAnalyzer{
	Name:     "",
	Analyzer: JsAnalyzer,
}

// Register - Register in the Analyzer the JSAnalyzer function
func Register() {
	analyzer.RegisterExtensionAnalyzer(".js", javascriptAnalyzer)
}

// AddMatcher Add a RegExp that will match the path for the dependency inside of the file
// expects a named capture group 'path', if the matchers returns a match but there's no
// capture group the match will be ignored!
func AddMatcher(matcher *regexp.Regexp) {
	dependeciesMatcher = append(dependeciesMatcher, matcher)
}
