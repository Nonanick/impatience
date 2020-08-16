package analyzer

import (
	"path/filepath"
	"regexp"
)

var registeredAnalyzers = make(map[string][]ExtensionAnalyzer)

// RegisterExtensionAnalyzer Adds an analyzer to be used on an extension and check for dependencies
func RegisterExtensionAnalyzer(extension string, analyzer ExtensionAnalyzer) {
	registeredAnalyzers[extension] = append(registeredAnalyzers[extension], analyzer)
}

// HasAssociatedAnalyzer determine if the file has an associated analyzer
func HasAssociatedAnalyzer(filePath string) bool {
	extension := filepath.Ext(filePath)
	return len(registeredAnalyzers[extension]) > 0
}

// AnalyzeFile Analyzes the file using the extension and return all dependencies
func AnalyzeFile(filePath string) []string {
	allDependencies := []string{}

	extension := filepath.Ext(filePath)

	if len(registeredAnalyzers[extension]) > 0 {
		for _, registeredAnalyzer := range registeredAnalyzers[extension] {
			analyzedDeps := registeredAnalyzer.Analyzer(filePath)
			allDependencies = append(allDependencies, analyzedDeps...)
		}
	}

	return allDependencies
}

// FindCaptureGroupMatches find all matches of specified capture group inside buffer
func FindCaptureGroupMatches(captureGroup string, bytes *[]byte, matcher *regexp.Regexp) []string {

	allDependencies := []string{}

	// Find all submatches using this matcher
	matches := matcher.FindAllSubmatch(*bytes, -1)
	subNames := matcher.SubexpNames()

	// Find all matches -- First one is the whole match!
	for _, subMatches := range matches {
		for i, captureGroup := range subMatches {
			// Capture group name is 'path' ?
			if subNames[i] == "path" {
				allDependencies = append(allDependencies, string(captureGroup))
			}
		}
	}

	return allDependencies

}

// ExtensionAnalyzerFunc Function signature that receives a filepath and return the dependencies
type ExtensionAnalyzerFunc func(string) []string

// ExtensionAnalyzer Struct containing the name of the analyzer and its function
type ExtensionAnalyzer struct {
	Name     string
	Analyzer ExtensionAnalyzerFunc
}
