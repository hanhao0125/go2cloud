package main

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-ego/riot"
	"github.com/go-ego/riot/types"
)

func init() {
	log.SetPrefix("[SEARCH]")
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.LUTC)
}

var (
	// searcher is coroutine safe
	searcher         = riot.Engine{}
	Language         = "lang"
	TextFilesMapping = map[string]string{"pdf": "pdf", "txt": "txt", "go": Language, "java": Language, "py": Language}
	TextFiles        = []string{"txt", "go", "java", "py", "c", "log", "md"}
	ReadableFiles    = map[string]BaseDoc{}
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

func Index(path string) {
	var files = GetTextFilesPath(path)

	//create object
	for i, f := range files {
		id := strconv.Itoa(i + 1)
		b := BaseDoc{Id: id, Path: f, Type: "txt"}
		b.ReadContent()

		ReadableFiles[id] = b

	}

	searcher.Init(types.EngineOpts{})
	defer searcher.Close()

	for k, v := range ReadableFiles {
		searcher.Index(k, types.DocData{Content: v.Content})
	}

	searcher.Flush()

}
func Search(query string) []BaseDoc {
	var res []BaseDoc
	req := types.SearchReq{Text: query}
	resp := searcher.SearchDoc(req)
	for _, c := range resp.Docs {
		res = append(res, ReadableFiles[c.DocId])
	}
	return res
}
