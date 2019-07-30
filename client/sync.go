package main

import (
	"fmt"
	"log"
	"net"
	"os"
)

var (
	ServerAddress string = "localhost"
	ServerPort    string = "8888"
)

func init() {
	log.SetPrefix("[SYNC]")
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.LUTC)
}

func UploadFile(path string) {
	// 获取文件名,
	info, err := os.Stat(path)
	if err != nil {
		log.Println("os.Stat err = ", err)
		return
	}
	// 发送文件名
	conn, err1 := net.Dial("tcp", ServerAddress+":"+ServerPort)
	defer conn.Close()
	if err1 != nil {
		log.Println("net.Dial err = ", err1)
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
		log.Println("start upload file: " + path)
		sendFile(path, conn)
		log.Println("success upload file: " + path)
	}
}
func sendFile(path string, conn net.Conn) {
	defer conn.Close()
	fs, err := os.Open(path)
	defer fs.Close()
	if err != nil {
		log.Println("open file err = ", err)
		return
	}
	buf := make([]byte, 1024*10)
	for {
		//  打开之后读取文件
		n, err1 := fs.Read(buf)
		if err1 != nil {
			fmt.Println("finished write")
			return
		}
		//  发送文件
		conn.Write(buf[:n])
	}
}
