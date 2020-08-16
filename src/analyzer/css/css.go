package css

import (
	"bytes"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/nonanick/impatience/analyzer"
)

var defaultImporter, rgErr = regexp.Compile("url\\s*\\((?P<path>.*|.*)\\)\\s*(?:;?)")

var dependeciesMatcher = []*regexp.Regexp{
	defaultImporter,
}

// CSSAnalyzer - Open and analyzes a JS file searcing for its dependencies
var CSSAnalyzer = func(file string) []string {

	allDependencies := make([]string, 0)

	cssFile, fileErr := os.Open(file)
	if fileErr != nil {
		log.Fatal("Failed to open CSS file!", file, fileErr)
	}
	defer cssFile.Close()

	readBuffer := new(bytes.Buffer)
	readBuffer.ReadFrom(cssFile)

	// Iterate though all RegExp 'macthers'
	for _, matcher := range dependeciesMatcher {
		allDependencies = append(allDependencies, analyzer.FindCaptureGroupMatches("path", readBuffer, matcher)...)
	}

	strippedDeps := []string{}
	// Remove double/single quotes from string
	for _, deps := range allDependencies {
		removeSingleQuotes := strings.ReplaceAll(deps, "'", "")
		removeDoubleQuotes := strings.ReplaceAll(removeSingleQuotes, "\"", "")
		strippedDeps = append(strippedDeps, removeDoubleQuotes)
	}

	return strippedDeps
}

var cssAnalyzer = analyzer.ExtensionAnalyzer{
	Name:     "",
	Analyzer: CSSAnalyzer,
}

// Register - Register in the Analyzer the JSAnalyzer function
func Register() {
	analyzer.RegisterExtensionAnalyzer(".css", cssAnalyzer)
}

// AddMatcher - Add a RegExp that will match the path for the dependency inside of the file
// expects a named capture group 'path', if the matchers returns a match but there's no
// capture group the match will be ignored!
func AddMatcher(matcher *regexp.Regexp) {
	dependeciesMatcher = append(dependeciesMatcher, matcher)
}
