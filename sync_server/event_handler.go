package main

import (
	cn "go2cloud/common"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/pkg/sftp"
)

var (
	wg         sync.WaitGroup
	err        error
	sftpClient *sftp.Client
)

func eventHandler() {
	// Check if the mounted dir exist
	if _, err := os.Stat(MountedDir); err != nil {
		panic("Mounted dir dose not exist: " + MountedDir)
	}

	wg.Add(1)

	// Create Event handler
	go func() {
		for {
			select {
			// case createFileName := <-fileCreateEvent:
			// if isDir(createFileName) == 1 {
			// 	log.Println("update dir, ", createFileName)
			// 	UpdateDirHandler(createFileName)
			// } else {
			// 	log.Println("update file, ", createFileName)
			// 	UpdateFileHandler(createFileName)
			// }

			case writeFileName := <-fileWriteEvent:
				log.Print("write file: " + writeFileName)
				WriteHandler(writeFileName)

			case removeFileName := <-fileRemoveEvent:
				log.Print("remove file: " + removeFileName)
				RemoveHandler(removeFileName)

			case renameFileName := <-fileRenameEvent:
				log.Print("rename file: " + renameFileName)
				// TODO when have a rename operation, it will delete the whole dir and upload whole dir. It will be a big
				// TODO action when the dir contains very large files.
				RemoveHandler(renameFileName)

				// case chmodFileName := <-fileChmodEvent:

				// log.Print("chmod file" + chmodFileName)
			}
		}
	}()

	wg.Wait()
}

func UpdateFileHandler(path string) {
	// input is the abs path, first replace the mounted dir, then split the parent dir and real filename
	parentDir, parentId := findParent(path)

	fileInfo, err := os.Stat(path)
	if err != nil {
		log.Panic("err,", err)
	}
	canRead, fileType := cn.GetFileType(fileInfo.Name())

	if canRead {
		// readable file, send to nsq for search engine update index.
		cn.InsertFileNode(fileInfo, parentDir, parentId, fileType)
		// fmt.Println("message", id)
		// cn.PublishMessage(cn.IndexedTextTopic, strconv.Itoa(id))
	} else {
		cn.InsertFileNode(fileInfo, parentDir, parentId, "file")
	}
}

func findParent(path string) (string, int) {
	relativePath := strings.Replace(path, MountedDir, "", -1)
	ss := strings.Split(relativePath, "/")
	parentDir, _ := strings.Join(ss[:len(ss)-1], "/"), ss[len(ss)-1]
	if parentDir == "" {
		return "/", 0
	} else {
		parentId := 0
		parentNode, err := cn.FetchNodeByParentDir(parentDir)
		if err != nil {
			log.Panic("error when fetch node by parentDir,err=", err, "parent", parentDir)
		}
		parentId = parentNode.Id
		return parentDir + "/", parentId
	}
}
func WriteHandler(filePath string) {
	info, err := os.Stat(filePath)
	if err != nil {
		log.Println("error= ", err)
		return
	}
	relativePath := strings.Replace(filePath, MountedDir, "", -1)
	node, _ := cn.FetchNodeByFullPath(relativePath)
	// TODO may be other attr changed
	cn.UpdateNode(node, info.ModTime(), info.Size())
}

// only record the operation, update db.
func RemoveHandler(filePath string) {
	relativePath := strings.Replace(filePath, MountedDir, "", -1)
	node, err := cn.FetchNodeByFullPath(relativePath)
	if err == nil {
		cn.DeleteNodeById(node)
		if node.FileType == "dir" {
			cn.DeleteNodeByParentId(node.Id)
		}
	} else {
		log.Println("remove handler error,err= ", err)
	}
}
func RemoveFileHandler(fileName string) {
	relativePath := strings.Replace(fileName, MountedDir, "", -1)
	cn.DeleteNodeByFilePath(relativePath)
	log.Println("success remove file:", fileName)

}
func RemoveDirHandler(path string) {
	log.Println("in update dir path is ", path)
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Println(err)
	}
	// Recursive delete files and dirs
	for _, f := range files {
		// f.Name() only have name , path is not included
		if f.IsDir() {
			RemoveDirHandler(path + "/" + f.Name())
		} else {
			RemoveFileHandler(path + "/" + f.Name())
		}
	}
	// finally delete empty dir
	RemoveFileHandler(path)
}
