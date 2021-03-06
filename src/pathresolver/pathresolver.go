package pathresolver

import (
	"errors"
)

var registeredResolvers []Resolver = []Resolver{}

// AddResolver add a path resolver function
func AddResolver(resolver Resolver) {
	registeredResolvers = append(registeredResolvers, resolver)
}

// Resolve tries to resolve the path to a known file
func Resolve(
	path string,
	public string,
) (string, error) {

	for _, resolver := range registeredResolvers {
		resPath, pathErr := resolver(path, public)
		if pathErr == nil {
			return resPath, nil
		}
	}

	return "", errors.New("Could not locate path in known files")
}

// Resolver Function that tris to resolve the pathname
type Resolver = func(path string, public string) (string, error)
