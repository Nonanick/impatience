package watcher

import (
	"fmt"
	"log"
	"os"

	"github.com/nonanick/impatience/impatienceserver"

	"github.com/fsnotify/fsnotify"
)

// Watch watch for directory changes
func Watch() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalln("Failed to watch public directory!", err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go handleFSWatchEvents(watcher)
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
	fmt.Println("Removed file!", file)
}

func trackNewFile(file string) {
	fmt.Println("New file created!", file)
}

func updateFileLastModTime(file string) {

	stats, err := os.Stat(file)
	if err != nil {
		log.Println("Failed to update status of file", file, err)
	}

	newModTime := stats.ModTime()

	fileStats := (*impatienceserver.KnownFilesStats)[file]
	fileStats.LastModified = fmt.Sprint(newModTime)
	(*impatienceserver.KnownFilesStats)[file] = fileStats
}
