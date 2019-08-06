package common

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var db *gorm.DB

func init() {
	var err error
	db, err = gorm.Open("mysql", MysqlPath)

	if err != nil {
		panic(err)
	}
	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(100)
}

func FetchReadableFileNode() []ReadableFileNode {
	nodes := []Node{}
	db.Where("file_type = ?", "text").Find(&nodes)

	rnds := []ReadableFileNode{}

	for _, n := range nodes {
		readPath := MountedPath + n.ParentDir + n.Path
		content := ReadFiles(readPath)
		rnds = append(rnds, ReadableFileNode{Id: n.Id, Path: n.ParentDir + n.Path, Content: content, FileType: n.FileType})
	}
	log.Println("process file and success file", len(nodes), len(rnds))
	return rnds
}
func FetchReadableFileNodeById(id int) ReadableFileNode {
	n := FetchFileNodeById(id)
	readPath := MountedPath + n.ParentDir + n.Path
	log.Println(readPath)
	content := ReadFiles(readPath)
	readableNode := ReadableFileNode{Id: n.Id, Path: n.ParentDir + n.Path, Content: content, FileType: n.FileType}
	return readableNode
}
func GetPaths(c *gin.Context) {
	// default list *MountedPath* dir
	p := c.DefaultQuery("p", "")
	r, _ := redis.Dial("tcp", RedisAddress)
	defer r.Close()
	r.Do("AUTH", RedisPassword)
	exist, _ := redis.Bool(r.Do("EXISTS", p+":types"))

	folders := make([]Node, 0)
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
	p := MountedPath + path
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
	images := []Image{}
	nodes := Node{}
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
		db.Model(&img).UpdateColumns(Image{FileId: node.Id, Path: node.ParentDir + node.Path})
	}
	// defer db.Close()
	db.Find(&nodes)

}

func FindNodeIdByImageName(imgName string) (Node, error) {
	node := Node{}
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

func FetchFileNodeById(id int) Node {
	node := Node{}
	q := db.First(&node, id)
	if q.Error != nil {
		log.Printf("error,%v", q.Error)
	}
	return node
}

func FetchFileNodesByIds(ids []int) []Node {
	ret := []Node{}
	for _, v := range ids {
		ret = append(ret, FetchFileNodeById(v))
	}
	return ret
}
func InsertFileNode(file os.FileInfo, parentDir string, parentId int, fileType string) int {

	node := Node{FileType: fileType, Path: file.Name(), ParentDir: parentDir, ParentId: parentId,
		ModTime: file.ModTime(), FileSize: file.Size(), Share: ShareSingal, FullPath: parentDir + file.Name()}
	q := db.Create(&node)
	if q.Error != nil {
		log.Panic("error when insert file node,err= ", q.Error)
	}
	return node.Id
}

// init the mysql table filenode. can be updated
func WatcherWrite2DB(path string, parentId int) {
	p := MountedPath + path
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

func FetchNodeByParentDir(parentDir string) (Node, error) {
	node := Node{}
	q := db.Where("full_path = ? and file_type = ?", parentDir, "dir").First(&node)
	if q.Error != nil {
		log.Panic(parentDir, q.Error)
		return node, q.Error
		// panic(q.Error)
	}
	return node, nil
}

func FetchNodeByFullPath(fullPath string) (Node, error) {
	node := Node{}
	q := db.Where("full_path = ?", fullPath).First(&node)
	if q.Error != nil {
		log.Println(fullPath, q.Error)
		return node, q.Error
	}
	return node, nil
}
func FetchNodesByParentId(pid int) ([]Node, error) {
	nodes := []Node{}
	q := db.Where("parent_id = ?", pid).Find(&nodes)
	if q.Error != nil {
		log.Println(pid, q.Error)
		return nodes, q.Error
	}
	return nodes, nil
}
func DeleteNodeByFilePath(filePath string) {
	q := db.Where("full_path = ?", filePath).Delete(Node{})
	if q.Error != nil {
		log.Println("delete error, err=", q.Error, "\tfilePath:", filePath)
	}

}
func DeleteNodeById(node Node) {
	q := db.Delete(&node)
	if q.Error != nil {
		log.Println("delete error, err=", q.Error, "\tfilePath:", node)
	}

}

func DeleteNodeByParentId(parentId int) {
	nodes := []Node{}
	q := db.Where("parent_id = ?", parentId).Find(&nodes)

	if q.Error != nil {
		log.Println("delete error,error=", q.Error)
		return
	}
	for _, n := range nodes {
		if n.FileType == "dir" {
			DeleteNodeByParentId(n.Id)
		}
		db.Delete(&n)
	}
}
func UpdateNode(oldNode Node, modTime time.Time, size int64) {
	db.Model(&oldNode).UpdateColumns(Node{ModTime: modTime, FileSize: size})

}