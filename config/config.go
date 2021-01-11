package config

import (
	"github.com/kkkunny/GoMy/log"
	"os"
)

// 日志管理器
var LogMgr *log.Logger
// 根目录
var RootPath string
// 爬虫线程
const SpiderNum = 50

func init() {
	// 日志
	_ = os.Remove("./log.txt")
	file, err := os.OpenFile("./log.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil{
		panic(err)
	}
	LogMgr = log.NewToFile(file)
	_ = LogMgr.WriteInfoLog("app running...")
	// 根目录
	path, err := os.Getwd()
	if err != nil{
		panic(err)
	}
	RootPath = path
}