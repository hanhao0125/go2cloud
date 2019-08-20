package common

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	sll "github.com/emirpasic/gods/lists/singlylinkedlist"

	mapset "github.com/deckarep/golang-set"
)

func init() {
	log.SetPrefix("[ scan ] ")
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.LUTC)
}

var (
	Paths         = make([]string, 0)
	CurrentPaths  = mapset.NewSet()
	PreviousPaths = mapset.NewSet()
	InitMap       = make(map[string]string)
)

func Insert(fileInfo os.FileInfo, fullPath string, parentId int) int {
	// fullPath := strings.Replace(absPath, MountedPath, "", -1)

	// exist , return
	if ExistPath(fullPath) {
		log.Println("error")
		return -2
	}
	parentPath := ""
	if parentId == 0 {
		parentPath = "/"
	} else {
		parentNode := FetchFileNodeById(parentId)
		parentPath = parentNode.FullPath + "/"
	}

	fileType := ""
	readable, image := false, false

	if fileInfo.IsDir() {
		fileType = "dir"
	} else {
		readable, image, fileType = GetFileType(fileInfo.Name())
	}

	n := Node{FileSize: fileInfo.Size(), FullPath: fullPath, Path: fileInfo.Name(), Share: ShareSingal,
		ParentDir: parentPath, ModTime: fileInfo.ModTime().String(), FileType: fileType, ParentId: parentId,
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

// to init the db
func DBScanRootPath(path string, parentId int) {
	absPath := MountedPath + path
	files, _ := ioutil.ReadDir(absPath)
	for _, info := range files {
		full_path := path + info.Name()
		InitMap[full_path] = info.ModTime().String()
		if info.IsDir() {
			pid := Insert(info, full_path, parentId)
			DBScanRootPath(full_path+"/", pid)
		} else {
			Insert(info, full_path, parentId)
		}
	}
}

var wg = sync.WaitGroup{}

func DBScanRootPathWithGo(path string, parentId int) {
	absPath := MountedPath + path
	files, _ := ioutil.ReadDir(absPath)
	for _, info := range files {
		full_path := path + info.Name()
		// InitMap[full_path] = info.ModTime().String()
		if info.IsDir() {
			// pid := Insert(info, full_path, parentId)
			// wg.Add(1)
			// fmt.Println(runtime.NumGoroutine())
			func() {
				DBScanRootPathWithGo(full_path+"/", -1)
				// wg.Done()
			}()
		} else {
			// Insert(info, full_path, parentId)
		}
	}
}

func DBScanRootPathWithNonRecur(path string, parentId int) {
	absPath := MountedPath + path
	// P := []string{path}
	z := 0
	list := sll.New()
	list.Add(path)
	for !list.Empty() {
		k, _ := list.Get(0)
		path = k.(string)
		list.Remove(0)
		absPath = MountedPath + path
		// fmt.Println(P[0])
		files, _ := ioutil.ReadDir(absPath)
		for _, info := range files {
			z += 1
			full_path := path + info.Name()
			// InitMap[full_path] = info.ModTime().String()
			if info.IsDir() {
				list.Add(full_path + "/")
				// pid := Insert(info, full_path, parentId)
				// wg.Add(1)
				// fmt.Println(runtime.NumGoroutine())
				// func() {
				// DBScanRootPathWithGo(full_path+"/", -1)
				// wg.Done()
				// }()
			} else {
				// Insert(info, full_path, parentId)
			}
		}
	}
}
func GetAllFilesFromDisk(path string) {
	absPath := MountedPath + path
	files, _ := ioutil.ReadDir(absPath)
	for _, info := range files {
		if info.IsDir() {
			Paths = append(Paths, path+info.Name())
			GetAllFilesFromDisk(path + info.Name() + "/")
		} else {
			Paths = append(Paths, path+info.Name())
		}
	}
}

func DetectUpdate(path string) {
	absPath := MountedPath + path
	files, err := ioutil.ReadDir(absPath)
	if err != nil {
		log.Println("err,", err)
	}
	for _, info := range files {
		p := path + info.Name()
		if v, ok := InitMap[p]; ok {
			// exist path, check if update
			if v == info.ModTime().String() {
				// no update
			} else {

				if info.IsDir() {
					DetectUpdate(p + "/")
				}
				InitMap[p] = info.ModTime().String()
			}
			// doesn't exist, add to  initMap for next usage.
		} else {
			log.Println("add new path:", path)
			UpdateHandler(info, path)
			InitMap[p] = info.ModTime().String()
		}
	}
}
func UpdateHandler(info os.FileInfo, parentDir string) {
	if parentDir == "/" {
		pid := 0
		id := Insert(info, parentDir+info.Name(), pid)
		if info.IsDir() {
			DBScanRootPath(parentDir+info.Name(), id)
		}
		return
	}
	p, err := FetchNodeByFullPath(parentDir[:len(parentDir)-1])

	if err != nil {
		log.Println("no such parent,err=", err)
	}
	id := Insert(info, parentDir+info.Name(), p.Id)
	if info.IsDir() {
		DBScanRootPath(parentDir+info.Name()+"/", id)
	}

}
func DetectDelete(path string) {
	absPath := MountedPath + path
	files, err := ioutil.ReadDir(absPath)
	if err != nil {
		log.Println("cannot read from system, err=", err)
	}
	// first fetch all childs that belongs to path
	childs, err := FetchChildByParentDir(path[:len(path)-1])
	if err != nil {
		log.Println("cannot read from db, err= ", err)
	}
	// add system files to set for next quickly contains computation.
	t := mapset.NewSet()
	for _, v := range files {
		t.Add(v.Name())
	}

	for _, v := range childs {
		// system doesn't have the path , update db.
		if !t.Contains(v.Path) {
			log.Println("delete", path+v.Path)
			deleteHandler(path + v.Path)
			delete(InitMap, path+v.Path)
		} else {
			// only traversing the updated folders.( only changed folders can exist deleted files.)
			if InitMap[path+v.Path] != v.ModTime && v.FileType == "dir" {
				log.Println("fuck", path+v.Path, v.ModTime, InitMap[path+v.Path])
				DetectDelete(path + v.Path + "/")

			}
		}
	}
}
func deleteHandler(path string) {
	node, err := FetchNodeByFullPath(path)
	if NSQEnabled {
		PublishToRemoveIndex(node.Id)
	}
	if err != nil {
		log.Fatal("find error", err)

	}
	if node.FileType == "dir" {
		db.Where("parent_id = ?", node.Id).Delete(Node{})
	}
	db.Delete(&node)
}

func InitAndKeepScan() {
	start := time.Now()
	DBScanRootPath("/", 0)
	fmt.Println(" init need", time.Since(start), len(InitMap))
	for range time.Tick(time.Second * 2) {
		start = time.Now()
		DetectUpdate("/")
		fmt.Println(" detect update need", time.Since(start))
		start = time.Now()
		DetectDelete("/")
		fmt.Println(" detect delete need", time.Since(start))
	}
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
