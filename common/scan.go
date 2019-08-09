package common

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	mapset "github.com/deckarep/golang-set"
)

var (
	Paths             = make([]string, 0)
	CurrentPaths      = mapset.NewSet()
	PreviousPaths     = mapset.NewSet()
	cnt           int = 0
)

func getRelativePath(absPath string) string {
	return strings.Replace(absPath, MountedPath, "", -1) + "/"
}

//对比出差异
/*
	新增加的、删除掉的、更新的
	新增加的：直接添加即可
	更新的：直接更新
	删除掉的：
*/

func ExistPath(fullPath string) bool {
	node := Node{}
	q := db.Where("full_path = ?", fullPath).First(&node)
	if q.Error != nil {
		return false
	}
	return true
}

func GetNodeByFullPath(fullPath string) (Node, error) {
	node := Node{}
	q := db.Where("full_path = ?", fullPath).First(&node)
	if q.Error != nil {
		log.Println("find error , full_path", q.Error)
		return node, q.Error
	}
	return node, nil
}
func Insert(fileInfo os.FileInfo, absPath string, pid int) int {
	fullPath := strings.Replace(absPath, MountedPath, "", -1)

	// exist , return
	if ExistPath(fullPath) {
		log.Println("error")
		return -2
	}
	parentPath := strings.Replace(fullPath, fileInfo.Name(), "", -1)
	parentId := 0

	fileType := ""
	readable, image := false, false

	if fileInfo.IsDir() {
		fileType = "dir"
	} else {
		readable, image, fileType = GetFileType(fileInfo.Name())
	}

	// if parentPath == RootParentDir {
	// 	parentId = RootParentId
	// } else {
	// 	// delete the `/`
	// 	parentNode, err := GetNodeByFullPath(parentPath[:len(parentPath)-1])
	// 	if err != nil {
	// 		log.Println("err=", err)
	// 		log.Println(parentPath[:len(parentPath)-1], fullPath)
	// 	}
	// 	parentId = parentNode.Id
	// }
	parentId = pid

	n := Node{FileSize: fileInfo.Size(), FullPath: fullPath, Path: fileInfo.Name(), Share: ShareSingal,
		ParentDir: parentPath, ModTime: fileInfo.ModTime(), FileType: fileType, ParentId: parentId,
		Image: image, Readable: readable}
	q := db.Create(&n)
	if q.Error != nil {
		log.Println("error when create new node, err= ", q.Error)
	}

	// if readable , then publish to nsq for next indexed
	if NSQEnabled {
		if readable {
			sid := strconv.Itoa(n.Id)
			PublishMessage(IndexedTextTopic, sid)
		}
		// if it's image, then publis to nsq for next tag
		if image {
			sid := strconv.Itoa(n.Id)
			PublishMessage(TagImageTopic, sid)
		}
	}
	return n.Id
}

func T() {
	g := &sync.WaitGroup{}
	cnt = 0
	g.Add(1)
	DBScanRootPath("/", -1, g)
	fmt.Println("process fire count: ", cnt)
	g.Wait()
}

// to init the db
func DBScanRootPath(path string, parentId int, g *sync.WaitGroup) {
	defer g.Done()
	absPath := MountedPath + path
	files, _ := ioutil.ReadDir(absPath)
	for _, info := range files {
		cnt++
		ap := absPath + info.Name()
		if info.IsDir() {
			g.Add(1)
			pid := Insert(info, ap, parentId)
			go DBScanRootPath(path+info.Name()+"/", pid, g)
		} else {
			Insert(info, ap, parentId)
		}
	}
}

// TODO use global Paths
func GetAllFilesFromDisk(path string) {
	absPath := MountedPath + path
	files, _ := ioutil.ReadDir(absPath)
	for _, info := range files {
		// ap := absPath + info.Name()
		CurrentPaths.Add(path + info.Name())
		if info.IsDir() {
			Paths = append(Paths, path+info.Name())
			GetAllFilesFromDisk(path + info.Name() + "/")
		} else {
			Paths = append(Paths, path+info.Name())
		}
	}
}

//when compare the difference , use this function to handle create event, works for `dir` and regular file.
func InsertNotExistInDB(path string) {
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

// TODO may can be done by multi go
func Compare() {
	defer RunTime(time.Now())
	Paths = Paths[:0]
	// first read all path from file system
	GetAllFilesFromDisk("/")
	fmt.Println(len(Paths))
	neededAdd := CurrentPaths.Difference(PreviousPaths)
	neededDelete := PreviousPaths.Difference(CurrentPaths)

	for v := range neededAdd.Iter() {
		if v, ok := v.(string); ok {
			InsertNotExistInDB(v)
		}
	}
	for v := range neededDelete.Iter() {
		if v, ok := v.(string); ok {
			// fetch then delete it.
			node, err := FetchNodeByFullPath(v)
			if err != nil {
				log.Println("no such node,err = ", err, v)
			}
			if node.FileType == "dir" {
				db.Where("parent_id = ?", node.Id).Delete(Node{})
			}
			db.Delete(&node)
		}
	}

	// compare with db
	addCnt, deleteCnt := 0, 0
	for _, p := range Paths {
		_, err := GetNodeByFullPath(p)
		// db doesn't contain this path , insert this path
		// implement the high-level Insert method that can handle insert `dir` event.
		// err != nil means db doesn't contain this path, insert to db.
		if err != nil {
			log.Println("new path, insert to db:", p)
			// can handle `dir` and `file`
			InsertNotExistInDB(p)
			addCnt++
		}
		// for now, don't care update event.
	}
	if addCnt == 0 {
		log.Println("nothing needed to be add")
	} else {
		log.Println("files added:", addCnt)
	}
	nodes := GetAllNodes()
	pathSet := mapset.NewSet()
	for _, p := range Paths {
		pathSet.Add(p)
	}
	for _, n := range nodes {
		if pathSet.Contains(n.FullPath) {
			// do nothing. maybe update

		} else {
			log.Println("delete event")
			PublishToRemoveIndex(n.Id)
			// delete the old path
			// if dir, then delete itself and where parent_id = n.Id
			// if not dir, then only need delete itself
			if n.FileType == "dir" {
				// first delete childs
				db.Where("parent_id = ?", n.Id).Delete(Node{})
			}
			// delete node
			// TODO need publis message to search engine to remove the related index
			db.Delete(&n)
			deleteCnt++
		}
	}
	if deleteCnt == 0 {
		log.Println("nothing needed to be deleted")
	} else {
		log.Println("files deleted: ", deleteCnt)
	}
	Paths = make([]string, 0)

}

// publish batch message
func PublishToRemoveIndex(id int) {
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

func TimelyTask() {
	for range time.Tick(time.Second * 10) {
		Compare()
	}
}

func GetAllNodes() []Node {
	nodes := []Node{}
	db.Find(&nodes)
	return nodes
}
