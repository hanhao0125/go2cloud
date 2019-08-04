package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
)

func postFile(filePath string) error {

	//打开文件句柄操作
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("error opening file")
		return err
	}
	defer file.Close()

	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	//关键的一步操作
	fileWriter, err := bodyWriter.CreateFormFile("file", filePath)
	if err != nil {
		fmt.Println("error writing to buffer")
		return err
	}

	// iocopy
	_, err = io.Copy(fileWriter, file)
	if err != nil {
		return err
	}

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	resp, err := http.Post("http://127.0.0.1:8888/upload", contentType, bodyBuf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	resp_body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println(resp.Status)
	fmt.Println(string(resp_body))
	return nil
}
func listAll(path string) {
	files, _ := ioutil.ReadDir(path)
	cnt := 0
	var paths []string
	for _, fi := range files {
		if fi.IsDir() {
			// ignore dir
			//listAll(path + "/" + fi.Name())
			// println(path + "/" + fi.Name())
		} else {
			cnt++
			println(path + "/" + fi.Name())
			paths = append(paths, path+fi.Name())
			postFile(path + fi.Name())
		}
	}
	fmt.Print(cnt)

}

// sample usage
func main() {
	listAll("/Users/hanhao/Downloads/ILSVRC2012_img_test/")
}
