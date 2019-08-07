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
	ch    = make(chan string)
	Paths = make([]string, 0)
)

type M struct {
	Id2Node map[int]Node
	Path2Id map[string]int
	P2C     map[int][]int
	Lock    sync.Mutex
}

func (m *M) InitM() {
	m.Id2Node = make(map[int]Node)
	m.Path2Id = make(map[string]int)
	m.P2C = make(map[int][]int)
}

func (m *M) InsertFileNode(file os.FileInfo, parentDir string, parentId int, fileType string) int {
	node := Node{Id: GenerateGID(), FileType: fileType, Path: file.Name(), ParentDir: parentDir, ParentId: parentId,
		ModTime: file.ModTime(), FileSize: file.Size(), Share: ShareSingal, FullPath: parentDir + file.Name()}

	m.Id2Node[node.Id] = node
	m.Path2Id[node.FullPath] = node.Id
	return node.Id
}

func (m *M) ExistId(id int) bool {
	_, ok := m.Id2Node[id]
	return ok
}
func (m *M) ExistPath(path string) bool {
	_, ok := m.Path2Id[path]
	return ok
}
func (m *M) ExistP2C(pid int) bool {
	_, ok := m.P2C[pid]
	return ok
}

func (m *M) FetchNodeByFullPath(fullPath string) (Node, bool) {
	node := Node{}
	if !m.ExistPath(fullPath) {
		return node, false
	}
	return m.Id2Node[m.Path2Id[fullPath]], true
}
func (m *M) Update() {

}

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
func (m *M) Scan(path string) {
	absPath := MountedPath + path
	files, _ := ioutil.ReadDir(absPath)
	for _, info := range files {
		ap := absPath + info.Name()
		if info.IsDir() {
			if _, ok := m.Path2Id[getRelativePath(ap)]; ok {
				// exist
				// fmt.Println(id)
			} else {
				// didn't exist, add
				m.Insert(info, ap)
			}
			m.Scan(path + info.Name() + "/")
		} else {
			if _, ok := m.Path2Id[getRelativePath(ap)]; ok {
				// exist
			} else {
				// didn't exist, add
				m.Insert(info, ap)
			}
		}
	}
}

func (m *M) Insert(fileInfo os.FileInfo, absPath string) {
	m.Lock.Lock()
	fullPath := strings.Replace(absPath, MountedPath, "", -1)
	if m.ExistPath(fullPath) {
		return
	}
	parentPath := strings.Replace(fullPath, fileInfo.Name(), "", -1)

	fileType := ""
	if fileInfo.IsDir() {
		fileType = "dir"
	} else {
		_, _, fileType = GetFileType(fileInfo.Name())
	}
	parentId := 0
	if parentPath == RootParentDir {
		parentId = RootParentId
	} else {
		parentId = m.Path2Id[parentPath]
	}
	defer m.Lock.Unlock()
	n := Node{Id: GenerateGID(), FileSize: fileInfo.Size(), FullPath: fullPath, Path: fileInfo.Name(), Share: ShareSingal,
		ParentDir: parentPath, ModTime: fileInfo.ModTime(), FileType: fileType, ParentId: parentId}
	m.Id2Node[n.Id] = n
	m.Path2Id[fullPath+"/"] = n.Id
	if m.ExistP2C(parentId) {
		m.P2C[parentId] = append(m.P2C[parentId], n.Id)
	} else {
		m.P2C[parentId] = make([]int, 0)
		m.P2C[parentId] = append(m.P2C[parentId], n.Id)
	}
}

func (m *M) FetchNodesByParentId(pid int) []Node {
	nodes := []Node{}
	childs, ok := m.P2C[pid]
	if !ok {
		log.Println("empty dir,", pid)
		return nodes
	}

	for _, c := range childs {
		nodes = append(nodes, m.Id2Node[c])
	}
	return nodes
}
func (m *M) FetchFileNodeById(id int) Node {
	node, ok := m.Id2Node[id]
	if !ok {
		log.Panic("error when fetch node by id")
	}
	return node
}

func ScanRootPath(path string, m *M) {
	absPath := MountedPath + path
	files, _ := ioutil.ReadDir(absPath)
	for _, info := range files {
		ap := absPath + info.Name()
		if info.IsDir() {
			m.Insert(info, ap)
			// paths = append(paths, path+info.Name()+"/")
			// ch <- path + info.Name() + "/"
			ScanRootPath(path+info.Name()+"/", m)
		} else {
			m.Insert(info, ap)
		}
	}
}
func TimeScan(m *M) {
	for range time.Tick(time.Second * 1000) {
		log.Println("Starting scan path")
		m.Scan("/")
	}
}
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
func Insert(fileInfo os.FileInfo, absPath string) {
	fullPath := strings.Replace(absPath, MountedPath, "", -1)

	// exist , return
	if ExistPath(fullPath) {
		return
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

	if parentPath == RootParentDir {
		parentId = RootParentId
	} else {
		// delete the `/`
		parentNode, err := GetNodeByFullPath(parentPath[:len(parentPath)-1])
		if err != nil {
			log.Println("err=", err)
		}
		parentId = parentNode.Id
	}

	n := Node{FileSize: fileInfo.Size(), FullPath: fullPath, Path: fileInfo.Name(), Share: ShareSingal,
		ParentDir: parentPath, ModTime: fileInfo.ModTime(), FileType: fileType, ParentId: parentId,
		Image: image, Readable: readable}
	db.Create(&n)

	// if readable , then publish to nsq for next indexed
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

// to init the db
func DBScanRootPath(path string) {
	absPath := MountedPath + path
	files, _ := ioutil.ReadDir(absPath)
	for _, info := range files {
		ap := absPath + info.Name()
		if info.IsDir() {
			Insert(info, ap)
			DBScanRootPath(path + info.Name() + "/")
		} else {
			Insert(info, ap)
		}
	}
}

// TODO use global Paths
func GetAllFilesFromDisk(path string) {
	absPath := MountedPath + path
	files, _ := ioutil.ReadDir(absPath)
	for _, info := range files {
		// ap := absPath + info.Name()
		if info.IsDir() {
			Paths = append(Paths, path+info.Name())
			// t := make([]string, 0)
			// t = GetAllFilesFromDisk(path+info.Name()+"/", t)
			// paths = append(paths, GetAllFilesFromDisk(path+info.Name()+"/", paths)...)
			// return paths
			// paths = append(paths, GetAllFilesFromDisk(path+info.Name()+"/")...)
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
	relativePath := strings.Replace(path, MountedPath, "", -1)
	// first insert no matter `dir` or `file`
	Insert(info, path)
	if info.IsDir() {
		// dir , Scan from path
		DBScanRootPath(relativePath + "/")
	}
}

// TODO may can be done by multi go
func Compare() {
	defer RunTime(time.Now())
	Paths = Paths[:0]
	// first read all path from file system
	GetAllFilesFromDisk("/")
	fmt.Println(len(Paths))

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
