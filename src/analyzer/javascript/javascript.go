package javascript

import (
	"mime"
	"regexp"
	"strings"

	"github.com/kr/pretty"
	"github.com/nonanick/impatience/analyzer"
)

// ImportRegExp regular expression used to find import statements in js syntax
var ImportRegExp = regexp.MustCompile(
	"import\\s*(?P<name>.*)\\s*from.*(?P<path>\".*\"|'.*')(?:;?)",
)

// RequireRegExp regular expression used to fin require statements in js syntax
var RequireRegExp = regexp.MustCompile(
	"const\\s*(?P<name>.*)\\s*=\\s*require\\(\\s*(?P<path>\".*\"|'.*')\\s*\\)(?:;?)",
)

var dependeciesRegExp = []*regexp.Regexp{
	ImportRegExp,
	RequireRegExp,
}

// JsAnalyzer - Open and analyzes a JS file searching for its dependencies
var JsAnalyzer = func(file string, content []byte) []string {

	allDependencies := make([]string, 0)

	var jsFile []byte = content

	// Iterate though all RegExp 'macthers'
	for _, matcher := range dependeciesRegExp {
		allDependencies = append(allDependencies, analyzer.FindCaptureGroupMatches("path", jsFile, matcher)...)
	}

	strippedDeps := []string{}

	// Remove double/single quotes from string
	for _, deps := range allDependencies {
		removeSingleQuotes := strings.ReplaceAll(deps, "'", "")
		strippedPath := strings.ReplaceAll(removeSingleQuotes, "\"", "")

		// File does not end with '.js' and is a relative "." or absolute path "/"
		if !strings.HasSuffix(strippedPath, ".js") && (strings.HasPrefix(strippedPath, ".") || strings.HasPrefix(strippedPath, "/")) {
			strippedDeps = append(strippedDeps, strippedPath)
		} else
		// File does not end with '.js' and is NOT a relative or absolute path
		if !(strings.HasPrefix(removeSingleQuotes, ".") || strings.HasPrefix(strippedPath, "/")) {
			// Probably node module!
			strippedDeps = append(strippedDeps, strippedPath)
			pretty.Println(
				"JS Analyzer found a non relative path that does not contain an .js extension:",
				strippedPath,
				"\nIs it a node module?",
			)
		} else {
			strippedDeps = append(strippedDeps, strippedPath)
		}
	}

	return strippedDeps
}

var javascriptAnalyzer = analyzer.ExtensionAnalyzer{
	Name:     "Javascript Analyzer",
	Analyzer: JsAnalyzer,
}

// Register - Register in the Analyzer the JSAnalyzer function
func Register() {
	mime.AddExtensionType(".js", "text/javascript")
	analyzer.ForExtension(".js", javascriptAnalyzer)
}

// AddMatcher Add a RegExp that will match the path for the dependency inside of the file
// expects a named capture group 'path', if the matchers returns a match but there's no
// capture group the match will be ignored!
func AddMatcher(matcher *regexp.Regexp) {
	dependeciesRegExp = append(dependeciesRegExp, matcher)
}
