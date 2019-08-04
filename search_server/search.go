package search

import (
	"fmt"
	"go2cloud/db"
	"go2cloud/models"
	"go2cloud/nsq"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-ego/riot"
	"github.com/go-ego/riot/types"
)

func init() {
	log.SetPrefix("[SEARCH]")
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.LUTC)
	searcher.Init(types.EngineOpts{})
}

var (
	// searcher is coroutine safe
	searcher         = riot.Engine{}
	Language         = "lang"
	TextFilesMapping = map[string]string{"pdf": "pdf", "txt": "txt", "go": Language, "java": Language, "py": Language}
	TextFiles        = []string{"txt", "go", "java", "py", "c", "log", "md"}
)

func filterSuffix(p string) bool {
	if p == "" {
		return false
	}
	for _, v := range TextFiles {

		sv := strings.Split(p, ".")
		suffix := sv[len(sv)-1]
		if strings.Compare(suffix, v) == 0 {
			return true
		}
	}
	return false
}

func GetTextFilesPath(rootPath string) []string {
	var files []string
	err := filepath.Walk(rootPath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if path == "" {
				return nil
			}
			if filterSuffix(path) {
				files = append(files, path)
			}
			return nil
		})
	if err != nil {
		log.Println(err)
	}
	return files
}
func IndexFromDB() {
	readableNodes := db.FetchReadableFileNode()

	defer searcher.Close()
	for _, v := range readableNodes {
		searcher.Index(strconv.Itoa(v.Id), types.DocData{Content: v.Content})
	}
	// 等待索引刷新完毕
	searcher.Flush()
}

func Search(query string) []models.Node {
	req := types.SearchReq{Text: query}
	resp := searcher.SearchDoc(req)
	log.Println("start query")
	searchRet := []models.Node{}
	for _, c := range resp.Docs {
		fmt.Println(c.DocId)
		id, _ := strconv.Atoi(c.DocId)
		searchRet = append(searchRet, db.FetchFileNodeById(id))
	}
	log.Printf("find realted `%s` files:%d", query, len(searchRet))
	return searchRet
}

// update the engine index. This operation will lock the search engine.
func UpdateIndex(node models.ReadableFileNode) {
	searcher.Index(strconv.Itoa(node.Id), types.DocData{Content: node.Content})
}

func StartSearchHttpService() {
	IndexFromDB()
	go nsq.RecieveAndProcess(searcher)
	gin.ForceConsoleColor()
	router := gin.Default()
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"abc": "1"})
	})
	// router.LoadHTMLGlobtestmplates/*")
	router.GET("/search", func(c *gin.Context) {
		query := c.Query("query")
		res := Search(query)
		c.JSON(200, gin.H{
			"searchResults": res,
		})
	})
	router.Run(":9000")
}
