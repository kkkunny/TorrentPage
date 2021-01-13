package config

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
)

type loadJson struct {
	SpiderNum int `json:"搜索线程数"`
}

func init() {
	data, err := ioutil.ReadFile(filepath.Join(RootPath, "config.json"))
	if err != nil {
		panic(err)
	}
	var temp loadJson
	err = json.Unmarshal(data, &temp)
	if err != nil {
		panic(err)
	}
	SpiderNum = temp.SpiderNum
}
