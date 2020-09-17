package crawler

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/nonanick/impatience/files"
)

// IgnoredDirectories specifies which directory names should be avoided
var IgnoredDirectories = []string{
	"node_modules",
}

// Crawl will search for all the files and directories inside the root
func Crawl(root string) DirectoryGraph {

	rootDir, err := os.Open(root)
	if err != nil {
		log.Fatal("Failed to open Root Directory!", err)
	}

	rootGraph := crawlDirectory(root, rootDir)
	rootCloseErr := rootDir.Close()
	if rootCloseErr != nil {
		log.Fatal("Failed to close root directory", rootCloseErr)
	}

	return rootGraph
}

func crawlDirectory(
	dirPath string,
	directory *os.File,
) DirectoryGraph {

	innerDirectories := make([]DirectoryGraph, 0)
	innerFiles := make([]*files.File, 0)
	parentPath, directoryName := filepath.Split(directory.Name())

	// Is in ignored directories?
	for _, ignored := range IgnoredDirectories {
		if ignored == directoryName {
			return DirectoryGraph{
				ParentFolder: parentPath,
				Name:         directoryName,
				ChildLength:  uint16(0),
				Directories:  innerDirectories,
				Files:        innerFiles,
			}
		}
	}

	allDirChildren, childErr := directory.Readdir(0)
	if childErr != nil {
		log.Fatal("Could not list root directory!", childErr)
	}

	childLength := len(allDirChildren)

	for _, fileOrDir := range allDirChildren {

		if fileOrDir.IsDir() {

			dirPath := filepath.Join(dirPath, fileOrDir.Name())
			childDir, err := os.Open(dirPath)
			if err != nil {
				log.Fatal("Could not open child directory ", fileOrDir.Name())
			}

			directoryCrawl := crawlDirectory(dirPath, childDir)
			innerDirectories = append(innerDirectories, directoryCrawl)

			closeErr := childDir.Close()
			if closeErr != nil {
				log.Fatalln("Failed to close Root directory")
			}
		} else {

			filePath := filepath.Join(dirPath, fileOrDir.Name())
			fileInfo, addErr := files.Add(filePath)

			if addErr != nil {
				fmt.Println("Failed to add file ", filePath, ", returned error: ", addErr)
			} else {
				innerFiles = append(innerFiles, fileInfo)
			}
		}
	}

	// directory.
	return DirectoryGraph{
		ParentFolder: parentPath,
		Name:         directoryName,
		ChildLength:  uint16(childLength),
		Directories:  innerDirectories,
		Files:        innerFiles,
	}
}

// AllFilenames Return all file absolute paths inside the given directory graph
//
func AllFilenames(graph DirectoryGraph) []string {
	allFiles := make([]string, 0)

	for _, fileInfo := range graph.Files {
		allFiles = append(allFiles, filepath.Join(fileInfo.Dir, fileInfo.Name))
	}

	for _, dir := range graph.Directories {
		allFiles = append(allFiles, AllFilenames(dir)...)
	}

	return allFiles
}

// AllFiles Return all files.File found inside directory graph
func AllFiles(graph DirectoryGraph) []*files.File {
	allFiles := make([]*files.File, 0)

	for _, fileInfo := range graph.Files {
		allFiles = append(allFiles, fileInfo)
	}

	for _, dir := range graph.Directories {
		allFiles = append(allFiles, AllFiles(dir)...)
	}

	return allFiles
}

//DirectoryGraph -
type DirectoryGraph struct {
	ParentFolder string
	Name         string
	ChildLength  uint16
	Directories  []DirectoryGraph
	Files        []*files.File
}
