package spider

import (
	"fmt"
	"github.com/kkkunny/GoMy/http/requests"
	"github.com/kkkunny/GoMy/re"
	"strconv"
	"strings"
)

func NewSpiderCilixingqiu() *SpiderCilixingqiu {
	return &SpiderCilixingqiu{
		Name:   "Cilixingqiu",
		Domain: "https://cilixingqiu.co",
	}
}

type SpiderCilixingqiu struct {
	Name   string
	Domain string
}

func (this *SpiderCilixingqiu) Search(key string) ([]string, error) {
	var urls []string
	url := this.Domain + "/search?q=" + key
	req := requests.NewRequest()
	response, err := req.Get(url, nil)
	if err != nil {
		return urls, err
	}
	pages := re.FindAll(`<a href="/s/`+key+`/p/(.+?)" class="page">尾页`, response.Text())
	if len(pages) > 0 {
		pageStr := pages[0][1]
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			return urls, err
		}
		for i := 1; i <= page; i++ {
			urls = append(urls, fmt.Sprintf("%s/s/%s/p/%d", this.Domain, key, i))
		}
	} else {
		urls = append(urls, this.Domain+"/s/"+key)
	}
	return urls, nil
}
func (this *SpiderCilixingqiu) Fetch(url string) ([]string, error) {
	var urls []string
	req := requests.NewRequest()
	response, err := req.Get(url, nil)
	if err != nil {
		return urls, err
	}
	text := response.Text()
	if strings.Count(text, "相关链接") > 0 {
		return urls, nil
	}
	hrefss := re.FindAll(`<a href="(.+?)" title="`, text)
	if len(hrefss) > 0 {
		for _, hrefs := range hrefss {
			urls = append(urls, this.Domain+hrefs[1])
		}
	}
	return urls, nil
}
func (this *SpiderCilixingqiu) FetchOne(url string) (*Torrent, error) {
	req := requests.NewRequest()
	response, err := req.Get(url, nil)
	if err != nil {
		return nil, err
	}
	text := response.Text()
	titles := re.FindAll(`class="article-title">(.+?)</h1>`, text)
	if len(titles) == 0 {
		return nil, nil
	}
	title := titles[0][1]
	magnets := re.FindAll(`class="blue-color" href="(.+?)">`, text)
	if len(magnets) == 0 {
		return nil, nil
	}
	magnet := magnets[0][1]
	if title == "" || magnet == "" {
		return nil, nil
	}
	var result = Torrent{Title: title, Magnet: magnet, Src: url}
	sizes := re.FindAll(`文件大小：<strong>(.+?)</strong><br>`, text)
	if len(sizes) > 0 {
		size := sizes[0][1]
		result.Size = size
	}
	return &result, nil
}
