package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"time"
)

func init() {
	log.SetPrefix("[CLIENT]")
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.LUTC)
}

func singleFile(path string) {
	info, err := os.Stat(path)
	if err != nil {
		fmt.Println("os.Stat err = ", err)
		return
	}
	// 发送文件名
	conn, err1 := net.Dial("tcp", "localhost:8888")
	defer conn.Close()
	if err1 != nil {
		fmt.Println("net.Dial err = ", err1)
		return
	}
	conn.Write([]byte(info.Name()))
	// 接受到是不是ok
	buf := make([]byte, 1024)
	n, err2 := conn.Read(buf)
	if err2 != nil {
		fmt.Println("conn.Read err = ", err2)
		return
	}
	if "ok" == string(buf[:n]) {
		fmt.Println("成功")
		sendFile(path, conn)
	}
	// 如果是ok,那么开启一个连接,发送文件
}
func testMultiFiles() {
	defer timeCost(time.Now())
	listFile("/Users/hanhao/Documents/Master-Paper")

}
func timeCost(start time.Time) {
	tc := time.Since(start)
	fmt.Printf("time cost = %v\n", tc)
}
func listFile(folder string) {
	//specify the current dir
	files, _ := ioutil.ReadDir(folder)
	for _, file := range files {
		if file.IsDir() {
			listFile(folder + "/" + file.Name())
		} else {
			// log.Println(folder + "/" + file.Name())
			singleFile(folder + "/" + file.Name())
		}
	}

}
func TestClient() {
	for {
		fmt.Println("请输入一个全路径的文件,比如,D:\\a.jpg")
		//  获取命令行参数
		var path string
		fmt.Scan(&path)
		// 获取文件名,
		info, err := os.Stat(path)
		if err != nil {
			fmt.Println("os.Stat err = ", err)
			return
		}
		// 发送文件名
		conn, err1 := net.Dial("tcp", "localhost:8888")
		defer conn.Close()
		if err1 != nil {
			fmt.Println("net.Dial err = ", err1)
			return
		}
		conn.Write([]byte(info.Name()))
		// 接受到是不是ok
		buf := make([]byte, 1024)
		n, err2 := conn.Read(buf)
		if err2 != nil {
			fmt.Println("conn.Read err = ", err2)
			return
		}
		if "ok" == string(buf[:n]) {
			fmt.Println("成功")
			sendFile(path, conn)
		}
		// 如果是ok,那么开启一个连接,发送文件
	}
}

// func main() {
// 	StartWatchServer("/Users/hanhao/Documents/Master-Paper/")
// }
