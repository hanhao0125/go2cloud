package common

import "time"

type Node struct {
	Id        int
	Path      string
	ParentDir string
	ParentId  int
	ModTime   time.Time
	FileSize  int64
	Share     int
	FileType  string
	Indexed   int
	FullPath  string
}

func (Node) TableName() string {
	return "filenode"
}

type Image struct {
	Id         int
	Tag        string
	Top5       string
	Path       string
	Upath      string
	UploadDate time.Time
	FileId     int
}

func (Image) TableName() string {
	return "image"
}

type ReadableFileNode struct {
	Id       int
	Path     string
	Content  string
	FileType string
}
