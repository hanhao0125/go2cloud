package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"go2cloud/config"
	"go2cloud/models"
	search "go2cloud/search_server"
	"go2cloud/utils"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

func init() {
	log.SetPrefix("[ server ] ")
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.LUTC)
}

var ()

func FetchNodeById(pid int) models.Node {
	db, err := gorm.Open("mysql", config.MysqlPath)
	if err != nil {
		log.Println("connected error, ", err)
	}
	defer db.Close()
	node := models.Node{}
	db.First(&node, pid)

	return node
}

func StartHttpServices() {
	gin.ForceConsoleColor()
	router := gin.Default()

	router.LoadHTMLGlob("templates/*")
	// router.Static("/static", config.MountedPath)
	router.Static("/static", config.MountedPath)
	router.Static("/logo", "./static")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "base.html", gin.H{"title": "Go2Cloud"})
	})
	router.GET("/index", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{"title": "this a test from fserver"})
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
		path := config.MountedPath + c.Query("src")
		c.HTML(http.StatusOK, "iframe.html", gin.H{"title": "this a test from fserver", "text": utils.ReadFiles(path)})

	})

	router.GET("/edit", func(c *gin.Context) {
		pid := c.Query("pid")
		pidint, _ := strconv.Atoi(pid)
		n := FetchNodeById(pidint)
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
	router.Run(":9205")
}
func FetchFilePaths(c *gin.Context) {
	p := c.DefaultQuery("p", "0")
	pint, _ := strconv.Atoi(p)
	c.JSON(http.StatusOK, gin.H{"f": FetchFileNode(pint)})
}

func InsertFileNode(file os.FileInfo, parentDir string, parentId int, fileType string) int {
	node := models.Node{FileType: fileType, Path: file.Name(), ParentDir: parentDir, ParentId: parentId, ModTime: file.ModTime(), FileSize: file.Size(), Share: config.ShareSingal}
	db, err := gorm.Open("mysql", config.MysqlPath)
	if err != nil {
		fmt.Println("connection err:", err)
	}
	defer db.Close()
	db.Create(&node)
	return node.Id
}

func FetchFileNode(p int) []models.Node {
	db, err := gorm.Open("mysql", config.MysqlPath)
	if err != nil {
		log.Println("connected error, ", err)
	}
	defer db.Close()
	nodes := []models.Node{}
	db.Limit(100).Where("parent_id=?", p).Find(&nodes)
	return nodes
}

// init the mysql table filenode. can be updated
func Write2DB(path string, parentId int) {
	p := config.MountedPath + path
	files, _ := ioutil.ReadDir(p)
	for _, file := range files {
		if file.IsDir() {
			pid := InsertFileNode(file, path, parentId, "dir")
			Write2DB(path+file.Name()+"/", pid)
		} else {
			InsertFileNode(file, path, parentId, "file")
			log.Print(file.Name())
		}
	}
}

func main() {
	// StartHttpServices()
	// search.IndexFromDB()
	// search.Search("网络")
	search.StartSearchHttpService()
	// nsq.Fuck()
	// nsq.RecieveAndProcess()
	// nsq.PublishMessage(config.IndexedTextTopic, "378987")
	// db.UpdateImageDB()
	// db.FindNodeIdByImageName("1564669838915749000ILSVRC2012_test_00000001.JPEG")
}
