package common

import (
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	ch = make(chan string)
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
func FileHandler(fileName string) {

}
func DirHandler(dirName string) {

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
		_, fileType = GetFileType(fileInfo.Name())
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
	q := db.Where("full_path = ?", fullPath).First(node)
	if q.Error != nil {
		return false
	}
	return true
}

func GetIdByFullPath(fullPath string) int {
	node := Node{}
	q := db.Where("full_path = ?", fullPath).First(node)
	if q.Error != nil {
		return -1
	}
	return node.Id
}
func Insert(fileInfo os.FileInfo, absPath string) {
	fullPath := strings.Replace(absPath, MountedPath, "", -1)

	if ExistPath(fullPath) {
		return
	}
	parentPath := strings.Replace(fullPath, fileInfo.Name(), "", -1)

	fileType := ""
	if fileInfo.IsDir() {
		fileType = "dir"
	} else {
		_, fileType = GetFileType(fileInfo.Name())
	}
	parentId := 0
	if parentPath == RootParentDir {
		parentId = RootParentId
	} else {
		parentId = GetIdByFullPath(parentPath)
	}
	n := Node{FileSize: fileInfo.Size(), FullPath: fullPath, Path: fileInfo.Name(), Share: ShareSingal,
		ParentDir: parentPath, ModTime: fileInfo.ModTime(), FileType: fileType, ParentId: parentId}
	db.Create(&n)
}
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
