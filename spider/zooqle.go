package spider

import (
	"fmt"
	"github.com/kkkunny/GoMy/container/queue"
	"github.com/kkkunny/GoMy/http/requests"
	"github.com/kkkunny/GoMy/re"
	"strconv"
	"strings"
)

func NewSpiderZooqle(searchTask, detailTask *queue.Queue)*SpiderZooqle{
	return &SpiderZooqle{
		Name: "Zooqle",
		Domain: "https://zooqle.com/",
		SearchTask: searchTask,
		DetailTask: detailTask,
	}
}
type SpiderZooqle struct {
	Name string
	Domain string
	SearchTask *queue.Queue
	DetailTask *queue.Queue
}
func (this *SpiderZooqle)Search(key string)error{
	url := this.Domain + "/search?q="+key
	req := requests.NewRequest()
	response, err := req.Get(url, nil)
	if err != nil{
		return err
	}
	pages := re.FindAll(`><a href="\?pg=(.+?)&q=`+key+`&v=t">`, response.Text())
	if len(pages) > 0{
		pageStr := pages[len(pages)-1][1]
		page, err := strconv.Atoi(pageStr)
		if err != nil{
			return err
		}
		for i:=1; i<=page; i++{
			this.SearchTask.Put(SearchPage{this.Name: fmt.Sprintf("%ssearch?pg=%d&q=%s&v=t", this.Domain, i, key)})
		}
	}
	return nil
}
func (this *SpiderZooqle)Fetch(url string)error{
	req := requests.NewRequest()
	response, err := req.Get(url, nil)
	if err != nil{
		return err
	}
	text := response.Text()
	if strings.Count(text, "Sorry, no torrents match your query.") > 0{
		return nil
	}
	hrefss := re.FindAll(`small" href="(.+?)">`, text)
	if len(hrefss) > 0{
		for _, hrefs := range hrefss{
			detailUrl := this.Domain + hrefs[1]
			this.DetailTask.Put(DetailPage{this.Name: detailUrl})
		}
	}
	return nil
}
func (this *SpiderZooqle)FetchOne(url string)(*Torrent, error){
	req := requests.NewRequest()
	response, err := req.Get(url, nil)
	if err != nil{
		return nil, err
	}
	text := response.Text()
	titles := re.FindAll(`<h4 id="torname">(.+?)<span`, text)
	if len(titles) == 0{
		return nil, nil
	}
	title := titles[0][1]
	magnets := re.FindAll(`rel="nofollow" href="(.+?)"><i`, text)
	if len(magnets) == 0{
		return nil, nil
	}
	magnet := magnets[0][1]
	if title == "" || magnet == ""{
		return nil, nil
	}
	var result = Torrent{Title: title, Magnet: magnet, Src: url}
	sizes := re.FindAll(`File size"></i>(.+?)<span class="small pad-l2">`, text)
	if len(sizes) > 0{
		size := sizes[0][1]
		result.Size = size
	}
	return &result, nil
}