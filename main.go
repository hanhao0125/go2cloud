package main

import (
	"fmt"
	cn "go2cloud/common"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

func init() {
	log.SetPrefix("[ server ] ")
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.LUTC)
	// m.InitM()
	// cn.ScanRootPath("/", &m)
}

func StartHttpServices() {

	gin.ForceConsoleColor()
	router := gin.Default()

	router.LoadHTMLGlob("./app/templates/*")
	router.Static("/static", cn.MountedPath)
	router.Static("/logo", "./app/static")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "base.html", gin.H{"title": "Go2Cloud"})
	})
	router.GET("/index", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{"title": "this a test from fserver"})
	})

	router.GET("/search", func(c *gin.Context) {
		c.HTML(http.StatusOK, "search.html", gin.H{})
	})

	router.GET("/files", FetchFilePaths)

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

	router.GET("/ace", func(c *gin.Context) {
		path := cn.MountedPath + c.Query("src")
		c.HTML(http.StatusOK, "iframe.html", gin.H{"title": "this a test from fserver", "text": cn.ReadFiles(path)})

	})

	router.GET("/edit", func(c *gin.Context) {
		pid := c.Query("pid")
		pidint, _ := strconv.Atoi(pid)
		n := cn.FetchFileNodeById(pidint)
		finalPath := n.ParentDir + n.Path
		if n.FileType == "pdf" {
			log.Println("pdf coming")
			c.HTML(http.StatusOK, "edit.html", gin.H{"path": n.ParentDir + n.Path, "src": "static" + n.ParentDir + n.Path, "type": "pdf"})
		} else {
			// text := utils.ReadFiles(config.MountedPath + path)
			c.HTML(http.StatusOK, "edit.html", gin.H{"path": finalPath, "src": finalPath, "type": "text"})
		}
	})

	router.GET("/upload", func(c *gin.Context) {
		c.HTML(http.StatusOK, "upload.html", gin.H{"title": "this a test from fserver"})
	})

	router.POST("/upload1", func(c *gin.Context) {
		file, _ := c.FormFile("file")
		timestamp := strconv.FormatInt(time.Now().UnixNano(), 10)
		ss := strings.Split(file.Filename, "/")
		fileName := timestamp + ss[len(ss)-1]
		savePath := cn.MountedPath + fileName
		err := c.SaveUploadedFile(file, savePath)
		if err != nil {
			log.Println(err)
		}
	})
	router.Run(cn.WebServicePort)
}
func FetchFilePaths(c *gin.Context) {
	p := c.DefaultQuery("p", "-1")
	pint, _ := strconv.Atoi(p)
	// nodes, _ := cn.FetchNodesByParentId(pint)
	nodes, _ := cn.FetchNodesByParentId(pint)
	c.JSON(http.StatusOK, gin.H{"f": nodes})
}

func main() {
	// fileInfo, _ := os.Stat(cn.MountedPath + "/" + "abc/vuex/README.md")
	// cn.Insert(fileInfo, "/Users/hanhao/server/abc/vuex/README.md")
	// cn.Test()
	StartHttpServices()

	// cn.TimelyTask()
	// cn.DBScanRootPath("/")
}
