package common

const (
	SyncServerAddress    string = "localhost:8888"
	MysqlPath            string = "root:root@/cloud?charset=utf8mb4&parseTime=True&loc=Local"
	MountedPath          string = "/Users/hanhao"
	ReplaceSpace         string = "^"
	RedisAddress         string = "127.0.0.1:6379"
	RedisPassword        string = "123456"
	ShareSingal          int    = 0
	NSQAddress           string = "127.0.0.1:4150"
	TagImageTopic        string = "tag"
	TagImageChan         string = "c1"
	IndexedTextTopic     string = "index"
	IndexedTextChan      string = "c1"
	RemoveIndexTextTopic string = "remove_index"
	RemoveIndexTextChan  string = "c1"
	RootParentId         int    = -1
	RootParentDir        string = "/"
	SearchServicePort    string = ":9000"
	WebServicePort       string = ":9205"
	NSQEnabled           bool   = false
	SQLitePath           string = "../cloud.db"
)

var (
	GID int = 0
)
