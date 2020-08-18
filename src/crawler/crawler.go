package crawler

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

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
	innerFiles := make([]FileCrawlInfo, 0)
	parentPath, directoryName := filepath.Split(directory.Name())

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
			childFile, fileErr := os.Open(filePath)
			if fileErr != nil {
				log.Fatal("Could not open child file ", fileOrDir.Name())
			}

			fileCrawlInfo := GetFileInfo(filePath, childFile)
			innerFiles = append(innerFiles, fileCrawlInfo)
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

// GetFileInfo Return a file crawl info from a open file
func GetFileInfo(
	filePath string,
	file *os.File,
) FileCrawlInfo {

	fileStats, statErr := os.Stat(filePath)
	if statErr != nil {

	}

	parentFolder, fileName := filepath.Split(filePath)
	extension := filepath.Ext(fileStats.Name())
	size := fileStats.Size()
	modTime := fileStats.ModTime()
	return FileCrawlInfo{
		FilePath:     filePath,
		ParentFolder: parentFolder,
		Name:         fileName,
		Extension:    extension,
		Size:         uint32(size),
		LastModified: fmt.Sprint(modTime),
	}
}

// AllFilenames Return all file absolute paths inside the given directory graph
//
func AllFilenames(graph DirectoryGraph) []string {
	allFiles := make([]string, 0)

	for _, fileInfo := range graph.Files {
		allFiles = append(allFiles, filepath.Join(fileInfo.ParentFolder, fileInfo.Name))
	}

	for _, dir := range graph.Directories {
		allFiles = append(allFiles, AllFilenames(dir)...)
	}

	return allFiles
}

// AllFiles Return all FileCrawlInfo found inside directory graph
func AllFiles(graph DirectoryGraph) []FileCrawlInfo {
	allFiles := make([]FileCrawlInfo, 0)

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
	Files        []FileCrawlInfo
}

//FileCrawlInfo - gathered
type FileCrawlInfo struct {
	ParentFolder string
	Name         string
	FilePath     string
	Extension    string
	Size         uint32
	LastModified string
}
