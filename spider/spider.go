package spider

import (
	"TorrentPage/config"
	"github.com/kkkunny/GoMy/container/queue"
	"github.com/kkkunny/GoMy/worker"
	"sync"
	"time"
)

// 搜索页
type SearchPage map[string]string

// 详情页
type DetailPage map[string]string

// 种子
type Torrent struct {
	Title  string // 标题
	Size   string // 大小
	Magnet string // 磁链
	Src    string // 原网址
}

// 爬虫接口
type Spider interface {
	Search(key string) ([]string, error)   // 获取搜索页
	Fetch(url string) ([]string, error)    // 搜索，获取详情页
	FetchOne(url string) (*Torrent, error) // 获取单页
}

func NewSpiderManager() *SpiderManager {
	var sm SpiderManager
	//s1 := NewSpider0Mag()
	//s2 := NewSpiderBitcq()
	//s3 := NewSpiderZooqle()
	//s4 := NewSpiderCilixingqiu()
	s5 := NewSpiderZhongziso()
	sm.Spiders = map[string]Spider{
		//s1.Name: s1,
		//s2.Name: s2,
		//s3.Name: s3,
		//s4.Name: s4,
		s5.Name: s5,
	}
	return &sm
}

// 爬虫管理器
type SpiderManager struct {
	Spiders map[string]Spider // 爬虫
	Upload  chan *Torrent
}

// 搜索
func (this *SpiderManager) Search(key string) {
	defer func() {
		recover()
	}()
	ch := this.Upload
	wait := sync.WaitGroup{}
	for _, spider := range this.Spiders {
		wait.Add(1)
		fun := func(s Spider) {
			defer wait.Done()
			this.searchOneSpider(s, key)
		}
		go fun(spider)
	}
	wait.Wait()
	time.Sleep(1)
	close(ch)
}

// 搜索一个爬虫
func (this *SpiderManager) searchOneSpider(spider Spider, key string) {
	// 获取搜索页
	var searchUrlQueue = queue.New()
	searchUrls, err := spider.Search(key)
	if err != nil {
		_ = config.LogMgr.WriteErrorLog(err.Error())
		return
	}
	for _, searchUrl := range searchUrls {
		searchUrlQueue.Put(searchUrl)
	}
	// 获取详情页
	var detailUrlQueue = queue.New()
	fun1 := func() bool {
		urlObj := searchUrlQueue.Get()
		if urlObj == nil {
			return false
		}
		detailUrls, err := spider.Fetch(urlObj.(string))
		if err == nil {
			for _, detailUrl := range detailUrls {
				detailUrlQueue.Put(detailUrl)
			}
		} else {
			_ = config.LogMgr.WriteErrorLog(err.Error())
		}
		return true
	}
	searchWorker := worker.NewWorker(config.SpiderNum, fun1, true)
	searchWorker.Start()
	searchWorker.Wait()
	// 获取种子
	fun2 := func() bool {
		urlObj := detailUrlQueue.Get()
		if urlObj == nil {
			return false
		}
		torrent, err := spider.FetchOne(urlObj.(string))
		if err == nil && torrent != nil {
			if closed := SafeSendTorrent(this.Upload, torrent); closed {
				return false
			}
		} else if err != nil {
			_ = config.LogMgr.WriteErrorLog(err.Error())
		}
		return true
	}
	detailWorker := worker.NewWorker(config.SpiderNum, fun2, true)
	detailWorker.Start()
	detailWorker.Wait()
}
