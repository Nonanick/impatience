package main

import (
	"fmt"
	"path/filepath"

	"github.com/nonanick/impatience/analyzer"
	"github.com/nonanick/impatience/crawler"
)

// DependenciesMap Hold all the file paths that rely on on they key path
var DependenciesMap = map[string][]string{}

// FileInfoMap Maps the fileabsolute path to its FileCrawlInfo
var FileInfoMap = map[string]crawler.FileCrawlInfo{}

// Get Return all file dependencies from a file
func Get(filePath string, root string) []crawler.FileCrawlInfo {

	allDeps := []crawler.FileCrawlInfo{}

	return allDeps
}

// GenerateDependenciesMap populates the dependency map for files tracked by
// Impatience
func GenerateDependenciesMap(graph crawler.DirectoryGraph) map[string][]string {

	allFiles := crawler.AllFiles(graph)

	for _, file := range allFiles {

		FileInfoMap[file.FilePath] = file

		if !analyzer.HasAssociatedAnalyzer(file.FilePath) {
			fmt.Println("File ", file.Name, " does not have an associated dependency analyzer")
		}

		dependenciesFromFile := analyzer.AnalyzeFile(file.FilePath)

		for _, dep := range dependenciesFromFile {
			if filepath.IsAbs(dep) {
				DependenciesMap[file.FilePath] = append(DependenciesMap[file.FilePath], dep)
			} else {
				DependenciesMap[file.FilePath] = append(DependenciesMap[file.FilePath], filepath.Join(file.ParentFolder, dep))
			}
		}

	}
	return DependenciesMap
}
