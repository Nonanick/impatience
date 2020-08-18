package impatienceserver

import (
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kr/pretty"
	"github.com/nonanick/impatience/analyzer"
	"github.com/nonanick/impatience/cache"
	"github.com/nonanick/impatience/crawler"
	"github.com/nonanick/impatience/pathresolver"
	"github.com/nonanick/impatience/transform"
)

// HeaderInstructionNoPush Instruct impatience to avoid pushing
const (
	HeaderInstructionNoPush = "X-No-Further-Pushs"
	CookieCachedFilesName   = "ImpatienceCacheState"
	CookieFileSeparator     = "_&_"
)

// Port which will be used to run Impatience Server
var Port uint16 = 443

// PublicRoot Defines the absolute public folder to be served
var PublicRoot string

// HTTPSCertificatePath define the path of the https file
var HTTPSCertificatePath = "../cert/cert.cer"

// HTTPSKeyPath define th path of the HTTPSKeyPath
var HTTPSKeyPath = "../cert/key.key"

// KnownFiles Hold all the absolute path for the known files
var KnownFiles []string

// KnownFilesStats hold all the states from known files
var KnownFilesStats *map[string]crawler.FileCrawlInfo

// KnownFilesHashs hold the calculated hashs of the known files
var KnownFilesHashs *map[string]string

// MaxPushSizeInBytes hold the max size allowed to be pushed
var MaxPushSizeInBytes uint32 = 999999999

// MaxPushDependencyDepth hold the max dependency depth that will be pushed when a
// file is requested
var MaxPushDependencyDepth uint8 = 3

// FileDependencies Map containing all file dependencies for a known file
var FileDependencies *map[string][]string

// TransformedFiles Map containing all transformed files as byte
var TransformedFiles *map[string]transform.File

// Configure configure the ImpatienceServer with given properties
func Configure(config *ImpatienceConfig) {

	KnownFiles = config.KnownFiles
	KnownFilesStats = config.KnownFilesStats
	PublicRoot = config.Root
	Port = config.Port
	FileDependencies = config.FileDependencies

	knownHashs := make(map[string]string, 0)

	for path, file := range *KnownFilesStats {
		knownHashs[path] = cache.CalculateHash(file)
	}

	KnownFilesHashs = &knownHashs

}

// Launch will launch the Impatience HTTP2 server
func Launch() *http.Server {

	server := http.Server{
		Addr:    "localhost:" + fmt.Sprint(Port),
		Handler: http.HandlerFunc(HandleHTTP),
	}

	fmt.Println("---------------------\nLaunching ImpatienceServer at port :", Port)
	serverErr := server.ListenAndServeTLS(HTTPSCertificatePath, HTTPSKeyPath)

	if serverErr != nil {
		log.Fatal("Failed to start server in address 443 with provided certifcate and key!", serverErr)
	}

	return &server
}

// HandleHTTP function used to handle all HTTP requests
func HandleHTTP(
	response http.ResponseWriter,
	request *http.Request,
) {

	// Check if server can push
	push, canPush := response.(http.Pusher)

	if canPush {
		handlePushableRequest(response, push, request)
		return
	}

	handleRequest(response, request)
	return

}

// UpdateFile update file information inside the server
func UpdateFile(file crawler.FileCrawlInfo) {
	(*KnownFilesStats)[file.FilePath] = file
	(*KnownFilesHashs)[file.FilePath] = cache.CalculateHash(file)
	pretty.Println("Updating file info!")
	UpdateDependenciesForFile(file.FilePath)
}

// RemoveFile remove all information regarding a tracked file in server
func RemoveFile(file string) {
	(*KnownFilesStats)[file] = crawler.FileCrawlInfo{}
	(*KnownFilesHashs)[file] = ""

	newKnownFiles := []string{}

	for _, f := range KnownFiles {
		if f != file {
			newKnownFiles = append(newKnownFiles, f)
		}
	}

	KnownFiles = newKnownFiles

	pretty.Println("Removed file from tracked files: ", file)

}

// AddFile add a file to be tracked by the server
func AddFile(file string) {

	addedFile, fileErr := os.Open(file)

	if fileErr != nil {
		log.Fatal("Could not open added file ", file)
	}

	fileCrawlInfo := crawler.GetFileInfo(file, addedFile)

	(*KnownFilesStats)[file] = fileCrawlInfo
	(*KnownFilesHashs)[file] = cache.CalculateHash(fileCrawlInfo)

	UpdateDependenciesForFile(file)

	KnownFiles = append(KnownFiles, file)

	pretty.Println("Added to tracked files: ", file)
}

// UpdateDependenciesForFile re-analyzes files when they are updated
// to be pushed
func UpdateDependenciesForFile(file string) {
	if analyzer.HasAssociatedAnalyzer(file) {

		deps := analyzer.AnalyzeFile(file)
		absPaths := []string{}

		for _, p := range deps {
			absPaths = append(absPaths, filepath.Join(PublicRoot, p))
		}

		(*FileDependencies)[file] = absPaths

		if len(absPaths) > 0 {
			pretty.Println("File", file, "specified this dependencies:", absPaths)
		}

	}
}

// handlePushableRequest handle a HTTP request that supports HTTP2 Push capabilities
func handlePushableRequest(
	response http.ResponseWriter,
	push http.Pusher,
	request *http.Request,
) {

	path, pErr := pathresolver.Resolve(request.RequestURI, PublicRoot, KnownFiles)
	if pErr != nil {
		http.Error(response, "Failed to find path "+request.RequestURI+"<br />", 404)
		return
	}

	var servedFiles = []string{path}

	cachedFiles := cache.Extract(
		request,
		*KnownFilesHashs,
	)

	// Check if requested file is in cache
	pathHash := (*KnownFilesHashs)[path]
	fileInfo := (*KnownFilesStats)[path]

	if hashExistsInCache(pathHash, cachedFiles) && !isPushRequest(request) && acceptsCache(request, path) {

		response.Header().Add("date", time.Now().String())
		response.Header().Add("Contenty-Type", mime.TypeByExtension(fileInfo.Extension))
		response.Header().Add("Contenty-Length", fmt.Sprint(fileInfo.Size))

		response.Header().Add("Cache-Control", "max-age=9000, private, must-revalidate")
		response.Header().Add("ETag", pathHash)

		http.Error(response, "Not Modified", 304)
		//response.WriteHeader(http.StatusNotModified)
		//http.Error(response, "Not Modified", 304)
		//response.WriteHeader(304)
		//response.Write([]byte("Not Modified"))
		return
	}

	if !isPushRequest(request) {
		var totalSize uint32 = 0
		fileDeps := FlattenDependencies(path, 0, map[string]bool{}, &totalSize)

		for _, filePush := range fileDeps {
			if truePath, err := pathresolver.Resolve(filePush, PublicRoot, KnownFiles); err == nil {
				fileHash := (*KnownFilesHashs)[truePath]
				fmt.Println("FileHash", fileHash)

				if !hashExistsInCache(fileHash, cachedFiles) {
					//if true {
					pushFile(push, filePush)
					servedFiles = append(servedFiles, filePush)
				}
			} else {
				fmt.Println("WARN: File", path, "declares the dependency", filePush, "but it's not present in public directory!")
			}
		}
	}

	fileStats := (*KnownFilesStats)[path]
	respFile := getBytesOfFile(path)
	mimeType := mime.TypeByExtension(fileStats.Extension)

	hashes := []string{pathHash}

	for _, servedFile := range servedFiles {
		hashes = append(hashes, (*KnownFilesHashs)[servedFile])
	}

	if !isPushRequest(request) {
		cache.Insert(
			response,
			cachedFiles,
			hashes,
			*KnownFilesHashs,
		)
	}

	response.Header().Add("Content-Type", mimeType)
	response.Header().Add("ETag", pathHash)
	response.Header().Add("Cache-Control", "max-age=9000, private, must-revalidate")
	response.Write(respFile)

}

func acceptsCache(request *http.Request, file string) bool {
	var cacheControl = "no-cache"
	var checkEtag = ""

	if len(request.Header["Cache-Control"]) > 0 {
		cacheControl = request.Header["Cache-Control"][0]
	}

	if len(request.Header["If-None-Match"]) > 0 {
		checkEtag = request.Header["If-None-Match"][0]
	}

	// Has If none match?
	if cacheControl == "max-age=0" {
		pretty.Println("CacheControl is max-age", checkEtag)
		return (*KnownFilesHashs)[file] == checkEtag && checkEtag != ""
	}

	return false
}

func isPushRequest(request *http.Request) bool {
	return len(request.Header[HeaderInstructionNoPush]) > 0
}

func hashExistsInCache(fileHash string, cache map[string]bool) bool {
	return cache[fileHash] == true
}

// FlattenDependencies flatten all dependencies in one single array up to Max Depth, Max Size
func FlattenDependencies(path string, depth uint8, previousDependencies map[string]bool, sizeAmount *uint32) []string {

	// Depth extrapolates?
	if depth > MaxPushDependencyDepth {
		return []string{}
	}

	var allDependencies = []string{}

	if len((*FileDependencies)[path]) > 0 {

		for _, pathDep := range (*FileDependencies)[path] {
			stat := (*KnownFilesStats)[pathDep]

			// Extrapolate max size?
			if *sizeAmount+stat.Size > MaxPushSizeInBytes {
				break
			}

			*sizeAmount += stat.Size
			previousDependencies[pathDep] = true

			FlattenDependencies(pathDep, depth+1, previousDependencies, sizeAmount)
		}
	}

	for dep := range previousDependencies {
		allDependencies = append(allDependencies, dep)
	}

	return allDependencies
}

// getBytesOfFile return bytes of a KnownFile first seraching for transformed files
// falling back to disk when they don't exists on 'transformed' memory
func getBytesOfFile(file string) []byte {
	// Has transformed file?
	transfFile := transform.GetTransformedFile(file)

	if len(transfFile.Bytes) > 0 {
		return transfFile.Bytes
	}

	diskFile, err := ioutil.ReadFile(file)

	if err == nil {
		return diskFile
	}

	log.Println("Could not read any bytes from file ", file, "!Error: ", err)

	return []byte{}

}

// pushFile pushs a file in a HTTP2 connection, the function only "invoke"
// the push request, the file is actually sent by "handlePushableRequest"
// the only way to "differ" them is the presence of the custom header
// "X-No-Further-Pushs"
func pushFile(push http.Pusher, file string) {
	ext := filepath.Ext(file)
	mimeType := mime.TypeByExtension(ext)

	removePubRoot := file[len(PublicRoot):]

	opts := http.PushOptions{
		Header: map[string][]string{
			"Content-Type":          []string{mimeType},
			"Cache-Control":         []string{"max-age=600"},
			HeaderInstructionNoPush: []string{"true"},
		},
		Method: "GET",
	}
	err := push.Push(strings.ReplaceAll(removePubRoot, "\\", "/"), &opts)
	if err != nil {
		pretty.Println("Failed to push file", removePubRoot, err)
	}
}

// handleRequest handle a non pushable request (can be HTTP1/1 or HTTP2 client with
// 'disallow push' policy) this solution won't use t
func handleRequest(response http.ResponseWriter, request *http.Request) {
	response.Write([]byte("Does not accept push requests!"))
}

// ImpatienceConfig Structure holding all the required configuration for Impatience
type ImpatienceConfig struct {
	Root             string
	KnownFiles       []string
	KnownFilesStats  *map[string]crawler.FileCrawlInfo
	FileDependencies *map[string][]string
	Port             uint16
}
