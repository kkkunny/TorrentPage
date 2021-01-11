package spider

import (
	"TorrentPage/config"
	"github.com/kkkunny/GoMy/container/queue"
	"sync"
	"time"
)

// 搜索页
type SearchPage map[string]string
// 详情页
type DetailPage map[string]string

// 种子
type Torrent struct {
	Title string  // 标题
	Size string  // 大小
	Magnet string  // 磁链
	Src string  // 原网址
}

// 爬虫接口
type Spider interface {
	Search(key string)error  // 获取搜索页
	Fetch(url string)error  // 搜索，获取详情页
	FetchOne(url string)(*Torrent, error)  // 获取单页
}

func NewSpiderManager()*SpiderManager{
	sm := &SpiderManager{
		SearchTask: queue.New(),
		DetailTask: queue.New(),
	}
	s1 := NewSpider0Mag(sm.SearchTask, sm.DetailTask)
	s2 := NewSpiderBitcq(sm.SearchTask, sm.DetailTask)
	s3 := NewSpiderZooqle(sm.SearchTask, sm.DetailTask)
	s4 := NewSpiderCilixingqiu(sm.SearchTask, sm.DetailTask)
	sm.Spiders = map[string]Spider{
		s1.Name: s1,
		s2.Name: s2,
		s3.Name: s3,
		s4.Name: s4,
	}
	return sm
}
// 爬虫管理器
type SpiderManager struct {
	Spiders    map[string]Spider // 爬虫
	SearchTask *queue.Queue
	DetailTask *queue.Queue
	Upload chan *Torrent
}
// 搜索
func (this *SpiderManager)Search(key string){
	defer func() {
		recover()
	}()
	ch := this.Upload
	wait := sync.WaitGroup{}
	// 获取搜索页
	for _, sp := range this.Spiders {
		wait.Add(1)
		fun := func(spider Spider) {
			defer wait.Done()
			if err := spider.Search(key); err != nil{
				_ = config.LogMgr.WriteErrorLog(err.Error())
			}
		}
		go fun(sp)
	}
	wait.Wait()
	// 获取详情页
	for i:=0; i<config.SpiderNum; i++{
		wait.Add(1)
		fun := func() {
			defer wait.Done()
			for{
				searchUrlObj := this.SearchTask.Get()
				if searchUrlObj == nil{
					break
				}
				searchUrl := searchUrlObj.(SearchPage)
				for k, url := range searchUrl{
					if err := this.Spiders[k].Fetch(url); err != nil{
						_ = config.LogMgr.WriteErrorLog(err.Error())
						break
					}
				}
			}
		}
		go fun()
	}
	wait.Wait()
	// 搜索详情页
	for i:=0; i<config.SpiderNum; i++{
		wait.Add(1)
		fun := func() {
			defer wait.Done()
			for{
				detailUrlObj := this.DetailTask.Get()
				if detailUrlObj == nil{
					break
				}
				detailUrl := detailUrlObj.(DetailPage)
				for k, url := range detailUrl{
					t, err := this.Spiders[k].FetchOne(url)
					if err != nil {
						_ = config.LogMgr.WriteErrorLog(err.Error())
						break
					}
					if t == nil{
						continue
					}
					if closed := SafeSendTorrent(ch, t); closed{
						return
					}
				}
			}
		}
		go fun()
	}
	wait.Wait()
	time.Sleep(1)
	close(ch)
}