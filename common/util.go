package common

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set"
)

var (
	r = mapset.NewSet()
	I = mapset.NewSet()
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

	// image suffix
	I.Add("png")
	I.Add("jpg")
	I.Add("jpeg")
	I.Add("svg")

}

func ReadFiles(path string) string {
	b, e := ioutil.ReadFile(path)

	if e != nil {
		log.Println("read file error")
		return "error"
	}
	return string(b)
}

// TODO now only judge the filetype by filename suffix
// return (canRead,isImage,fileType). only file (not dir) should be fileName
func GetFileType(fileName string) (bool, bool, string) {
	readable, image := false, false
	ss := strings.Split(fileName, ".")
	// no specifal suffix, return file
	if len(ss) == 1 {
		return readable, image, "file"

	}
	suffix := ss[len(ss)-1]
	suffix = strings.ToLower(suffix)

	if r.Contains(suffix) {
		readable = true
	}
	if I.Contains(suffix) {
		image = true
	}

	return readable, image, suffix
}

func IsImage(fileName string) bool {
	_, image, _ := GetFileType(fileName)
	return image
}
func GenerateGID() int {
	GID++
	return GID
}
func RunTime(start time.Time) {
	elapsed := time.Since(start)
	fmt.Println("run time:", elapsed)
}
