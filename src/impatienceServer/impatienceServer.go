package impatienceserver

import (
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/kr/pretty"
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
	} else {
		handleRequest(response, request)
	}

}

func handlePushableRequest(response http.ResponseWriter, push http.Pusher, request *http.Request) {

	var servedFiles = []string{}

	path, pErr := pathresolver.Resolve(request.RequestURI, PublicRoot, KnownFiles)

	if pErr != nil {
		http.Error(response, "Failed to find path "+request.RequestURI+"<br />", 404)
		return
	}

	servedFiles = append(servedFiles, path)

	cachedFiles := cache.Extract(request, *KnownFilesStats)
	pretty.Println("Cached files found: ", cachedFiles)

	if len(request.Header[HeaderInstructionNoPush]) == 0 {
		var totalSize uint32 = 0
		fileDeps := FlattenDependencies(path, 0, map[string]bool{}, &totalSize)

		for _, filePush := range fileDeps {
			if _, err := pathresolver.Resolve(filePush, PublicRoot, KnownFiles); err == nil {

				pushFile(push, filePush)
				servedFiles = append(servedFiles, filePush)
			} else {
				fmt.Println("WARN: File", path, "declares the dependency", filePush, "but it's not present in public directory!")
			}
		}
	}

	fileStats := (*KnownFilesStats)[path]
	respFile := getBytesOfFile(path)
	mimeType := mime.TypeByExtension(fileStats.Extension)

	hashes := []string{}

	for _, servedFile := range servedFiles {
		hashes = append(hashes, cache.CalculateHash((*KnownFilesStats)[servedFile]))
	}

	cookie := http.Cookie{
		HttpOnly: true,
		Name:     CookieCachedFilesName,
		Value:    strings.Join(hashes, CookieFileSeparator),
		Path:     "/",
		Secure:   true,
		Expires:  time.Now().Add(2 * 24 * 60 * 60 * 1000 * 100),
	}

	http.SetCookie(response, &cookie)
	response.Header().Add("Content-Type", mimeType)
	response.Header().Add("Cache-Control", "max-age=600")

	response.Write(respFile)

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

func handleRequest(response http.ResponseWriter, request *http.Request) {
	response.Write([]byte("Does not accept push requests!"))
}

func getRequestedFileFromHTTP2(request *http.Request) []string {
	return request.Header[":path"]
}

// ImpatienceConfig Structure holding all the required configuration for Impatience
type ImpatienceConfig struct {
	Root             string
	KnownFiles       []string
	KnownFilesStats  *map[string]crawler.FileCrawlInfo
	FileDependencies *map[string][]string
	Port             uint16
}
