package db

import (
	"fmt"
	"go2cloud/config"
	"go2cloud/models"
	"go2cloud/utils"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/garyburd/redigo/redis"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var db *gorm.DB

func init() {
	var err error
	db, err = gorm.Open("mysql", config.MysqlPath)
	if err != nil {
		panic(err)
	}
	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(100)
}

func FetchReadableFileNode() []models.ReadableFileNode {
	nodes := []models.Node{}
	db.Where("file_type = ?", "text").Find(&nodes)

	rnds := []models.ReadableFileNode{}

	for _, n := range nodes {
		readPath := config.MountedPath + n.ParentDir + n.Path
		content := utils.ReadFiles(readPath)
		rnds = append(rnds, models.ReadableFileNode{Id: n.Id, Path: n.ParentDir + n.Path, Content: content, FileType: n.FileType})
	}
	log.Println("process file and success file", len(nodes), len(rnds))
	return rnds
}
func FetchReadableFileNodeById(id int) models.ReadableFileNode {
	n := FetchFileNodeById(id)
	readPath := config.MountedPath + n.ParentDir + n.Path
	content := utils.ReadFiles(readPath)
	readableNode := models.ReadableFileNode{Id: n.Id, Path: n.ParentDir + n.Path, Content: content, FileType: n.FileType}
	return readableNode
}
func GetPaths(c *gin.Context) {
	// default list *MountedPath* dir
	p := c.DefaultQuery("p", "")
	r, _ := redis.Dial("tcp", config.RedisAddress)
	defer r.Close()
	r.Do("AUTH", config.RedisPassword)
	exist, _ := redis.Bool(r.Do("EXISTS", p+":types"))

	folders := make([]models.Node, 0)
	if exist {
		fmt.Println("cached")
		// types, _ := redis.Values(r.Do("lrange", p+":types", "0", "-1"))
		// paths, _ := redis.Values(r.Do("lrange", p+":paths", "0", "-1"))
		// pds, _ := redis.Values(r.Do("lrange", p+":pds", "0", "-1"))
		// for index, _ := range types {
		// 	folders = append(folders, models.Node{Type: string(types[index].([]byte)), Path: string(paths[index].([]byte)), ParentDir: string(pds[index].([]byte))})
		// }
	} else {
		// folders = GetAllFoldersAndFiles(p, folders)
		// for _, k := range folders {
		// 	r.Do("lpush", p+":types", k.Type)
		// 	r.Do("lpush", p+":paths", k.Path)
		// 	r.Do("lpush", p+":pds", k.ParentDir)
		// }
	}
	c.JSON(http.StatusOK, gin.H{"f": folders})
}

// init the mysql table filenode. can be updated
func Write2DB(path string, parentId int) {
	p := config.MountedPath + path
	files, _ := ioutil.ReadDir(p)
	for _, file := range files {
		if file.IsDir() {
			// pid := InsertFileNode(file, path, parentId, "dir")
			Write2DB(path+file.Name()+"/", 0)
		} else {
			// InsertFileNode(file, path, parentId, "file")
			log.Print(file.Name())
		}
	}
}

func UpdateImageDB() {
	images := []models.Image{}
	nodes := models.Node{}
	// db, err := gorm.Open("mysql", config.MysqlPath)
	// if err != nil {
	// 	fmt.Println("connection err:", err)
	// }
	db.Find(&images)
	for _, img := range images {
		is := strings.Split(img.Path, "/")
		node, err := FindNodeIdByImageName(is[len(is)-1])
		if err != nil {
			continue
		}
		db.Model(&img).UpdateColumns(models.Image{FileId: node.Id, Path: node.ParentDir + node.Path})
	}
	// defer db.Close()
	db.Find(&nodes)

}

func FindNodeIdByImageName(imgName string) (models.Node, error) {
	node := models.Node{}
	// db, _ := gorm.Open("mysql", config.MysqlPath)
	// defer db.Close()
	q := db.Where("path = ?", imgName).First(&node)
	if q.Error != nil {
		fmt.Println("error! no sucn image,", imgName)
		return node, q.Error
		// panic(q.Error)
	}
	return node, nil
}

func FetchFileNodeById(id int) models.Node {
	node := models.Node{}
	q := db.First(&node, id)
	if q.Error != nil {
		log.Printf("error,%v", q.Error)
	}
	return node
}

func FetchFileNodesByIds(ids []int) []models.Node {
	ret := []models.Node{}
	for _, v := range ids {
		ret = append(ret, FetchFileNodeById(v))
	}
	return ret
}
