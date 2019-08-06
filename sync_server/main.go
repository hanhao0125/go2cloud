package main

import (
	"log"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Package const
// Config file path name
const (
	ConfigFilePathName = "./mancy_config.json"
)

// Package variables
var (
	MountedDir string = "/Users/hanhao/server"

	// Global chan variables
	// file_watcher will write the chan and file_handle will read the chan
	// create file
	fileCreateEvent = make(chan string)

	// write
	fileWriteEvent = make(chan string)

	// remove
	fileRemoveEvent = make(chan string)

	// rename
	fileRenameEvent = make(chan string)

	// chmod
	fileChmodEvent = make(chan string)

	// watchMainJob chan
	watcherHandlerDone = make(chan bool)

	// fileHandleMainJob chan
	fileHandlerDone = make(chan bool)

	// timeout for watcher event
	fileHandleTimeOut = time.Second * 4
)

func init() {
	// Reset log format
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func main() {

	watch, _ := fsnotify.NewWatcher()

	w := Watch{
		watch: watch,
	}

	go func() {
		w.watchDir(MountedDir)
		watcherHandlerDone <- true
	}()

	// handle the file events
	go func() {
		// Handle file with sftp (autoUpload changes)
		// And you can change the handler whatever you need like rsync
		eventHandler()

		fileHandlerDone <- true
	}()

	<-watcherHandlerDone
	<-fileHandlerDone
}
