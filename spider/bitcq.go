package spider

import (
	"fmt"
	"github.com/kkkunny/GoMy/http/requests"
	"github.com/kkkunny/GoMy/re"
	"strconv"
	"strings"
)

func NewSpiderBitcq() *SpiderBitcq {
	return &SpiderBitcq{
		Name:   "Bitcq",
		Domain: "https://bitcq.com",
	}
}

type SpiderBitcq struct {
	Name   string
	Domain string
}

func (this *SpiderBitcq) Search(key string) ([]string, error) {
	var urls []string
	url := this.Domain + "/search?q=" + key
	req := requests.NewRequest()
	response, err := req.Get(url, nil)
	if err != nil {
		return urls, err
	}
	pages := re.FindAll(`href='/search\?q=`+key+`&page=(.+?)'`, response.Text())
	if len(pages) > 0 {
		pageStr := pages[len(pages)-1][1]
		//pageStr := pages[0][1]
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			return urls, err
		}
		for i := 1; i <= page; i++ {
			urls = append(urls, fmt.Sprintf("%s&page=%d", url, i))
		}
	}
	return urls, nil
}
func (this *SpiderBitcq) Fetch(url string) ([]string, error) {
	var urls []string
	req := requests.NewRequest()
	response, err := req.Get(url, nil)
	if err != nil {
		return urls, err
	}
	text := response.Text()
	if strings.Count(text, "About 0 results") > 0 {
		return urls, nil
	}
	hrefss := re.FindAll(`</td>\s*<td>\s*<a href="(.+?)">`, text)
	if len(hrefss) > 0 {
		for _, hrefs := range hrefss {
			urls = append(urls, this.Domain+hrefs[1])
		}
	}
	return urls, nil
}
func (this *SpiderBitcq) FetchOne(url string) (*Torrent, error) {
	req := requests.NewRequest()
	response, err := req.Get(url, nil)
	if err != nil {
		return nil, err
	}
	text := response.Text()
	titles := re.FindAll(`<meta property="og:title" content="(.+?)" />`, text)
	if len(titles) == 0 {
		return nil, nil
	}
	title := titles[0][1]
	magnets := re.FindAll(`<a rel="nofollow" href="(.+?)" target="_blank" class="btn btn-default btn-lg">`, text)
	if len(magnets) == 0 {
		return nil, nil
	}
	magnet := magnets[0][1]
	if title == "" || magnet == "" {
		return nil, nil
	}
	var result = Torrent{Title: title, Magnet: magnet, Src: url}
	sizes := re.FindAll(`<h5>\s*Size:\s*(.+?)\s*</h5>`, text)
	if len(sizes) > 0 {
		size := sizes[0][1]
		result.Size = size
	}
	return &result, nil
}
