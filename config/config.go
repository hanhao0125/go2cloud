package config

const (
	SyncServerAddress string = "localhost:8888"
	MysqlPath         string = "root:root@/cloud?charset=utf8&parseTime=True&loc=Local"
	MountedPath       string = "/Users/hanhao/netserver"
	ReplaceSpace      string = "^"
	RedisAddress      string = "127.0.0.1:6379"
	RedisPassword     string = "123456"
	ShareSingal       int    = 0
	NSQAddress        string = "127.0.0.1:4150"
	TagImageTopic     string = "tag"
	TagImageChan      string = "c1"
	IndexedTextTopic  string = "index"
	IndexedTextChan   string = "c1"
)

var (
	ReadableFileType []string = []string{"txt", "code"}
)
