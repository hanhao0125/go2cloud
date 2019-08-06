package main

import (
	"fmt"
	cn "go2cloud/common"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

type Watch struct {
	watch *fsnotify.Watcher
}

// handler jobs done
var eventDone = make(chan bool)

// Watch a directory
func (w *Watch) watchDir(dir string) {

	// Walk all directory
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		// Just watch directory(all child can be watched)
		if info.IsDir() {
			path, err := filepath.Abs(path)
			if err != nil {
				log.Fatal(err)
			}
			err = w.watch.Add(path)
			if err != nil {
				log.Fatal(err)
			}
		}
		return nil
	})

	log.Print("Watching: ", dir)

	// Handle the watch events
	go eventsHandler(w)

	// Await
	<-eventDone
}
func fileChecker(filename string) bool {

	// for _, ignoreFile := range config.IgnoreFiles {
	// 	if strings.Contains(filename, string(ignoreFile)) {
	// 		return true
	// 	}
	// }

	return false

}

// Handle the watch events
func eventsHandler(w *Watch) {
	for {
		select {
		case ev := <-w.watch.Events:
			{
				// create
				if ev.Op&fsnotify.Create == fsnotify.Create {
					fi, err := os.Stat(ev.Name)
					log.Println("update dir, ", ev.Name)
					if err != nil {
						log.Panic("error when meets create event,err= ", err)
						panic("err")
					}
					if fi.IsDir() {
						UpdateDirHandler(ev.Name, w)
					} else {
						UpdateFileHandler(ev.Name)
					}

					// if err == nil && fi.IsDir() {
					// 	w.watch.Add(ev.Name)
					// }
					// fileCreateEvent <- ev.Name
				}

				// write
				if ev.Op&fsnotify.Write == fsnotify.Write {
					fileWriteEvent <- ev.Name
				}

				// delete event
				if ev.Op&fsnotify.Remove == fsnotify.Remove {
					w.watch.Remove(ev.Name)
					fileRemoveEvent <- ev.Name
				}

				// rename
				if ev.Op&fsnotify.Rename == fsnotify.Rename {
					w.watch.Remove(ev.Name)
					fileRenameEvent <- ev.Name
				}
				// chmod
				// if ev.Op&fsnotify.Chmod == fsnotify.Chmod {
				// 	fileChmodEvent <- ev.Name
				// }
			}
		case err := <-w.watch.Errors:
			{
				log.Fatal(err)
				eventDone <- true
				return
			}
		}
	}

	// eventDone <- true
}

func UpdateDirHandler(path string, w *Watch) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		fmt.Println(err)
	}
	parentDir, parentId := findParent(path)
	info, err := os.Stat(path)
	if err != nil {
		fmt.Println("err", err)
	}
	cn.InsertFileNode(info, parentDir, parentId, "dir")
	for _, f := range files {
		// f.Name() only have name , path is not included
		if f.IsDir() {
			UpdateDirHandler(path+"/"+f.Name(), w)
		} else {
			UpdateFileHandler(path + "/" + f.Name())
		}
	}
	w.watch.Add(path)
}
