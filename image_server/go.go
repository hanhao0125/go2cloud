//Nsq发送测试
package main

import (
	"fmt"
	"io/ioutil"

	"github.com/nsqio/go-nsq"
)

var producer *nsq.Producer

func PublishMessage(topic string, mes string) {
	strIP1 := "127.0.0.1:4150"
	InitProducer(strIP1)
	defer producer.Stop()
	publish(topic, mes)
}

func PublishBatchMessage(topic string, message []string) {
	strIP1 := "127.0.0.1:4150"
	InitProducer(strIP1)
	defer producer.Stop()
	for _, m := range message {
		publish(topic, m)
	}
}
func listAll(path string) {
	files, _ := ioutil.ReadDir(path)
	cnt := 0
	var paths []string
	for _, fi := range files {
		if fi.IsDir() {
			// ignore dir
			//listAll(path + "/" + fi.Name())
			// println(path + "/" + fi.Name())
		} else {
			cnt++
			println(path + "/" + fi.Name())
			paths = append(paths, path+fi.Name())
		}
	}
	fmt.Print(cnt)
	PublishBatchMessage("tag", paths)
}

// // 主函数
// func main() {
// 	listAll("/Users/hanhao/Downloads/ILSVRC2012_img_test/")
// }

// 初始化生产者
func InitProducer(str string) {
	var err error
	fmt.Println("address: ", str)
	producer, err = nsq.NewProducer(str, nsq.NewConfig())
	if err != nil {
		panic(err)
	}
}

//发布消息
func publish(topic string, message string) error {
	var err error
	if producer != nil {
		if message == "" { //不能发布空串，否则会导致error
			return nil
		}
		err = producer.Publish(topic, []byte(message)) // 发布消息
		return err
	}
	return fmt.Errorf("producer is nil=", err)
}
