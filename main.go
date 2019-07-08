package main

import (
	"fmt"
	httpDownloader "github.com/Mrs4s/go-http-downloader"
	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
	"github.com/whatl3y/argv"
	"onedl/models"
	"onedl/onedrive"
	"onedl/view"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	Input       string
	OutputDir   string
	ThreadCount int

	currentPath string
)

func initCli() {
	//onedl url -i test.file -o path -tc 16
	args := argv.Parse(os.Args[2:])
	Input = args.Keys["i"]
	OutputDir = args.Keys["o"]
	threadCountStr, ext := args.Keys["tc"]
	if ext {
		i, err := strconv.Atoi(threadCountStr)
		if err != nil {
			fmt.Println(threadCountStr + " cannot parse to int.")
			return
		}
		ThreadCount = i
	} else {
		ThreadCount = 16
	}
	fmt.Println("[one] Initializing...")
	var shared onedrive.IShared
	if strings.Contains(os.Args[1], "sharepoint") {
		shared = &onedrive.SharePointShared{
			Url: os.Args[1],
		}
	}
	if shared == nil {
		fmt.Println("[one] Unknown url.")
		return
	}
	if !shared.Init() {
		fmt.Println("[one] Cannot initialize shared info.")
		return
	}
	parentFiles := shared.GetFilesByPath(models.GetParentPath(Input))
	if parentFiles == nil {
		fmt.Println("[one] Get parent files failed.")
		return
	}

	var targetFile onedrive.SharedFile
	for _, file := range parentFiles {
		if file.Name == models.GetFileName(Input) {
			targetFile = file
		}
	}
	if targetFile.Owner == nil {
		fmt.Println("[one] Cannot find target file.")
		return
	}
	for key, file := range targetFile.GetLocalTree(OutputDir) {
		fmt.Println("[one] Create download task " + key)
		info, err := httpDownloader.NewDownloaderInfo([]string{file.GetDownloadLink(file)}, key, 0, ThreadCount, map[string]string{"User-Agent": "Mozilla/5.0 (Windows NT 10.0; WOW64; rv:67.0) Gecko/20100101 Firefox/67.0"})
		if err != nil {
			fmt.Println("[one] Cannot create download task " + key)
			continue
		}
		downloader := httpDownloader.NewClient(info)
		ch := make(chan bool)
		err = downloader.BeginDownload()
		if err != nil {
			fmt.Println("[one] Cannot start download " + downloader.Info.TargetFile + ", error: " + err.Error())
			continue
		}
		fmt.Println("[one] Start download " + downloader.Info.TargetFile)
		p := mpb.New(mpb.WithWidth(60), mpb.WithRefreshRate(1000*time.Millisecond))
		bar := p.AddBar(downloader.Info.ContentSize, mpb.BarStyle("[=>-|"),
			mpb.PrependDecorators(
				decor.CountersKibiByte("% 6.1f / % 6.1f"),
			),
			mpb.AppendDecorators(
				decor.Name(" ] "),
				decor.AverageSpeed(decor.UnitKiB, "% .2f"),
			),
		)
		downloader.OnCompleted(func() {
			bar.SetTotal(downloader.DownloadedSize, true)
			time.Sleep(time.Duration(1) * time.Second)
			fmt.Println("[one] Download task " + downloader.Info.TargetFile + " completed.")
			ch <- true
		})
		downloader.OnFailed(func(err error) {
			bar.SetTotal(downloader.DownloadedSize, true)
			time.Sleep(time.Duration(1) * time.Second)
			fmt.Println("[one] Download task " + downloader.Info.TargetFile + " failed: " + err.Error())
			ch <- false
		})
		go func() {
			for downloader.Downloading {
				bar.SetCurrent(downloader.DownloadedSize)
				time.Sleep(time.Duration(1) * time.Second)
			}
		}()
		<-ch
		p.Wait()
	}
}

func main() {
	switch {
	case len(os.Args) == 1:
		fmt.Println("Please input onedrive share url.")
		return
	case len(os.Args) == 2:
		//show gui
		if err := view.InitUi(); err != nil {
			fmt.Println("[one] Cannot create cui: " + err.Error())
			return
		}
		go func() {
			view.AppendLog("Initializing...")
			OutputDir = "Downloads"
			var shared onedrive.IShared
			if strings.Contains(os.Args[1], "sharepoint") {
				shared = &onedrive.SharePointShared{
					Url: os.Args[1],
				}
			}
			if shared == nil {
				view.AppendLog("Unknown url.")
				return
			}
			if !shared.Init() {
				view.AppendLog("Init failed.")
				return
			}
			view.SetFiles(shared.GetRootFiles(), 0)
			currentPath = "/"
			view.OnEnterFolder(func(file onedrive.SharedFile) {
				if file.IsDir {
					view.ShowLoading()
					view.SetFiles(file.GetChildren(), 0)
					view.SetPath(file.Path)
					currentPath = file.Path
					view.HideLoading()
				}
			})
			view.OnEnterPrevious(func() {
				if currentPath != "/" {
					view.ShowLoading()
					path := models.GetParentPath(currentPath)
					view.SetFiles(shared.GetFilesByPath(path), 0)
					view.SetPath(path)
					currentPath = path
					view.HideLoading()
				}
			})
			view.OnDownloadFile(func(file onedrive.SharedFile) {
				view.ShowLoading()
				for key, file := range file.GetLocalTree(OutputDir) {
					info, err := httpDownloader.NewDownloaderInfo([]string{file.GetDownloadLink(file)}, key, 0, 16, map[string]string{"User-Agent": "Mozilla/5.0 (Windows NT 10.0; WOW64; rv:67.0) Gecko/20100101 Firefox/67.0"})
					if err != nil {
						view.AppendLog("Cannot create task " + key + ": " + err.Error())
						continue
					}
					view.DownloaderList = append(view.DownloaderList, httpDownloader.NewClient(info))
				}
				view.HideLoading()
			})
		}()
		view.BeginLoop()
	default:
		initCli()
	}
}
