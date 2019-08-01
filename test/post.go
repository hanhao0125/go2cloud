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
	// if err != nil {
	// 	fmt.Println("error writing to buffer")
	// 	return err
	// }

	// iocopy
	_, err = io.Copy(fileWriter, file)
	if err != nil {
		return err
	}

	////设置其他参数
	//params := map[string]string{
	//	"user": "test",
	//	"password": "123456",
	//}
	//
	////这种设置值得仿佛 和下面再从新创建一个的一样
	//for key, val := range params {
	//	_ = bodyWriter.WriteField(key, val)
	//}

	//和上面那种效果一样
	//建立第二个fields
	// if fileWriter, err = bodyWriter.CreateFormField("user"); err != nil {
	// 	fmt.Println(err, "----------4--------------")
	// }
	// if _, err = fileWriter.Write([]byte("test")); err != nil {
	// 	fmt.Println(err, "----------5--------------")
	// }
	// //建立第三个fieds
	// if fileWriter, err = bodyWriter.CreateFormField("password"); err != nil {
	// 	fmt.Println(err, "----------4--------------")
	// }
	// if _, err = fileWriter.Write([]byte("123456")); err != nil {
	// 	fmt.Println(err, "----------5--------------")
	// }

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
