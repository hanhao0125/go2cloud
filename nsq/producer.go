package nsq

import (
	"fmt"
	"go2cloud/config"
	"io/ioutil"

	"github.com/nsqio/go-nsq"
)

var (
	producer *nsq.Producer
)

func PublishMessage(topic string, mes string) {
	InitProducer(config.NSQAddress)
	defer producer.Stop()
	publish(topic, mes)
}

func PublishBatchMessage(topic string, message []string) {
	InitProducer(config.NSQAddress)
	defer producer.Stop()
	for _, m := range message {
		publish(topic, m)
	}
}
func publishFromDir(path string) {
	files, _ := ioutil.ReadDir(path)
	cnt := 0
	var paths []string
	for _, fi := range files {
		if fi.IsDir() {
			// ignore dir
		} else {
			cnt++
			println(path + "/" + fi.Name())
			paths = append(paths, path+fi.Name())
		}
	}
	fmt.Print(cnt)
	PublishBatchMessage(config.TagImageTopic, paths)
}

func InitProducer(str string) {
	var err error
	fmt.Println("address: ", str)
	producer, err = nsq.NewProducer(str, nsq.NewConfig())
	if err != nil {
		panic(err)
	}
}

func publish(topic string, message string) error {
	var err error
	if producer != nil {
		if message == "" { //不能发布空串，否则会导致error
			return nil
		}
		err = producer.Publish(topic, []byte(message)) // 发布消息
		return err
	}
	return fmt.Errorf("producer is nil=%v", err)
}
