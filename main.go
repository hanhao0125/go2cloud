package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
)

func watch() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add("./") //也可以监听文件夹
	if err != nil {
		log.Fatal(err)
	}
	<-done
}
func watchServer() {
	go watch()
	gin.ForceConsoleColor()
	// 初始化引擎
	router := gin.Default()

	router.LoadHTMLGlob("templates/*")
	router.POST("/upload", func(c *gin.Context) {
		name := c.PostForm("name")
		fmt.Println(name)
		file, header, err := c.Request.FormFile("upload")
		if err != nil {
			c.String(http.StatusBadRequest, "Bad request")
			return
		}
		filename := header.Filename

		fmt.Println(file, err, filename)

		out, err := os.Create(filename)
		if err != nil {
			log.Fatal(err)
		}
		defer out.Close()
		_, err = io.Copy(out, file)
		if err != nil {
			log.Fatal(err)
		}
		c.String(http.StatusCreated, "upload successful")
	})

	router.GET("/upload", func(c *gin.Context) {
		c.HTML(http.StatusOK, "upload.html", gin.H{"title": "this a test from fserver"})
	})

	// 注册一个路由和处理函数
	router.Any("/", WebRoot)
	// 绑定端口，然后启动应用
	router.Run(":9205")

}
func main() {
	// var wg sync.WaitGroup
	// wg.Add(1)
	// wg.Add(1)
	// go StartServer()
	// go TestClient()
	// wg.Wait()

}

/**
* 根请求处理函数
* 所有本次请求相关的方法都在 context 中，完美
* 输出响应 hello, world
 */
func WebRoot(context *gin.Context) {
	context.JSON(200, gin.H{"a": "abcs"})
	// context.String(http.StatusOK, "helloworld")
}
