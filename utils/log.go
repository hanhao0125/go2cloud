package utils

import (
	"log"
	"strings"
)

func init() {
	log.SetPrefix("[SEARCH]")
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.LUTC)
}

func GenerateLogger(keywords string) {
	log.SetPrefix("[" + strings.ToUpper(keywords) + "]")
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.LUTC)
}
