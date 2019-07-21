package main

import (
	"fmt"
	"net"
	"os"
)

func sendFile(path string, conn net.Conn) {
	defer conn.Close()
	fs, err := os.Open(path)
	defer fs.Close()
	if err != nil {
		fmt.Println("os.Open err = ", err)
		return
	}
	buf := make([]byte, 1024*10)
	for {
		//  打开之后读取文件
		n, err1 := fs.Read(buf)
		if err1 != nil {
			fmt.Println("fs.Open err = ", err1)
			return
		}

		//  发送文件
		conn.Write(buf[:n])
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
func main() {
	TestClient()
}
