package main

import (
	"fmt"
	cn "go2cloud/common"
	"log"
	"strconv"
	"sync"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-ego/riot"
	"github.com/go-ego/riot/types"
	"github.com/nsqio/go-nsq"
)

func init() {
	log.SetPrefix("[SEARCH]")
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.LUTC)
	// searcher.Init(types.EngineOpts{})
	searcher.Init(types.EngineOpts{UseStore: true, StoreFolder: "/Users/hanhao/netserver/riot_index/"})
}

var (
	// searcher is coroutine safe
	searcher = riot.Engine{}
)

func IndexFromDB() {
	readableNodes := cn.FetchReadableFileNode()
	log.Println("find readable files:", len(readableNodes))
	defer searcher.Close()
	fmt.Println("readable files:", len(readableNodes))
	for _, v := range readableNodes {
		searcher.Index(strconv.Itoa(v.Id), types.DocData{Content: v.Content})
	}
	fmt.Println("indexed num:", searcher.NumDocsIndexed())
	// 等待索引刷新完毕
	searcher.Flush()
}

func Search(query string) []cn.Node {
	req := types.SearchReq{Text: query}
	resp := searcher.SearchDoc(req)
	log.Println("start query")
	searchRet := []cn.Node{}

	for _, c := range resp.Docs {
		id, _ := strconv.Atoi(c.DocId)
		searchRet = append(searchRet, cn.FetchFileNodeById(id))
	}
	searchByName := searchImage(query)
	searchRet = append(searchRet, searchByName...)
	log.Printf("find realted `%s` files:%d", query, len(searchRet))
	// currently return 1000 amlost for speed
	if len(searchRet) > 200 {
		searchRet = searchRet[:200]
	}
	return searchRet
}

type AddIndexConsumer struct {
}

func (a *AddIndexConsumer) HandleMessage(message *nsq.Message) error {
	log.Println("NSQ message received:")
	// process message
	mes := string(message.Body)
	id, err := strconv.Atoi(mes)
	if err != nil {
		log.Printf("cannot atoi %s", mes)
		return err
	}
	node := cn.FetchReadableFileNodeById(id)
	// force flush index
	searcher.Index(strconv.Itoa(node.Id), types.DocData{Content: node.Content}, true)

	log.Println("success process message:", mes, "success indexed the search engine")
	return nil
}

// search image by filename and tag
func searchImage(query string) []cn.Node {
	nodes := cn.FetchImageByTagAndName(query, query)
	return nodes
}

func AddIndexService() {
	wg := &sync.WaitGroup{}
	wg.Add(1)

	decodeConfig := nsq.NewConfig()
	c, err := nsq.NewConsumer(cn.IndexedTextTopic, cn.IndexedTextChan, decodeConfig)
	if err != nil {
		log.Panic("Could not create consumer")
	}
	//c.MaxInFlight defaults to 1

	c.AddHandler(&AddIndexConsumer{})

	err = c.ConnectToNSQD(cn.NSQAddress)
	if err != nil {
		log.Panic("Could not connect")
	}
	log.Println("Awaiting messages from NSQ topic ", cn.IndexedTextTopic)
	wg.Wait()
}

func RemoveIndexService() {
	wg := &sync.WaitGroup{}
	wg.Add(1)

	decodeConfig := nsq.NewConfig()
	c, err := nsq.NewConsumer(cn.RemoveIndexTextTopic, cn.RemoveIndexTextChan, decodeConfig)
	if err != nil {
		log.Panic("Could not create consumer")
	}
	//c.MaxInFlight defaults to 1

	c.AddHandler(&RemoveIndexConsumer{})

	err = c.ConnectToNSQD(cn.NSQAddress)
	if err != nil {
		log.Panic("Could not connect")
	}
	log.Println("Awaiting messages from NSQ topic ", cn.RemoveIndexTextTopic)
	wg.Wait()
}

type RemoveIndexConsumer struct {
}

func (r *RemoveIndexConsumer) HandleMessage(message *nsq.Message) error {
	log.Println("NSQ message received:")
	// process message
	mes := string(message.Body)
	id := mes
	// force flush index
	searcher.RemoveDoc(id, true)

	log.Println("success process message:", mes, "success remove index from search engine")
	return nil
}

func StartSearchHttpService() {
	// TODO init from stored index
	IndexFromDB()

	// nsq consumer
	go AddIndexService()
	go RemoveIndexService()

	gin.ForceConsoleColor()
	router := gin.Default()
	router.Use(cors.Default())

	router.GET("/search", func(c *gin.Context) {
		query := c.Query("query")
		res := Search(query)
		c.JSON(200, gin.H{
			"searchResults": res,
		})
	})
	router.GET("/index_info", func(c *gin.Context) {
		c.JSON(200, gin.H{"indexedDocNum": searcher.NumIndexed(), "indexedTokenNum": searcher.NumTokenAdded()})
	})
	router.Run(cn.SearchServicePort)
}

func main() {
	StartSearchHttpService()
}
