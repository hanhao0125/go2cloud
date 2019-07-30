package main

import (
	"io/ioutil"
	"log"
	"net"
	"os"
)

var (
	ServerAddress string = "localhost"
	ServerPort    string = "8888"
	ServerPath    string = "/Users/hanhao/netserver/"
)

//文件的抽象，需要为不同的文件实现不同的 read 方法：pdf, doc, txt, etc.
type File struct {
	FileType string
	Path     string
	Content  string
}

/*
用接口屏蔽file的底层实现，传给搜索引擎接口，接口需要实现下列方法：
1. 返回文本内容（建立索引）
2. 返回文件路径（位置）
搜索引擎只需要上述两个数据即可完成相关功能，无需关系底层文件是什么类型的。
*/
type FileData interface {
	ReadContent() string
	Path() string
}

func (f *File) ReadContent() string {
	file, err := ioutil.ReadFile(f.Path)
	if err != nil {
		panic(err)
	}
	f.Content = string(file)
	return f.Content
}

func init() {
	log.SetPrefix("[SYNCSERVER]")
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.LUTC)
}
func testMultiFiles() {
	listFile("/Users/hanhao/Documents/Master-Paper")

}
func listFile(folder string) {
	//specify the current dir
	files, _ := ioutil.ReadDir(folder)
	for _, file := range files {
		if file.IsDir() {
			listFile(folder + "/" + file.Name())
		} else {
			log.Println(folder + "/" + file.Name())
		}
	}

}
func revFile(fileName string, conn net.Conn) {
	defer conn.Close()
	p := ServerPath + fileName
	fs, err := os.Create(p)
	defer fs.Close()

	if err != nil {
		log.Println("os.Create err =", err)
		return
	}

	// 拿到数据
	buf := make([]byte, 1024*10)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Println("conn.Read err =", err)

			return
		}
		if n == 0 {
			log.Println("接受文件成功：" + fileName)
			return
		}
		fs.Write(buf[:n])
	}
}
func StartSyncServer() {
	// 创建一个服务器
	Server, err := net.Listen("tcp", ServerAddress+":"+ServerPort)

	if err != nil {
		log.Println("net.Listen err =", err)
		return
	}
	log.Println("start server success, listing on " + ServerAddress + ":" + ServerPort)
	defer Server.Close()
	// 接受文件名
	for {
		conn, err := Server.Accept()
		defer conn.Close()
		if err != nil {
			log.Println("Server.Accept err =", err)
			return
		}
		buf := make([]byte, 1024)
		n, err1 := conn.Read(buf)
		if err1 != nil {
			log.Println("conn.Read err =", err1)
			return
		}
		// 拿到了文件的名字
		fileName := string(buf[:n])
		// 返回ok
		conn.Write([]byte("ok"))
		// 接收文件,
		go revFile(fileName, conn)
	}

}
func main() {
	StartSyncServer()
	// testMultiFiles()
}
