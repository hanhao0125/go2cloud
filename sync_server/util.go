package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

func isDir(fileName string) int {
	fileInfo, err := os.Stat(fileName)

	if err != nil {
		log.Println("file dose not exist: " + fileName)
		return -1
	}

	if fileInfo.IsDir() {
		return 1
	} else {
		return 0
	}
}

// Get the change part of filePath
func getChangedFileName(fileName string) string {
	fileName = strings.Replace(fileName, MountedDir, "", -1)
	return filepath.ToSlash(fileName)
}
