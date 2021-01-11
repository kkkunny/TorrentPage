package spider

import (
	"github.com/kkkunny/GoMy/container/queue"
	"github.com/kkkunny/GoMy/http/requests"
	"github.com/kkkunny/GoMy/re"
	"strings"
)

func NewSpider0Mag(searchTask, detailTask *queue.Queue)*Spider0Mag{
	return &Spider0Mag{
		Name: "0mag",
		Domain: "https://0mag.net",
		SearchTask: searchTask,
		DetailTask: detailTask,
	}
}
type Spider0Mag struct {
	Name string
	Domain string
	SearchTask *queue.Queue
	DetailTask *queue.Queue
}
func (this *Spider0Mag)Search(key string)error{
	url := this.Domain + "/search?q="+key
	this.SearchTask.Put(SearchPage{this.Name: url})
	return nil
}
func (this *Spider0Mag)Fetch(url string)error{
	req := requests.NewRequest()
	response, err := req.Get(url, nil)
	if err != nil{
		return err
	}
	text := response.Text()
	if strings.Count(text, "0 results") > 0{
		return nil
	}
	hrefss := re.FindAll(`<td>\s*<a href="(.+?)">`, text)
	if len(hrefss) > 0{
		for _, hrefs := range hrefss{
			detailUrl := this.Domain + hrefs[1]
			this.DetailTask.Put(DetailPage{this.Name: detailUrl})
		}
	}
	return nil
}
func (this *Spider0Mag)FetchOne(url string)(*Torrent, error){
	req := requests.NewRequest()
	response, err := req.Get(url, nil)
	if err != nil{
		return nil, err
	}
	text := response.Text()
	titles := re.FindAll(`<h2 class="magnet-title">(.+?)</h2>`, text)
	if len(titles) == 0{
		return nil, nil
	}
	title := titles[0][1]
	magnets := re.FindAll(`id="input-magnet" class="form-control" type="text" value="(.+?)" spellcheck="false"`, text)
	if len(magnets) == 0{
		return nil, nil
	}
	magnet := magnets[0][1]
	if title == "" || magnet == ""{
		return nil, nil
	}
	var result = Torrent{Title: title, Magnet: magnet, Src: url}
	sizes := re.FindAll(`<dt>Content Size :</dt> <dd>(.+?)</dd>`, text)
	if len(sizes) > 0{
		size := sizes[0][1]
		result.Size = size
	}
	return &result, nil
}