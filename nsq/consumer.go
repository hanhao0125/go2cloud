package nsq

import (
	"go2cloud/config"
	"go2cloud/db"
	"log"
	"strconv"
	"sync"

	"github.com/go-ego/riot"
	"github.com/go-ego/riot/types"
	"github.com/nsqio/go-nsq"
)

type Consumer struct {
	engine riot.Engine
}

func (c *Consumer) HandleMessage(message *nsq.Message) error {
	log.Println("NSQ message received:")
	// process message
	mes := string(message.Body)
	id, err := strconv.Atoi(mes)
	if err != nil {
		log.Printf("cannot atoi %s", mes)
		return err
	}
	node := db.FetchReadableFileNodeById(id)
	c.engine.Index(strconv.Itoa(node.Id), types.DocData{Content: node.Content})
	// search.UpdateIndex(node)

	log.Println("success process message:", mes)
	return nil
}

// func handler(message *nsq.Message) error {
// 	log.Println("NSQ message received:")
// 	// process message
// 	mes := string(message.Body)
// 	id, err := strconv.Atoi(mes)
// 	if err != nil {
// 		log.Printf("cannot atoi %s", mes)
// 		return err
// 	}
// 	node := db.FetchReadableFileNodeById(id)
// 	search.UpdateIndex(node)

// 	log.Println("success process message:", mes)
// 	return nil
// }
// TODO: how to process engine to cousumer
func RecieveAndProcess(engine riot.Engine) {
	wg := &sync.WaitGroup{}
	wg.Add(1)

	decodeConfig := nsq.NewConfig()
	c, err := nsq.NewConsumer(config.IndexedTextTopic, config.IndexedTextChan, decodeConfig)
	if err != nil {
		log.Panic("Could not create consumer")
	}
	//c.MaxInFlight defaults to 1

	c.AddHandler(&Consumer{engine: engine})

	err = c.ConnectToNSQD(config.NSQAddress)
	if err != nil {
		log.Panic("Could not connect")
	}
	log.Println("Awaiting messages from NSQ topic \"My NSQ Topic\"...")
	wg.Wait()
}
