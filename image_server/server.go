package main

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var (
	RootPath  string = "/Users/hanhao/netserver/img/"
	MysqlPath string = "root:root@/cloud?charset=utf8&parseTime=True&loc=Local"
)

func ImageServer() {
	gin.ForceConsoleColor()
	router := gin.Default()

	router.LoadHTMLGlob("../templates/*")

	router.Use(cors.Default())
	router.Static("/image", "/Users/hanhao/netserver/img")

	router.GET("/image", func(c *gin.Context) {
		c.HTML(http.StatusOK, "image.html", gin.H{"p": ""})
	})

	router.GET("/i", func(c *gin.Context) {
		paths, tags := GetImagePath(100)
		c.JSON(http.StatusOK, gin.H{"paths": paths, "tags": tags})
	})

	router.GET("/is", func(c *gin.Context) {
		keyword := c.Query("query")
		paths, tags := GetImageByTag(keyword)
		c.JSON(http.StatusOK, gin.H{"paths": paths, "tags": tags})
	})

	router.GET("/staticis", func(c *gin.Context) {
		tags, cnt := GroupTags()
		c.JSON(http.StatusOK, gin.H{"tags": tags, "cnt": cnt})
	})

	//这个时候就得写数据库，保证用户上传的文件不会丢失。后面消息队列打 tag 能够容忍失败。
	router.POST("/upload", func(c *gin.Context) {
		file, _ := c.FormFile("file")
		timestamp := strconv.FormatInt(time.Now().UnixNano(), 10)
		ss := strings.Split(file.Filename, "/")
		fileName := timestamp + ss[len(ss)-1]
		savePath := RootPath + fileName
		err := c.SaveUploadedFile(file, savePath)
		if err != nil {
			fmt.Println("save errr=", err)
		}

		if ret := InsertImage(savePath); ret != -1 {
			message := strconv.Itoa(ret) + "|" + savePath
			fmt.Println(message)
			PublishMessage("tag", message)
			c.JSON(http.StatusOK, gin.H{"success": true})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		}
	})

	router.Run(":8888")
}

func InsertImage(path string) int {
	image := Image{Path: path, Upath: path, Uploaddate: time.Now()}
	db, err := gorm.Open("mysql", "root:root@/cloud?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		fmt.Println("connection err:", err)
		return -1
	}
	defer db.Close()
	db.Create(&image)
	return image.Id
}
func updateTag() {
	db, err := gorm.Open("mysql", MysqlPath)
	if err != nil {
		fmt.Println("connection err:", err)
	}
	defer db.Close()
	images := []Image{}
	db.Find(&images)
	for _, k := range images {
		db.Model(&k).Update("tag", strings.Split(k.Tag, "|")[0])
	}

}

func GetImageByTag(tag string) ([]string, []string) {
	db, err := gorm.Open("mysql", "root:root@/cloud?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		fmt.Println("connection err:", err)
	}
	defer db.Close()
	image := []Image{}
	db.Where("tag like ?", "%"+tag+"%").Find(&image)
	var paths []string
	var tags []string
	for _, k := range image {
		p := strings.Split(k.Path, "/")

		paths = append(paths, "image/"+p[len(p)-1])
		tags = append(tags, k.Tag)
	}
	return paths, tags
}

func GetImagePath(limit int) ([]string, []string) {
	image := []Image{}
	db, err := gorm.Open("mysql", "root:root@/cloud?charset=utf8&parseTime=True&loc=Local")
	defer db.Close()
	if err != nil {
		fmt.Println("connect to mysql error", err)
	}
	db.Order("uploaddate desc").Limit(limit).Find(&image)
	var paths []string
	var tags []string
	for _, k := range image {
		p := strings.Split(k.Path, "/")

		paths = append(paths, "image/"+p[len(p)-1])
		tags = append(tags, k.Tag)
	}
	return paths, tags
}

func GroupTags() ([]string, []int) {
	db, err := gorm.Open("mysql", "root:root@/cloud?charset=utf8&parseTime=True&loc=Local")
	defer db.Close()
	rows, err := db.Table("image").Select("tag as t, count(*) as cnt").Group("tag").Rows()
	if err != nil {
		fmt.Println("err=", err)
	}
	tagCount := make(map[string]int)
	for rows.Next() {
		tag := ""
		cnt := 0
		rows.Scan(&tag, &cnt)
		tagCount[tag] = cnt
	}
	type kv struct {
		Key   string
		Value int
	}
	var sortedTagCount []kv
	for k, v := range tagCount {
		sortedTagCount = append(sortedTagCount, kv{k, v})
	}
	sort.Slice(sortedTagCount, func(i, j int) bool {
		return sortedTagCount[i].Value > sortedTagCount[j].Value
	})
	var tags []string
	var cnt []int
	for _, kv := range sortedTagCount {
		if kv.Value < 100 {
			continue
		}
		tags = append(tags, kv.Key)
		cnt = append(cnt, kv.Value)
	}
	fmt.Print(tags, cnt)
	return tags, cnt
}

func main() {
	// GroupTags()
	ImageServer()
}
