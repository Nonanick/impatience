package watcher

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/kr/pretty"
	"github.com/nonanick/impatience/impatienceserver"

	"github.com/fsnotify/fsnotify"
)

// TrackedDirectories map of all the tracked directories
var TrackedDirectories = map[string]bool{}

// Watch watch for directory changes
func Watch() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalln("Failed to watch public directory!", err)
	}
	defer watcher.Close()

	done := make(chan bool)

	go handleFSWatchEvents(watcher)

	for _, file := range impatienceserver.KnownFiles {

		dir, _ := filepath.Split(file)

		if TrackedDirectories[dir] != true {
			err := watcher.Add(dir)
			if err != nil {
				pretty.Println("Failed to add file to watcher!")
			} else {
				TrackedDirectories[dir] = true
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
	impatienceserver.RemoveFile(file)
}

func trackNewFile(file string) {
	impatienceserver.AddFile(file)
}

func updateFileLastModTime(file string) {

	stats, err := os.Stat(file)
	if err != nil {
		log.Println("Failed to update status of file", file, err)
	}

	newModTime := stats.ModTime()

	fileStats := (*impatienceserver.KnownFilesStats)[file]
	fileStats.LastModified = fmt.Sprint(newModTime)

	impatienceserver.UpdateFile(fileStats)
}
