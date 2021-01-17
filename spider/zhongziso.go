package spider

import (
	"fmt"
	"github.com/kkkunny/GoMy/http/requests"
	"github.com/kkkunny/GoMy/re"
	"strconv"
)

func NewSpiderZhongziso() *SpiderZhongziso {
	return &SpiderZhongziso{
		Name:   "Zhongziso",
		Domain: "https://zhongziso88.xyz",
	}
}

type SpiderZhongziso struct {
	Name   string
	Domain string
}

func (this *SpiderZhongziso) Search(key string) ([]string, error) {
	var urls []string
	url := fmt.Sprintf("%s/list/%s/1", this.Domain, key)
	req := requests.NewRequest()
	response, err := req.Get(url, nil)
	if err != nil {
		return urls, err
	}
	pages := re.FindAll(`<a href="/list_ctime/`+key+`/(\d+?)">`, response.Text())
	if len(pages) > 0 {
		pageStr := pages[len(pages)-1][1]
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			return urls, err
		}
		for i := 1; i <= page; i++ {
			urls = append(urls, fmt.Sprintf("%s/list_ctime/%s/%d", this.Domain, key, i))
		}
	} else {
		urls = append(urls, this.Domain+"/s/"+key)
	}
	return urls, nil
}
func (this *SpiderZhongziso) Fetch(url string) ([]string, error) {
	var urls []string
	req := requests.NewRequest()
	response, err := req.Get(url, nil)
	if err != nil {
		return urls, err
	}
	text := response.Text()
	hrefss := re.FindAll(`</span>&nbsp;&nbsp;&nbsp;&nbsp;<a href="(.+?)"><span class="highlight">`, text)
	if len(hrefss) > 0 {
		for _, hrefs := range hrefss {
			urls = append(urls, this.Domain+hrefs[1])
		}
	}
	return urls, nil
}
func (this *SpiderZhongziso) FetchOne(url string) (*Torrent, error) {
	req := requests.NewRequest()
	response, err := req.Get(url, nil)
	if err != nil {
		return nil, err
	}
	text := response.Text()
	titles := re.FindAll(`<div class="text-left">\s*(.+?)\s*</div>`, text)
	if len(titles) == 0 {
		return nil, nil
	}
	title := titles[0][1]
	magnets := re.FindAll(`readonly="readonly">\s*(.+?)\s*</textarea>`, text)
	if len(magnets) == 0 {
		return nil, nil
	}
	magnet := magnets[0][1]
	if title == "" || magnet == "" {
		return nil, nil
	}
	var result = Torrent{Title: title, Magnet: magnet, Src: url}
	sizes := re.FindAll(`文件大小:</dt>\s*<dd class="text-left">(.+?)</dd>`, text)
	if len(sizes) > 0 {
		size := sizes[0][1]
		result.Size = size
	}
	return &result, nil
}
