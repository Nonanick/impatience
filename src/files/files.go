package files

import (
	"errors"
	"fmt"
	"io/ioutil"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nonanick/impatience/analyzer"
	"github.com/nonanick/impatience/cache"
	"github.com/nonanick/impatience/options"
	"github.com/nonanick/impatience/transform"
)

// File hold all the information of a file necessary
// to server the file through impatience
type File struct {
	PublicPath string
	Path       string
	Name       string
	Dir        string

	Extension string
	MimeType  string

	Bytes []byte

	Etag         string
	LastModified string

	AnalyzedBy    []string
	TransformedBy []string
	Dependencies  []string

	Size uint32
}

// knownFiles easy way to check if files is known / being tracked
var knownFiles = map[string]bool{}

// hold all the known/tracked files inside Impatience
var allFiles = map[string]File{}

// publicKnownFiles easy way to check if a public request is known
var publicKnownFiles = map[string]bool{}

// publicMap maps a public path to a real file path
var publicMap = map[string]string{}

// All Return all tracked files
func All() []File {
	all := make([]File, len(allFiles))
	for _, f := range allFiles {
		all = append(all, f)
	}
	return all
}

// Create a file definition whitout adding it to the 'known' files
func Create(file string) (File, error) {

	fileStats, statErr := os.Stat(file)
	if statErr != nil {
		return File{}, errors.New("Failed to obtain file stats: " + statErr.Error())
	}

	ext := filepath.Ext(file)
	directory, name := filepath.Split(file)
	mimeType := mime.TypeByExtension(ext)

	fileDef := File{
		PublicPath: strings.Replace(file, options.PublicRoot, "", 0),
		Path:       file,
		Dir:        directory,
		Name:       name,

		Bytes: []byte{},

		Extension: ext,
		MimeType:  mimeType,

		LastModified: fileStats.ModTime().String(),
		Etag:         cache.CalculateHash(file, fileStats.ModTime().String()),

		AnalyzedBy:    []string{},
		TransformedBy: []string{},
		Dependencies:  []string{},

		Size: uint32(fileStats.Size()),
	}

	go processFile(&fileDef)

	return fileDef, nil
}

// Will transform and analyze the file!
func processFile(file *File) {
	applyTransformers(file)
	analyzeFile(file)
}

func applyTransformers(file *File) *File {
	// Has file transformers associated ?
	if transform.HasFileTransformer(file.Path) {
		newContent := transform.Transform(file.Path)
		file.Bytes = newContent
	}

	return file
}

func analyzeFile(file *File) *File {
	if analyzer.HasAssociatedAnalyzer(file.Path) {
		dependencies := analyzer.AnalyzeFile(file.Path, file.GetContent())
		absoluteDependencies := []string{}

		// Add relative path if not absolute
		for _, dep := range dependencies {

			// Fails to identify / as absolute on windows
			adaptedSlashes := strings.ReplaceAll(dep, "/", string(os.PathSeparator))
			if strings.HasPrefix(adaptedSlashes, string(os.PathSeparator)) {
				absoluteDependencies = append(absoluteDependencies, filepath.Join(options.PublicRoot, dep))
			} else {
				absoluteDependencies = append(absoluteDependencies, filepath.Join(file.Dir, dep))
			}
		}
		file.Dependencies = absoluteDependencies
	}

	return file
}

// Add add a new physical file to the known/tracked files
// uses Create to generate the file definition from given
// file path
func Add(file string) (*File, error) {
	fileDef, crtErr := Create(file)

	if crtErr != nil {
		return &File{}, errors.New("Failed to create file definition! " + crtErr.Error())
	}

	allFiles[file] = fileDef
	knownFiles[file] = true

	// Update public file information
	publicKnownFiles[fileDef.PublicPath] = true
	publicMap[fileDef.PublicPath] = file

	return &fileDef, nil
}

// AddDefinition add a new file definition
func AddDefinition(file File) {
	allFiles[file.Path] = file
	knownFiles[file.Path] = true

	// Update public file information
	publicKnownFiles[file.PublicPath] = true
	publicMap[file.PublicPath] = file.Path
}

// Update asks for the system to update a file definition
func Update(file string) (*File, error) {

	if IsKnown(file) {
		fileInfo := allFiles[file]

		// Update LastModified and ETag
		fileInfo.LastModified = time.Now().String()
		fileInfo.Etag = cache.CalculateHash(file, fileInfo.LastModified)

		fileInfo.Bytes = []byte{}
		fileInfo.Dependencies = []string{}

		// Reapply transformers
		applyTransformers(&fileInfo)

		// Dependencies need to be updated aswell!
		analyzeFile(&fileInfo)

		allFiles[file] = fileInfo

		return &fileInfo, nil
	}

	return Add(file)

}

// Remove removes a file from the Known files, it will not delete the file from
// file system
func Remove(file string) {
	knownFiles[file] = false
	delete(allFiles, file)
}

// IsKnown either the file is known to Impatience
func IsKnown(file string) bool {
	return knownFiles[file] != false
}

// IsPubliclyKnown if the  public path is known to Impatience
func IsPubliclyKnown(publicPath string) bool {
	return publicKnownFiles[publicPath] != false
}

// Get will return a File definition or nil if the file
// is not known
func Get(file string) *File {
	f := allFiles[file]
	return &f
}

// GetPublic return a file definition or nil if the file is not found
// using its public path
func GetPublic(publicPath string) *File {
	fPath := publicMap[publicPath]
	return Get(fPath)
}

// MapEtags return all known etags
func MapEtags() map[string]string {
	var etags = map[string]string{}

	for path, file := range allFiles {
		etags[path] = file.Etag
	}

	return etags
}

// RecognizeEtag check if the etag is known to files
func RecognizeEtag(tag string) bool {
	for _, file := range allFiles {
		if file.Etag == tag {
			return true
		}
	}

	return false
}

// WasTransformed check if the file was transformed and have
// its transformed bytes on memory
func (f *File) WasTransformed() bool {
	return len(f.Bytes) > 0
}

// TrueSize return the size of the bytes that shall be transferred
// when a file is transformed its size may diverge from the original
// file
func (f *File) TrueSize() uint32 {
	if f.WasTransformed() {
		return uint32(len(f.Bytes))
	}
	return f.Size

}

// GetContent return the file content, when the file is transformed
// it will return the transformed bytes, else it shall read from the
// file system
func (f *File) GetContent() []byte {
	if f.WasTransformed() {
		return f.Bytes
	}

	content, readErr := ioutil.ReadFile(f.Path)
	if readErr != nil {
		fmt.Println("Could not read file content of ", f.Path, " file system returned error: ", readErr)
		return []byte{}
	}

	return content
}
