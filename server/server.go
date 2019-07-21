package main

import (
	"fmt"
	"io"
	"net"
	"os"
)

func revFile(fileName string, conn net.Conn) {
	defer conn.Close()
	fs, err := os.Create(fileName)
	defer fs.Close()
	if err != nil {
		fmt.Println("os.Create err =", err)
		return
	}

	// 拿到数据
	buf := make([]byte, 1024*10)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("conn.Read err =", err)
			if err == io.EOF {
				fmt.Println("文件结束了", err)
			}
			return
		}
		if n == 0 {
			fmt.Println("文件结束了", err)
			return
		}
		fs.Write(buf[:n])
	}
}
func StartServer() {
	// 创建一个服务器
	Server, err := net.Listen("tcp", "localhost:8888")

	if err != nil {
		fmt.Println("net.Listen err =", err)
		return
	}
	fmt.Println("start server success, listing on localhost:8888")
	defer Server.Close()
	// 接受文件名
	for {
		conn, err := Server.Accept()
		defer conn.Close()
		if err != nil {
			fmt.Println("Server.Accept err =", err)
			return
		}
		buf := make([]byte, 1024)
		n, err1 := conn.Read(buf)
		if err1 != nil {
			fmt.Println("conn.Read err =", err1)
			return
		}
		// 拿到了文件的名字
		fileName := string(buf[:n])
		// 返回ok
		conn.Write([]byte("ok"))
		// 接收文件,
		revFile(fileName, conn)

	}

}
func main() {
	StartServer()
}
