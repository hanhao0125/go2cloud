// package main

// import (
// 	"bytes"
// 	"fmt"
// 	cn "go2cloud/common"
// 	"io"
// 	"io/ioutil"
// 	"log"
// 	"mime/multipart"
// 	"net/http"
// 	"os"
// 	"strings"
// 	"time"

// 	mapset "github.com/deckarep/golang-set"
// 	wuid "github.com/edwingeng/wuid/pgsql"
// )

// var (
// 	s1     = mapset.NewSet()
// 	s2     = mapset.NewSet()
// 	c1 int = 0
// 	c2 int = 0
// )

// // init the mysql table filenode. can be updated
// func Write2DB(path string, parentId int) {
// 	p := cn.MountedPath + path
// 	files, _ := ioutil.ReadDir(p)
// 	for _, file := range files {
// 		if file.IsDir() {
// 			pid := cn.InsertFileNode(file, path, parentId, "dir")
// 			Write2DB(path+file.Name()+"/", pid)
// 		} else {
// 			cn.InsertFileNode(file, path, parentId, "file")
// 			log.Print(file.Name())
// 		}
// 	}
// }

// func postFile(filePath string) error {
// 	//打开文件句柄操作
// 	file, err := os.Open(filePath)
// 	if err != nil {
// 		fmt.Println("error opening file")
// 		return err
// 	}
// 	defer file.Close()

// 	bodyBuf := &bytes.Buffer{}
// 	bodyWriter := multipart.NewWriter(bodyBuf)

// 	//关键的一步操作
// 	fileWriter, err := bodyWriter.CreateFormFile("file", filePath)
// 	if err != nil {
// 		fmt.Println("error writing to buffer")
// 		return err
// 	}

// 	// iocopy
// 	_, err = io.Copy(fileWriter, file)
// 	if err != nil {
// 		return err
// 	}

// 	contentType := bodyWriter.FormDataContentType()
// 	bodyWriter.Close()

// 	resp, err := http.Post("http://127.0.0.1:8888/upload", contentType, bodyBuf)
// 	if err != nil {
// 		return err
// 	}
// 	defer resp.Body.Close()
// 	resp_body, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		return err
// 	}
// 	fmt.Println(resp.Status)
// 	fmt.Println(string(resp_body))
// 	return nil
// }
// func listAll(path string) {
// 	files, _ := ioutil.ReadDir(path)
// 	cnt := 0
// 	var paths []string
// 	for _, fi := range files {
// 		if fi.IsDir() {
// 			// ignore dir
// 			//listAll(path + "/" + fi.Name())
// 			// println(path + "/" + fi.Name())
// 		} else {
// 			cnt++
// 			println(path + "/" + fi.Name())
// 			paths = append(paths, path+fi.Name())
// 			postFile(path + fi.Name())
// 		}
// 	}
// 	fmt.Print(cnt)

// }
// func deletePrefix(path, root string) string {
// 	return strings.Replace(path, root, "", -1)
// }
// func CountFileNum(path string, w string, root string) {
// 	files, _ := ioutil.ReadDir(path)
// 	for _, fi := range files {
// 		c1++
// 		p := deletePrefix(path+"/"+fi.Name(), root)
// 		if fi.IsDir() {
// 			if w == "1" {
// 				s1.Add(p)
// 			} else {
// 				s2.Add(p)
// 			}
// 			CountFileNum(path+"/"+fi.Name(), w, root)
// 		} else {
// 			if w == "1" {
// 				s1.Add(p)

// 			} else {
// 				s2.Add(p)
// 			}
// 		}
// 	}
// }
// func fuck() {
// 	p1 := "/Users/hanhao"
// 	CountFileNum(p1, "2", p1)
// 	// CountFileNum(p2, "2", p2)
// 	// fmt.Println(s1.Difference(s2))
// 	fmt.Println(c1)
// }
// func main() {
// 	defer cn.RunTime(time.Now())

// 	// Setup
// 	g := wuid.NewWUID("default", nil)
// 	g.LoadH24FromRedis("127.0.0.1:6379", "", "wuid")

// 	// Generate
// 	for i := 0; i < 10; i++ {
// 		fmt.Println(g.Next())
// 	}
// 	// fuck()
// 	// cn.T()
// 	cn.GetAllFilesFromDisk("/")
// 	// cn.Compare()
// 	fmt.Println(len(cn.Paths))
// 	// listAll("/Users/hanhao/Downloads/ILSVRC2012_img_test/")
// }
package main

import (
	cn "go2cloud/common"
	"time"
)

func main() {
	// Setup
	defer cn.RunTime(time.Now())
	cn.T()
	// c, err := redis.Dial("tcp", "127.0.0.1:6379")
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// c.Do("HMSET", "a", "z", "b")

}
