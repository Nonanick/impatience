package watcher

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/kr/pretty"
	"github.com/nonanick/impatience/files"
	"github.com/nonanick/impatience/transform/nodemodules"

	"github.com/fsnotify/fsnotify"
)

// TrackedDirectories map of all the tracked directories
var TrackedDirectories = map[string]bool{}

// IgnoreDirectories directories that should be ignored by file watcher
// all subdirectories will also be ignored!
var IgnoreDirectories = []string{}

// Watch watch for directory changes
func Watch() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalln("Failed to watch public directory!", err)
	}
	defer watcher.Close()

	done := make(chan bool)

	go handleFSWatchEvents(watcher)

	for _, file := range files.All() {

		if TrackedDirectories[file.Dir] != true &&
			!strings.HasSuffix(
				file.Dir,
				strings.ReplaceAll(
					nodemodules.NodePublicRoot,
					"/",
					string(os.PathSeparator),
				),
			) &&
			file.Path != "" {
			err := watcher.Add(file.Dir)
			if err != nil {
				pretty.Println("Failed to add directory to watcher!", file.Dir)
			} else {
				pretty.Println("Added directory to watcher!", file.Dir)
				TrackedDirectories[file.Dir] = true
			}

		}

	}
	<-done
}

func handleFSWatchEvents(watcher *fsnotify.Watcher) {
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				fmt.Println("INFO: FS Watcher reported 'not ok'...")
				return
			}

			// Write event --> Update LastModified
			if event.Op&fsnotify.Write == fsnotify.Write {
				pretty.Println("FS Watch, triggered write event!", event)
				updateFileLastModTime(event.Name)
			}
			// Create event --> add file to trackers
			if event.Op&fsnotify.Create == fsnotify.Create {
				trackNewFile(event.Name)
			}
			// Remove event --> remove file from trackers
			if event.Op&fsnotify.Remove == fsnotify.Remove {
				updateRemovedFile(event.Name)
			}
			// Rename event --> CREATE event will be triggered, removing old trackers
			if event.Op&fsnotify.Rename == fsnotify.Rename {
				updateRemovedFile(event.Name)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			fmt.Println("ERROR: FS Watcher reported an error", err)
		}
	}
}

func updateRemovedFile(file string) {
	files.Remove(file)
}

func trackNewFile(file string) {
	files.Add(file)
}

func updateFileLastModTime(file string) {
	files.Update(file)
}
