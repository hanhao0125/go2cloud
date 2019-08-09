package common

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set"
	"github.com/garyburd/redigo/redis"
	"github.com/rs/xid"
)

var Previous = mapset.NewSet()
var Current = mapset.NewSet()
var pool *redis.Pool //创建redis连接池
var P []string
var V []Tmp

func init() {
	pool = &redis.Pool{ //实例化一个连接池
		MaxIdle: 30, //最初的连接数量
		// MaxActive:1000000,    //最大连接数量
		// MaxActive:   3000, //连接池最大连接数量,不确定可以用0（0表示自动定义），按需分配
		IdleTimeout: 300, //连接关闭时间 300秒 （300秒不使用自动关闭）
		Dial: func() (redis.Conn, error) { //要连接的redis数据库
			return redis.Dial("tcp", "localhost:6379")
		},
	}
}

func set(name, pd, pid string, modTime time.Time, size int64,
	share int, fileType string, fullPath string, image, readable bool, tag string, r redis.Conn) string {

	id := xid.New().String()

	// r.Do("HMSET", id, "id", id, "name", name, "pd", pd, "pid", pid, "modTime",
	// 	modTime, "size", size, "share", share, "fileType", fileType, "fullPath", fullPath, "image", image, "readable", readable, "tag", tag)
	r.Do("HMSET", id, "id", id)
	r.Do("HMSET", id, "name", name)
	r.Do("HMSET", id, "pd", pd)
	r.Do("HMSET", id, "pid", pid)
	r.Do("HMSET", id, "modTime", modTime)
	r.Do("HMSET", id, "size", size)
	r.Do("HMSET", id, "share", share)
	r.Do("HMSET", id, "fileType", fileType)
	r.Do("HMSET", id, "fullPath", fullPath)
	r.Do("HMSET", id, "image", image)
	r.Do("HMSET", id, "readable", readable)
	r.Do("HMSET", id, "tag", tag)
	return id
}

type Tmp struct {
	Id        string
	Path      string
	ParentDir string
	ParentId  string
	ModTime   time.Time
	FileSize  int64
	Share     int
	FileType  string
	Indexed   int
	FullPath  string
	Image     bool
	Readable  bool
	Tag       string
}

// how to quickly find files belong to a parent dir
func RInsert(fileInfo os.FileInfo, absPath string, pid string) string {
	fullPath := strings.Replace(absPath, MountedPath, "", -1)

	parentPath := strings.Replace(fullPath, fileInfo.Name(), "", -1)
	parentId := ""

	fileType := ""
	readable, image := false, false

	if fileInfo.IsDir() {
		fileType = "dir"
	} else {
		readable, image, fileType = GetFileType(fileInfo.Name())
	}

	parentId = pid
	id := xid.New().String()
	n := Tmp{Id: id, FileSize: fileInfo.Size(), FullPath: fullPath, Path: fileInfo.Name(), Share: ShareSingal,
		ParentDir: parentPath, ModTime: fileInfo.ModTime(), FileType: fileType, ParentId: parentId,
		Image: image, Readable: readable}

	V = append(V, n)

	// if readable , then publish to nsq for next indexed
	// if NSQEnabled {
	// 	if readable {
	// 		sid := strconv.Itoa(n.Id)
	// 		PublishMessage(IndexedTextTopic, sid)
	// 	}
	// 	// if it's image, then publis to nsq for next tag
	// 	if image {
	// 		sid := strconv.Itoa(n.Id)
	// 		PublishMessage(TagImageTopic, sid)
	// 	}
	// }
	return n.Id
}
func RSet(r redis.Conn) {
	for _, v := range V {
		r.Do("HMSET", v.Id,
			"id", v.Id,
			"name", v.Path,
			"pd", v.ParentDir,
			"pid", v.ParentId,
			"modTime", v.ModTime,
			"size", v.FileSize,
			"share", v.Share,
			"fileType", v.FileType,
			"fullPath", v.FullPath,
			"image", v.Image,
			"readable", v.Readable,
			"tag", v.Tag)
	}
}

func RT() {
	r := pool.Get()
	_, err := r.Do("auth", "123456")
	defer r.Close()
	if err != nil {
		log.Panic("auth error,err=", err)
	}
	RDBScanRootPath("/", "-1")
	fmt.Println(len(V))
	RSet(r)
}

// to init the db
func RDBScanRootPath(path string, parentId string) {
	// defer g.Done()
	absPath := MountedPath + path
	files, _ := ioutil.ReadDir(absPath)
	for _, info := range files {
		ap := absPath + info.Name()
		Previous.Add(ap)
		if info.IsDir() {
			pid := RInsert(info, ap, parentId)
			RDBScanRootPath(path+info.Name()+"/", pid)
		} else {
			RInsert(info, ap, parentId)
		}
	}
}

//when compare the difference , use this function to handle create event, works for `dir` and regular file.
func RInsertNotExistInDB(path string) {
	path = MountedPath + path
	info, err := os.Stat(path)
	if err != nil {
		log.Println("err = ", err)
	}
	// relativePath := strings.Replace(path, MountedPath, "", -1)
	// first insert no matter `dir` or `file`
	Insert(info, path, -1)
	if info.IsDir() {
		// dir , Scan from path
		// DBScanRootPath(relativePath+"/", -1)
	}
}

// TODO use global Paths
func RGetAllFilesFromDisk(path string) {
	absPath := MountedPath + path
	files, _ := ioutil.ReadDir(absPath)
	for _, info := range files {
		ap := absPath + info.Name()
		Current.Add(ap)
		if info.IsDir() {
			RGetAllFilesFromDisk(path + info.Name() + "/")
		} else {
			// Paths = append(Paths, path+info.Name())
		}
	}
}

// TODO may can be done by multi go
func RCompare() {
	Previous.Clear()
	Previous, Current = Current, Previous
	// first read all path from file system. in Current
	RGetAllFilesFromDisk("/")
	fmt.Println("load all")
	for v := range Current.Iter() {
		if v, ok := v.(string); ok {
			if strings.Contains(v, "log") {
				fmt.Println(v)
			}
		}
	}
	// Current means lastest files, needed to compare with previous. Previous also means the redis data
	// first find new path. new path = Current-Previous
	neededAdd := Current.Difference(Previous)
	neededDelete := Previous.Difference(Current)
	fmt.Println(neededAdd.Cardinality())
	fmt.Println(neededDelete.Cardinality())
	for v := range neededAdd.Iter() {
		if v, ok := v.(string); ok {
			// added to redis
			fmt.Println(v)
		}
	}
	for v := range neededDelete.Iter() {
		if v, ok := v.(string); ok {
			// delete from redis
			fmt.Println(v)
		}
	}
	// compare with db
	// addCnt, deleteCnt := 0, 0
	// for _, p := range Paths {
	// 	if !Previous.Contains(p) {
	// 		log.Println("new path, insert to db:", p)
	// 		// can handle `dir` and `file`
	// 		InsertNotExistInDB(p)
	// 		addCnt++
	// 	}
	// 	// judge by redis, redis cache all fullpath in a set.
	// 	_, err := GetNodeByFullPath(p)
	// 	// db doesn't contain this path , insert this path
	// 	// implement the high-level Insert method that can handle insert `dir` event.
	// 	// err != nil means db doesn't contain this path, insert to db.
	// 	if err != nil {
	// 		log.Println("new path, insert to db:", p)
	// 		// can handle `dir` and `file`
	// 		InsertNotExistInDB(p)
	// 		addCnt++
	// 	}
	// 	// for now, don't care update event.
	// }
	// if addCnt == 0 {
	// 	log.Println("nothing needed to be add")
	// } else {
	// 	log.Println("files added:", addCnt)
	// }
	// nodes := GetAllNodes()
	// pathSet := mapset.NewSet()
	// for _, p := range Paths {
	// 	pathSet.Add(p)
	// }
	// for _, n := range nodes {
	// 	if pathSet.Contains(n.FullPath) {
	// 		// do nothing. maybe update

	// 	} else {
	// 		log.Println("delete event")
	// 		PublishToRemoveIndex(n.Id)
	// 		// delete the old path
	// 		// if dir, then delete itself and where parent_id = n.Id
	// 		// if not dir, then only need delete itself
	// 		if n.FileType == "dir" {
	// 			// first delete childs
	// 			db.Where("parent_id = ?", n.Id).Delete(Node{})
	// 		}
	// 		// delete node
	// 		// TODO need publis message to search engine to remove the related index
	// 		db.Delete(&n)
	// 		deleteCnt++
	// 	}
	// }
	// if deleteCnt == 0 {
	// 	log.Println("nothing needed to be deleted")
	// } else {
	// 	log.Println("files deleted: ", deleteCnt)
	// }
	Paths = make([]string, 0)

}

// publish batch message
func RPublishToRemoveIndex(id int) {
	// first fetch the childs
	nodes, _ := FetchNodesByParentId(id)
	// only need ids
	ids := make([]string, len(nodes)+1)
	for _, n := range nodes {
		ids = append(ids, strconv.Itoa(n.Id))
	}
	ids = append(ids, strconv.Itoa(id))
	PublishBatchMessage(RemoveIndexTextTopic, ids)
}

func RTimelyTask() {
	for range time.Tick(time.Second * 10) {
		RCompare()
	}
}

func RGetAllNodes() []Node {
	nodes := []Node{}
	db.Find(&nodes)
	return nodes
}
