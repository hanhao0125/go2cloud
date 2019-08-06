package common

import (
	"fmt"
	"io/ioutil"
	"strings"

	mapset "github.com/deckarep/golang-set"
)

var (
	r = mapset.NewSet()
)

func init() {
	// backend language
	r.Add("go")
	r.Add("py")
	r.Add("java")
	r.Add("c")
	r.Add("cpp")
	r.Add("h")

	// frontend language
	r.Add("js")
	r.Add("ts")
	r.Add("html")
	r.Add("css")

	// txt
	r.Add("txt")

	r.Add("json")
	r.Add("xml")
}

func ReadFiles(path string) string {
	b, e := ioutil.ReadFile(path)

	if e != nil {
		fmt.Println("read file error")
		return "error"
	}
	return string(b)
}

// TODO now only judge the filetype by filename suffix
// return (canRead,fileType). only file (not dir) should be fileName
func GetFileType(fileName string) (bool, string) {
	ss := strings.Split(fileName, ".")
	suffix := ss[len(ss)-1]
	if len(suffix) == 0 {
		return false, "file"
	}
	if r.Contains(suffix) {
		return true, suffix
	}
	return false, "file"
}
func GenerateGID() int {
	GID++
	return GID
}
