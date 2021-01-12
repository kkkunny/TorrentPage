package spider

import (
	"github.com/kkkunny/GoMy/http/requests"
	"github.com/kkkunny/GoMy/re"
	"strings"
)

func NewSpider0Mag() *Spider0Mag {
	return &Spider0Mag{
		Name:   "0mag",
		Domain: "https://0mag.net",
	}
}

type Spider0Mag struct {
	Name   string
	Domain string
}

func (this *Spider0Mag) Search(key string) ([]string, error) {
	url := this.Domain + "/search?q=" + key
	return []string{url}, nil
}
func (this *Spider0Mag) Fetch(url string) ([]string, error) {
	var urls []string
	req := requests.NewRequest()
	response, err := req.Get(url, nil)
	if err != nil {
		return urls, err
	}
	text := response.Text()
	if strings.Count(text, "0 results") > 0 {
		return urls, nil
	}
	hrefss := re.FindAll(`<td>\s*<a href="(.+?)">`, text)
	if len(hrefss) > 0 {
		for _, hrefs := range hrefss {
			urls = append(urls, this.Domain+hrefs[1])
		}
	}
	return urls, nil
}
func (this *Spider0Mag) FetchOne(url string) (*Torrent, error) {
	req := requests.NewRequest()
	response, err := req.Get(url, nil)
	if err != nil {
		return nil, err
	}
	text := response.Text()
	titles := re.FindAll(`<meta content="(.+?)\s*torrent" name="description"`, text)
	if len(titles) == 0 {
		return nil, nil
	}
	title := titles[0][1]
	magnets := re.FindAll(`id="input-magnet" class="form-control" type="text" value="(.+?)" spellcheck="false"`, text)
	if len(magnets) == 0 {
		return nil, nil
	}
	magnet := magnets[0][1]
	if title == "" || magnet == "" {
		return nil, nil
	}
	var result = Torrent{Title: title, Magnet: magnet, Src: url}
	sizes := re.FindAll(`<dt>Content Size :</dt> <dd>(.+?)</dd>`, text)
	if len(sizes) > 0 {
		size := sizes[0][1]
		result.Size = size
	}
	return &result, nil
}
