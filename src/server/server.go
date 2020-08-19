package server

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/kr/pretty"
	"github.com/nonanick/impatience/cache"
	"github.com/nonanick/impatience/files"
	"github.com/nonanick/impatience/pathresolver"
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
var HTTPSCertificatePath = "../cert/cert.pem"

// HTTPSKeyPath define th path of the HTTPSKeyPath
var HTTPSKeyPath = "../cert/key.pem"

// MaxPushSizeInBytes hold the max size allowed to be pushed
var MaxPushSizeInBytes uint32 = 999999999

// MaxPushDependencyDepth hold the max dependency depth that will be pushed when a
// file is requested
var MaxPushDependencyDepth uint8 = 3

// FileDependencies Map containing all file dependencies for a known file
var FileDependencies map[string][]string = map[string][]string{}

// Configure configure the ImpatienceServer with given properties
func Configure(config *ImpatienceConfig) {
	PublicRoot = config.Root
	Port = config.Port
}

// Launch will launch the Impatience HTTP2 server
func Launch() *http.Server {

	server := http.Server{
		Addr:    "localhost:" + fmt.Sprint(Port),
		Handler: http.HandlerFunc(HandleHTTP),
	}

	fmt.Println("---------------------\nLaunching ImpatienceServer at port :", Port)
	fmt.Println("Serving static files in :", PublicRoot)
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

// handlePushableRequest handle a HTTP request that supports HTTP2 Push capabilities
func handlePushableRequest(
	response http.ResponseWriter,
	push http.Pusher,
	request *http.Request,
) {

	path, pErr := pathresolver.Resolve(request.RequestURI, PublicRoot)
	if pErr != nil {
		http.Error(response, "Failed to find path "+request.RequestURI+"<br />", 404)
		return
	}

	var requestedFile = files.Get(path)
	var servedFiles = []string{}
	var cachedFiles = cache.Extract(
		request,
		files.MapEtags(),
	)

	if isPushRequest(request) {
		pretty.Println("Is Push request for", request.RequestURI, cachedFiles)
	} else {
		pretty.Println("Is NOT Push request for", request.RequestURI)
	}
	if !isPushRequest(request) {
		var totalSize uint32 = 0
		fileDeps := FlattenDependencies(requestedFile, 0, map[string]bool{}, &totalSize)
		pretty.Println("All file dependencies flattened", fileDeps)

		for _, filePush := range fileDeps {
			if truePath, err := pathresolver.Resolve(filePush, PublicRoot); err == nil {
				depFileInfo := files.Get(truePath)

				pushFile(push, depFileInfo, filePush, cachedFiles)

				servedFiles = append(servedFiles, filePush)
			} else {
				fmt.Println("WARN: File", path, "declares the dependency", filePush, "but it's not present in public directory!")
			}
		}
	}

	hashes := []string{}
	// Was any file pushed ?
	for _, servedFile := range servedFiles {
		hashes = append(hashes, files.Get(servedFile).Etag)
	}

	// Must always accepts cache!
	//- If it exists on cache ( cookie ) or has a X-Push-304 header, push 304
	if (hashExistsInCache(requestedFile.Etag, cachedFiles) ||
		hasPush304Header(request)) &&
		acceptsCache(request, path) {

		response.Header().Add("date", time.Now().String())
		response.Header().Add("Content-Type", requestedFile.MimeType)
		response.Header().Add("Content-Length", fmt.Sprint(requestedFile.Size))

		response.Header().Add("Cache-Control", "private, must-revalidate")
		response.Header().Add("ETag", requestedFile.Etag)

		response.WriteHeader(http.StatusNotModified)

		return
	}

	// Void cookie cache if server does not accepts cache!
	if !acceptsCache(request, path) {
		cache.Insert(
			response,
			map[string]bool{},
			[]string{},
		)
	}

	// If it did not fall under 304 push requested file into server hashs!
	hashes = append(hashes, requestedFile.Etag)

	// Only push cookies when something was served!
	if !isPushRequest(request) && len(hashes) > 0 && acceptsCache(request, path) {
		cache.Insert(
			response,
			cachedFiles,
			hashes,
		)
	}

	response.Header().Add("Content-Type", requestedFile.MimeType)
	response.Header().Add("Content-Length", fmt.Sprint(requestedFile.TrueSize()))
	response.Header().Add("ETag", requestedFile.Etag)
	response.Header().Add("Cache-Control", "private, must-revalidate")
	response.Write(requestedFile.GetContent())

}

func hasPush304Header(request *http.Request) bool {
	return len(request.Header["X-Push-304"]) > 0
}

func acceptsCache(request *http.Request, file string) bool {
	var checkEtag = ""

	if len(request.Header["If-None-Match"]) > 0 {
		checkEtag = request.Header["If-None-Match"][0]
	}

	return files.Get(file).Etag == checkEtag && checkEtag != ""
}

func isPushRequest(request *http.Request) bool {
	return len(request.Header[HeaderInstructionNoPush]) > 0
}

func hashExistsInCache(fileHash string, cache map[string]bool) bool {
	return cache[fileHash] == true
}

// FlattenDependencies flatten all dependencies in one single array up to Max Depth, Max Size
func FlattenDependencies(
	file *files.File,
	depth uint8,
	previousDependencies map[string]bool,
	sizeAmount *uint32,
) []string {

	// Depth extrapolates?
	if depth > MaxPushDependencyDepth {
		return []string{}
	}

	var allDependencies = []string{}

	if len(file.Dependencies) > 0 {

		for _, pathDep := range file.Dependencies {

			// Extrapolate max size?
			if *sizeAmount+file.TrueSize() > MaxPushSizeInBytes {
				break
			}

			*sizeAmount += file.TrueSize()
			previousDependencies[pathDep] = true

			FlattenDependencies(files.Get(pathDep), depth+1, previousDependencies, sizeAmount)
		}
	}

	for dep := range previousDependencies {
		allDependencies = append(allDependencies, dep)
	}

	return allDependencies
}

// pushFile pushs a file in a HTTP2 connection, the function only "invoke"
// the push request, the file is actually sent by "handlePushableRequest"
// the only way to "differ" them is the presence of the custom header
// "X-No-Further-Pushs"
func pushFile(push http.Pusher, file *files.File, requestedURL string, cachedFiles map[string]bool) {

	removePubRoot := requestedURL[len(PublicRoot):]

	opts := http.PushOptions{
		Header: map[string][]string{
			"Content-Type":          {file.MimeType},
			"Cache-Control":         {"private, must-revalidate"},
			HeaderInstructionNoPush: {"true"},
			"X-Push-URL":            {requestedURL},
		},
		Method: "GET",
	}

	if hashExistsInCache(file.Etag, cachedFiles) {
		opts.Header["X-Push-304"] = []string{"true"}
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
	Root string
	Port uint16
}
