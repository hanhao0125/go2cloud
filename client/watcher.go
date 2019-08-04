package main

import (
	"log"

	"github.com/fsnotify/fsnotify"
)

func init() {
	log.SetPrefix("[WATCHER]")
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.LUTC)
}

func watch(rootPath string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("sync client start!")
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				// These are the generalized file operations that can trigger a notification.
				// copy operation: is create. but only copy the folder, no inside file
				// log.Println("event:", event)

				/*
					create path
					rename path：means delete operation, since it will first create file and then rename
					remove path: 3 steps: 1. create
					upload path
				*/

				//write
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)
				}
				//create
				if event.Op&fsnotify.Create == fsnotify.Create {
					log.Println("create file:", event.Name)
					// UploadFile(event.Name)
				}
				//remove

				if event.Op&fsnotify.Remove == fsnotify.Remove {
					log.Println("remove file", event.Name)

				}
				//rename
				if event.Op&fsnotify.Rename == fsnotify.Rename {
					log.Println("rename file", event.Name)
				}
				if event.Op&fsnotify.Chmod == fsnotify.Chmod {
					log.Println("chmod", event.Name)
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(rootPath)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}
func StartWatchServer(rootPath string) {
	watch(rootPath)
}
func main() {
	StartWatchServer("/Users/hanhao/Downloads/")
}
