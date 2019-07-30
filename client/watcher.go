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
				log.Println("event:", event)
				//write
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)
				}
				//create
				if event.Op&fsnotify.Create == fsnotify.Create {
					log.Println("create file:", event.Name)
					UploadFile(event.Name)
				}
				//remove
				if event.Op&fsnotify.Remove == fsnotify.Remove {
				}
				//rename
				if event.Op&fsnotify.Rename == fsnotify.Rename {
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
