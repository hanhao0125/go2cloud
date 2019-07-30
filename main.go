package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func GinServer() {
	// init search engine
	Index("/Users/hanhao/Documents/")

	gin.ForceConsoleColor()
	// 初始化引擎
	router := gin.Default()

	router.LoadHTMLGlob("templates/*")
	router.GET("/index", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{"title": "this a test from fserver"})
	})
	router.GET("/search", func(c *gin.Context) {
		query := c.Query("query")
		res := Search(query)
		c.JSON(200, gin.H{
			"docs": res,
		})
	})
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

	router.Run(":9205")

}
func main() {
	StartWatchServer("./")
	// GinServer()

}
