package main

import (
	"io/ioutil"
	"testing"
)

var Path []string
var ScanPath = "/Users/hanhao/Documents"
var MountedPath = "/Users/hanhao/Documents"

func ScanFilePaths(path string) {
	absPath := MountedPath + path
	files, _ := ioutil.ReadDir(absPath)
	for _, info := range files {
		full_path := path + info.Name()
		Path = append(Path, full_path)
		if info.IsDir() {
			ScanFilePaths(full_path + "/")
		} else {
		}
	}
}

func TestInitServer(t *testing.T) {
	IndexFromDB()
}
