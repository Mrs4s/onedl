package view

import (
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"onedl/models"
	"strconv"
)

var (
	header           *widgets.Paragraph
	tab              *widgets.TabPane
	logsView         *widgets.Paragraph
	filesView        *widgets.Table
	loadingView      *widgets.Paragraph
	downloadListView *widgets.Table
)

func setupViews() {
	width, height := ui.TerminalDimensions()

	header = widgets.NewParagraph()
	header.Title = "Path"
	header.Text = "/"

	tab = widgets.NewTabPane("Welcome", "FileExplorer", "DownloadList")
	tab.Border = true

	logsView = widgets.NewParagraph()
	logsView.Title = "Logs"
	logsView.Text = welcomeMessage
	logsView.WrapText = true

	filesView = widgets.NewTable()
	filesView.TextStyle = ui.NewStyle(ui.ColorWhite)
	filesView.TextAlignment = ui.AlignLeft
	filesView.RowSeparator = false

	downloadListView = widgets.NewTable()
	downloadListView.Title = "Tasks"
	filesView.TextStyle = ui.NewStyle(ui.ColorWhite)
	downloadListView.TextAlignment = ui.AlignLeft

	loadingView = widgets.NewParagraph()
	loadingView.Text = "Loading..."

	drawViews(ui.Resize{
		Width:  width,
		Height: height,
	})
}

func drawViews(tSize ui.Resize) {
	ui.Clear()

	header.SetRect(1, 0, tSize.Width-1, 3)

	tab.SetRect(1, tSize.Height-3, tSize.Width-1, tSize.Height)

	logsView.SetRect(1, 3, tSize.Width-1, tSize.Height-3)

	filesView.SetRect(1, 3, tSize.Width-1, tSize.Height-3)
	filesView.ColumnWidths = []int{tSize.Width - 15, 10}

	downloadListView.ColumnWidths = []int{tSize.Width - 37, 10, 10, 12}
	ui.Render(header, tab)
	switch tab.ActiveTabIndex {
	case 0:
		ui.Render(logsView)
	case 1:
		filesView.Rows = [][]string{
			{"Name", "Size"},
		}
		filesView.RowStyles = make(map[int]ui.Style)
		if len(viewFileList) > 0 {
			filesView.Rows = append(filesView.Rows, []string{"...", ""})
			for i, file := range viewFileList {
				filesView.Rows = append(filesView.Rows, []string{models.FilterChinese(file.Name), models.GetSizeString(file.Size)})
				if file.IsDir {
					filesView.RowStyles[i+2] = ui.NewStyle(ui.ColorGreen)
				}
			}
			if len(viewFileList) == tSize.Height-13 {
				filesView.Rows = append(filesView.Rows, []string{"->", ""})
			}
			filesView.RowStyles[lightIndex+1] = ui.NewStyle(ui.ColorWhite, ui.ColorRed, ui.ModifierBold)
		}
		ui.Render(filesView)
	case 2:
		downloadListView.Rows = [][]string{
			{"File", "Status", "Progress", "Speed"},
		}
		downloadListView.SetRect(1, 3, tSize.Width-1, tSize.Height-3)
		if len(DownloaderList) > 0 {
			for i, client := range DownloaderList {
				status := "Wait"
				progress := 0
				if client.Downloading {
					progress = int(float64(client.DownloadedSize) / float64(client.Info.ContentSize) * 100)
					status = "Down"
					downloadListView.RowStyles[i+1] = ui.NewStyle(ui.ColorBlue)
				}
				downloadListView.Rows = append(downloadListView.Rows, []string{client.Info.TargetFile, status, strconv.FormatInt(int64(progress), 10) + "%", models.GetSizeString(client.Speed) + "/s"})

			}
		}
		ui.Render(downloadListView)
	}

	if loading {
		loadingView.SetRect(tSize.Width/2-7, tSize.Height/2-2, tSize.Width/2+7, tSize.Height/2+1)
		ui.Render(loadingView)
	}
}

func SetPath(path string) {
	header.Text = path
	RedrawUi()
}

func AppendLog(log string) {
	logsView.Text += "\n" + log
	RedrawUi()
}
