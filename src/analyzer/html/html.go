package html

import (
	"mime"
	"regexp"
	"strings"

	"github.com/nonanick/impatience/analyzer"
)

var jsImporter, _ = regexp.Compile("<script(.*)?src=(?P<path>\".*\"|'.*').*>[\\s\\S]*?<\\/script>")
var cssImporter, _ = regexp.Compile("<link(.*)?href=(?P<path>\".*\"|'.*').*>")
var imgImporter, _ = regexp.Compile("<img.*src=(?P<path>\".*\"|'.*').*/>")

var dependeciesMatcher = []*regexp.Regexp{
	jsImporter,
	cssImporter,
	imgImporter,
}

// HTMLAnalyzer - Open and analyzes a JS file searcing for its dependencies
var HTMLAnalyzer = func(file string, content []byte) []string {

	allDependencies := make([]string, 0)
	htmlFile := content
	// Iterate though all RegExp 'macthers'
	for _, matcher := range dependeciesMatcher {
		allDependencies = append(allDependencies, analyzer.FindCaptureGroupMatches("path", htmlFile, matcher)...)
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
	mime.AddExtensionType(".html", "text/html")
	analyzer.ForExtension(".html", htmlAnalyzer)
}
