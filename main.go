package main

import (
	"TorrentPage/config"
	"TorrentPage/ui"
)

func main() {
	app := ui.App{}
	if err := app.Init(); err != nil{
		_ = config.LogMgr.WriteErrorLog(err.Error())
		panic(err)
	}
}