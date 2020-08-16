package html

import (
	"bytes"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/nonanick/impatience/analyzer"
)

var jsImporter, _ = regexp.Compile("<script.*src=(?P<path>\".*\"|'.*').*>.*<\\/script>")
var cssImporter, _ = regexp.Compile("<link.*href=(?P<path>\".*\"|'.*').*>")

var dependeciesMatcher = []*regexp.Regexp{
	jsImporter,
	cssImporter,
}

// HTMLAnalyzer - Open and analyzes a JS file searcing for its dependencies
var HTMLAnalyzer = func(file string) []string {

	allDependencies := make([]string, 0)

	htmlFile, fileErr := os.Open(file)
	if fileErr != nil {
		log.Fatal("Failed to open JS file!", file, fileErr)
	}
	defer htmlFile.Close()

	readBuffer := new(bytes.Buffer)
	readBuffer.ReadFrom(htmlFile)

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

var htmlAnalyzer = analyzer.ExtensionAnalyzer{
	Name:     "HTML Analyzer",
	Analyzer: HTMLAnalyzer,
}

// Register - Register in the Analyzer the JSAnalyzer function
func Register() {
	analyzer.RegisterExtensionAnalyzer(".html", htmlAnalyzer)
}
