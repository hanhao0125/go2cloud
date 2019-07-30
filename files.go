package main

import (
	"fmt"
	"io/ioutil"
)

type BaseDoc struct {
	Id      string
	Path    string
	Content string
	Type    string
}

func (b *BaseDoc) ReadContent() string {
	switch b.Type {
	case "pdf":
		b.Content = "Not supported pdf parse"
	default:
		file, err := ioutil.ReadFile(b.Path)
		if err != nil {
			fmt.Println("error when open file", b.Path, err)
		}
		b.Content = string(file)
	}
	return b.Content
}

type Readable interface {
	Read() string
}

func (t BaseDoc) Read() string {
	file, err := ioutil.ReadFile(t.Path)
	if err != nil {
		fmt.Println("error when open file", t.Path, err)
	}
	t.Content = string(file)
	return t.Content

}

type PDFDoc struct {
	BaseDoc
	DocType string
}

func (p *PDFDoc) Read() string {
	return ""
}

type TxtDoc struct {
	BaseDoc
	DocType string
}

func (p *TxtDoc) Read() string {
	return ""
}

type LanguageDoc struct {
	BaseDoc
	DocType string
}

type MdDoc struct {
	BaseDoc
	DocType string
}
type MSDoc struct {
	BaseDoc
	DocType string
}

func DocFactory(DocType string) Readable {
	switch DocType {
	case "pdf":
		return &PDFDoc{DocType: "pdf"}
	case "txt":
		return &TxtDoc{DocType: "txt"}
	case "lang":
		return &LanguageDoc{DocType: "lang"}
	case "md":
		return &MdDoc{DocType: "md"}
	default:
		return &BaseDoc{}
	}

}
