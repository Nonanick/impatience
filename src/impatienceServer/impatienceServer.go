package impatienceServer

import (
	"log"
	"net/http"

	"github.com/nonanick/impatience/transform"
)

// Port which will be used to run Impatience Server
var Port uint16 = 443

// PublicRoot Defines the absolute public folder to be served
var PublicRoot string

// HTTPSCertificatePath define the path of the https file
var HTTPSCertificatePath = "../../cert/cert.pem"

// HTTPSKeyPath define th path of the HTTPSKeyPath
var HTTPSKeyPath = "../../cert/key.pem"

// KnownFiles Hold all the absolute path for the known files
var KnownFiles []string

// FileDependencies Map containing all file dependencies for a known file
var FileDependencies map[string][]string

// TransformedFiles Map containing all transformed files as byte
var TransformedFiles map[string]transform.File

// Configure configure the ImpatienceServer with given properties
func Configure(config *ImpatienceConfig) {

	KnownFiles = config.KnownFiles
	PublicRoot = config.Root
	Port = config.Port
	FileDependencies = config.FileDependencies

}

// Launch will launch the Impatience HTTP2 server
func Launch() *http.Server {

	server := http.Server{
		Addr:    ":443",
		Handler: http.HandlerFunc(HandleHTTP),
	}

	serverErr := server.ListenAndServeTLS(HTTPSCertificatePath, HTTPSKeyPath)

	if serverErr != nil {
		log.Fatal("Failed to start server in address 443 with provided certifcate and key!", serverErr)
	}

	return &server
}

// HandleHTTP function used to handle all HTTP requests
func HandleHTTP(response http.ResponseWriter, request *http.Request) {

	// Check if server can push
	push, canPush := response.(http.Pusher)

	if canPush {
		handlePushableRequest(&response, &push, request)
	} else {
		handleRequest(&response, request)
	}

}

func handlePushableRequest(response *http.ResponseWriter, push *http.Pusher, request *http.Request) {

}

func handleRequest(response *http.ResponseWriter, request *http.Request) {

}

// ImpatienceConfig Structure holding all the required configuration for Impatience
type ImpatienceConfig struct {
	Root             string
	KnownFiles       []string
	FileDependencies map[string][]string
	Port             uint16
}
