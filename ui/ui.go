package ui

import (
	"TorrentPage/config"
	"TorrentPage/spider"
	"errors"
	"fmt"
	"github.com/atotto/clipboard"
	"github.com/kkkunny/GoMy/container/linklist"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/uitools"
	"github.com/therecipe/qt/widgets"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
	"unsafe"
)

// 主ui
type App struct {
	window       *widgets.QMainWindow
	searchInput  *widgets.QLineEdit    // 输入框
	searchButton *widgets.QPushButton  // 搜索按钮
	displayTable *widgets.QTableWidget // 展示table
	data         *linklist.LinkList    // 磁力数据
	waitLabel    *widgets.QLabel       // 等待展示
	status       bool                  // 当前状态 true 爬虫运行中 false 暂停中
	statusImage  []*gui.QPixmap        // 状态
	spiderMgr    *spider.SpiderManager // 爬虫管理器
}

// 初始化应用
func (this *App) Init() error {
	widgets.NewQApplication(len(os.Args), os.Args)
	// 载入ui
	mainUi := core.NewQFile2(filepath.Join(filepath.Join(config.RootPath, "static"), "main.ui"))
	mainUi.Open(core.QIODevice__ReadOnly)
	loader := uitools.NewQUiLoader(nil)
	windowsObj := loader.Load(mainUi, nil)
	mainUi.Close()
	window := (*widgets.QMainWindow)(unsafe.Pointer(windowsObj))
	this.window = window
	this.window.SetWindowIcon(gui.NewQIcon5(filepath.Join(filepath.Join(config.RootPath, "static"), "icon.png")))
	// 各个控件
	inputObj := window.FindChild("searchInput", core.Qt__FindChildrenRecursively)
	if name := inputObj.ObjectName(); name == "" {
		return errors.New("no find QLineEdit widget")
	}
	this.searchInput = (*widgets.QLineEdit)(unsafe.Pointer(inputObj))
	buttonObj := window.FindChild("searchButton", core.Qt__FindChildrenRecursively)
	if name := buttonObj.ObjectName(); name == "" {
		return errors.New("no find QPushButton widget")
	}
	this.searchButton = (*widgets.QPushButton)(unsafe.Pointer(buttonObj))
	this.searchButton.SetIcon(gui.NewQIcon5(filepath.Join(filepath.Join(config.RootPath, "static"), "search.png")))
	this.searchButton.ConnectClicked(this.ButtonClicked)
	tableObj := window.FindChild("displayTable", core.Qt__FindChildrenRecursively)
	if name := tableObj.ObjectName(); name == "" {
		return errors.New("no find QTableWidget widget")
	}
	this.displayTable = (*widgets.QTableWidget)(unsafe.Pointer(tableObj))
	this.displayTable.SetColumnWidth(0, 100)
	this.displayTable.SetColumnWidth(1, 820)
	this.displayTable.SetColumnWidth(2, 123)
	this.displayTable.HorizontalHeader().SetSectionResizeMode(widgets.QHeaderView__Fixed)
	waitObj := window.FindChild("waitLabel", core.Qt__FindChildrenRecursively)
	if name := waitObj.ObjectName(); name == "" {
		return errors.New("no find QLabel widget")
	}
	this.waitLabel = (*widgets.QLabel)(unsafe.Pointer(waitObj))
	this.statusImage = append(this.statusImage, gui.NewQPixmap3(filepath.Join(filepath.Join(config.RootPath, "static"), "ok.png"), "png", core.Qt__NoFormatConversion))
	this.statusImage = append(this.statusImage, gui.NewQPixmap3(filepath.Join(filepath.Join(config.RootPath, "static"), "wait.png"), "png", core.Qt__NoFormatConversion))
	this.SetStopping()
	// 右键菜单
	rightMenu := widgets.NewQMenu(this.displayTable)
	copyMagnet := widgets.NewQAction2("复制磁链", rightMenu)
	copyMagnet.ConnectTriggered(this.CopyMagnet)
	copySrc := widgets.NewQAction2("复制原网页", rightMenu)
	copySrc.ConnectTriggered(this.CopySrc)
	rightMenu.AddActions([]*widgets.QAction{copyMagnet, copySrc})
	this.displayTable.ConnectCustomContextMenuRequested(func(pos *core.QPoint) {
		point := this.displayTable.MapToGlobal(pos)
		rightMenu.Move(point)
		rightMenu.Exec()
	})
	// 爬虫
	this.spiderMgr = spider.NewSpiderManager()
	// 运行
	this.window.Show()
	widgets.QApplication_Exec()
	return nil
}

// 爬虫运行中
func (this *App) SetRunning() {
	this.status = true
	this.waitLabel.SetPixmap(this.statusImage[1])
	go this.WaitImageTransform()
}

// 爬虫结束
func (this *App) SetStopping() {
	this.status = false
	this.waitLabel.SetPixmap(this.statusImage[0])
}

// 按钮按下
func (this *App) ButtonClicked(bool) {
	defer func() {
		if r := recover(); r != nil {
			this.spiderMgr.Upload = nil
			this.ButtonClicked(false)
		}
	}()
	if this.spiderMgr.Upload != nil {
		close(this.spiderMgr.Upload)
	}
	this.ClearTorrent()
	go this.Search()
}

// 搜索
func (this *App) Search() {
	key := this.searchInput.Text()
	if key == "" {
		return
	}
	_ = config.LogMgr.WriteInfoLog("开始搜索：" + key)
	this.SetRunning()
	ch := make(chan *spider.Torrent, 100)
	this.spiderMgr.Upload = ch
	wait := sync.WaitGroup{}
	// 接收
	wait.Add(1)
	fun := func() {
		defer wait.Done()
		for {
			t, ok := <-ch
			if !ok {
				return
			}
			this.AddTorrent(t)
		}
	}
	go fun()
	// 发送
	wait.Add(1)
	fun = func() {
		defer wait.Done()
		this.spiderMgr.Search(key)
	}
	go fun()
	wait.Wait()
	_ = config.LogMgr.WriteInfoLog(fmt.Sprintf("搜索[%s]完毕，共有%d个结果", key, this.data.GetLength()))
	this.SetStopping()
}

// 增加种子到列表
func (this *App) AddTorrent(t *spider.Torrent) {
	if !this.FilterTorrent(t) {
		return
	}
	this.data.Append(t)
	row := this.displayTable.RowCount()
	this.displayTable.SetRowCount(row + 1)
	// 序号
	item := widgets.NewQTableWidgetItem2(strconv.Itoa(row+1), 0)
	this.displayTable.SetItem(row, 0, item)
	// 标题
	item = widgets.NewQTableWidgetItem2(t.Title, 0)
	this.displayTable.SetItem(row, 1, item)
	// 大小
	item = widgets.NewQTableWidgetItem2(t.Size, 0)
	this.displayTable.SetItem(row, 2, item)
}

// 过滤种子
func (this *App) FilterTorrent(t *spider.Torrent) bool {
	if t == nil {
		return false
	}
	var same bool
	this.data.ErgodicFunc(func(i interface{}) {
		to := i.(*spider.Torrent)
		if to.Magnet == t.Magnet {
			same = true
		}
	})
	if same {
		return false
	}
	return true
}

// 清空种子
func (this *App) ClearTorrent() {
	this.data = linklist.New()
	this.displayTable.SetRowCount(0)
}

// 复制磁链
func (this *App) CopyMagnet(bool) {
	row := this.displayTable.CurrentRow()
	torrent := this.data.Get(row).(*spider.Torrent)
	err := clipboard.WriteAll(torrent.Magnet)
	if err != nil {
		_ = config.LogMgr.WriteErrorLog(err.Error())
	}
}

// 复制原网页
func (this *App) CopySrc(bool) {
	row := this.displayTable.CurrentRow()
	torrent := this.data.Get(row).(*spider.Torrent)
	err := clipboard.WriteAll(torrent.Src)
	if err != nil {
		_ = config.LogMgr.WriteErrorLog(err.Error())
	}
}

// 等待图片选在
func (this *App) WaitImageTransform() {
	var curAngle float64
	for this.status { // 运行时旋转
		matrix := gui.NewQMatrix2()
		matrix.Rotate(curAngle)
		this.waitLabel.SetPixmap(this.statusImage[1].Transformed2(matrix, core.Qt__SmoothTransformation))
		curAngle += 1
		if curAngle > 360 {
			curAngle = 0
		}
		time.Sleep(10 * time.Millisecond)
	}
}
