package view

import (
	httpDownloader "github.com/Mrs4s/go-http-downloader"
	ui "github.com/gizak/termui/v3"
	"onedl/onedrive"
	"time"
)

var (
	DownloaderList []*httpDownloader.DownloaderClient

	fileList []onedrive.SharedFile
	page     int

	viewFileList []onedrive.SharedFile
	lightIndex   int
	loading      bool

	onEnter         func(file onedrive.SharedFile)
	onDownload      func(file onedrive.SharedFile)
	onEnterPrevious func()
)

const welcomeMessage = "" +
	"Welcome to use onedl! \n" +
	"Default download folder is <WorkDir>/Downloads\n" +
	"Press [Tab] to switch tab. \n" +
	"Press [->] or [<-] to up/down page.\n" +
	"Press [Enter] to join the folder.\n" +
	"Press [Space] to download file or dir.\n" +
	"Press [Ctrl+C] to quit. \n"

func InitUi() error {
	if err := ui.Init(); err != nil {
		return err
	}
	setupViews()
	return nil
}

func SetFiles(files []onedrive.SharedFile, pg int) {
	page = pg
	fileList = files
	defer RedrawUi()
	_, height := ui.TerminalDimensions()
	if len(files) < height-13 {
		viewFileList = files
		return
	}
	begin := page * (height - 13)
	end := (page + 1) * (height - 13)
	if begin > len(files) {
		page--
		return
	}
	if end < len(files) {
		viewFileList = files[begin:end]
		return
	}
	viewFileList = files[begin:]
}

func ShowLoading() {
	loading = true
	RedrawUi()
}

func HideLoading() {
	loading = false
	RedrawUi()
}

func handleKeys(key string) {
	if !loading {
		switch key {
		case "<Tab>":
			if tab.ActiveTabIndex == 2 {
				tab.ActiveTabIndex = 0
			} else {
				tab.FocusRight()
			}
			RedrawUi()
		case "<Right>":
			lightIndex = 0
			page++
			RefreshUi()
		case "<Left>":
			if page > 0 {
				lightIndex = 0
				page--
			}
			RefreshUi()
		case "<Up>":
			if lightIndex > 0 {
				lightIndex--
			}
			RefreshUi()
		case "<Down>":
			if lightIndex < len(viewFileList) {
				lightIndex++
			}
			RefreshUi()
		case "<Enter>":
			if lightIndex > 0 {
				go onEnter(viewFileList[lightIndex-1])
			} else {
				go onEnterPrevious()
			}
		case "<Space>":
			if lightIndex > 0 {
				go onDownload(viewFileList[lightIndex-1])
			}
		default:
			//AppendLog(key)
		}
	}
}

func RedrawUi() {
	width, height := ui.TerminalDimensions()
	drawViews(ui.Resize{
		Width:  width,
		Height: height,
	})
}

func RefreshUi() {
	SetFiles(fileList, page)
	RedrawUi()
}

func BeginLoop() {
	uiEvents := ui.PollEvents()
	ticker := time.NewTicker(time.Second).C
	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "<C-c>":
				ui.Close()
				return
			case "<Resize>":
				ui.Clear()
				//AppendLog(strconv.FormatInt(int64(e.Payload.(ui.Resize).Height), 10) + ":" + strconv.FormatInt(int64(e.Payload.(ui.Resize).Width), 10))
				RefreshUi()
				//drawViews(e.Payload.(ui.Resize))
			default:
				handleKeys(e.ID)
			}
		case <-ticker:
			if len(DownloaderList) > 0 {
				var downloadingCount int
				waitingTask := -1
				for i, task := range DownloaderList {
					if task.Downloading {
						downloadingCount++
					}
					if waitingTask == -1 && !task.Downloading {
						waitingTask = i
					}
				}
				if downloadingCount < 2 && waitingTask != -1 {
					task := DownloaderList[waitingTask]
					err := task.BeginDownload()
					if err != nil {
						AppendLog("Cannot begin task " + task.Info.TargetFile)
						DownloaderList = append(DownloaderList[:waitingTask], DownloaderList[waitingTask+1:]...)
					}
					task.OnCompleted(func() {
						AppendLog("Task " + task.Info.TargetFile + " completed.")
						DownloaderList = append(DownloaderList[:waitingTask], DownloaderList[waitingTask+1:]...)
					})
					task.OnFailed(func(err error) {
						AppendLog("Task " + task.Info.TargetFile + " failed: " + err.Error())
						DownloaderList = append(DownloaderList[:waitingTask], DownloaderList[waitingTask+1:]...)
					})
				}
			}
			if tab.ActiveTabIndex == 2 {
				RedrawUi()
			}
		}

	}
}

func OnEnterFolder(fc func(file onedrive.SharedFile)) {
	onEnter = fc
}

func OnDownloadFile(fc func(file onedrive.SharedFile)) {
	onDownload = fc
}

func OnEnterPrevious(fc func()) {
	onEnterPrevious = fc
}
