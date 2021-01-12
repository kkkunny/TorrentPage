package spider

import (
	"fmt"
	"github.com/kkkunny/GoMy/http/requests"
	"github.com/kkkunny/GoMy/re"
	"strconv"
	"strings"
)

func NewSpiderZooqle() *SpiderZooqle {
	return &SpiderZooqle{
		Name:   "Zooqle",
		Domain: "https://zooqle.com",
	}
}

type SpiderZooqle struct {
	Name   string
	Domain string
}

func (this *SpiderZooqle) Search(key string) ([]string, error) {
	var urls []string
	url := this.Domain + "/search?q=" + key
	req := requests.NewRequest()
	response, err := req.Get(url, nil)
	if err != nil {
		return urls, err
	}
	pages := re.FindAll(`><a href="\?pg=(.+?)&q=`+key+`&v=t">`, response.Text())
	if len(pages) > 0 {
		pageStr := pages[len(pages)-1][1]
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			return urls, err
		}
		for i := 1; i <= page; i++ {
			urls = append(urls, fmt.Sprintf("%ssearch?pg=%d&q=%s&v=t", this.Domain, i, key))
		}
	}
	return urls, nil
}
func (this *SpiderZooqle) Fetch(url string) ([]string, error) {
	var urls []string
	req := requests.NewRequest()
	response, err := req.Get(url, nil)
	if err != nil {
		return urls, err
	}
	text := response.Text()
	if strings.Count(text, "Sorry, no torrents match your query.") > 0 {
		return urls, nil
	}
	hrefss := re.FindAll(`small" href="(.+?)">`, text)
	if len(hrefss) > 0 {
		for _, hrefs := range hrefss {
			urls = append(urls, this.Domain+hrefs[1])
		}
	}
	return urls, nil
}
func (this *SpiderZooqle) FetchOne(url string) (*Torrent, error) {
	req := requests.NewRequest()
	response, err := req.Get(url, nil)
	if err != nil {
		return nil, err
	}
	text := response.Text()
	titles := re.FindAll(`<h4 id="torname">(.+?)<span`, text)
	if len(titles) == 0 {
		return nil, nil
	}
	title := titles[0][1]
	magnets := re.FindAll(`rel="nofollow" href="(.+?)"><i`, text)
	if len(magnets) == 0 {
		return nil, nil
	}
	magnet := magnets[0][1]
	if title == "" || magnet == "" {
		return nil, nil
	}
	var result = Torrent{Title: title, Magnet: magnet, Src: url}
	sizes := re.FindAll(`File size"></i>(.+?)<span class="small pad-l2">`, text)
	if len(sizes) > 0 {
		size := sizes[0][1]
		result.Size = size
	}
	return &result, nil
}
