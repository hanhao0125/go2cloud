package utils

import (
	"fmt"
	"io/ioutil"
)

func TestUtils() {
	fmt.Println("fuckaaaa")
}
func ReadFiles(path string) string {
	b, e := ioutil.ReadFile(path)

	if e != nil {
		fmt.Println("read file error")
		return "error"
	}
	return string(b)
}
