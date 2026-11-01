package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"pdf2img/converter"
	"pdf2img/ui"
)

func main() {
	a := app.NewWithID("com.pdf2img.app")
	a.Settings().SetTheme(newModernTheme())

	w := a.NewWindow("PDF 转图片工具")
	w.SetIcon(appIcon)
	w.Resize(fyne.NewSize(880, 620))
	w.CenterOnScreen()

	checkDLL(w)

	var pdfPath string
	var saveFolder string
	outputFormat := "png"

	fileName := widget.NewEntry()
	fileName.SetPlaceHolder("输出文件名（默认使用PDF文件名）")

	pdfLabel := widget.NewLabel("未选择文件")
	pdfLabel.Wrapping = fyne.TextWrapBreak

	folderLabel := widget.NewLabel("未选择目录")
	folderLabel.Wrapping = fyne.TextWrapBreak

	progressBar := ui.NewCoolProgressBar()
	statusLabel := widget.NewLabel("就绪")
	statusLabel.Wrapping = fyne.TextWrapBreak

	formatSelect := widget.NewSelect([]string{"PNG", "JPG", "JPEG"}, func(s string) {
		outputFormat = strings.ToLower(s)
	})
	formatSelect.SetSelected("PNG")

	pdfBtn := widget.NewButtonWithIcon("选择 PDF 文件", theme.FileTextIcon(), func() {
		go func() {
			path, err := nativeSelectPDF()
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if path == "" {
				return
			}
			pdfPath = path
			pdfLabel.SetText(filepath.Base(pdfPath))
			base := strings.TrimSuffix(filepath.Base(pdfPath), filepath.Ext(pdfPath))
			fileName.SetText(base)
			saveFolder = filepath.Dir(pdfPath)
			folderLabel.SetText(saveFolder)
		}()
	})

	folderBtn := widget.NewButtonWithIcon("选择保存目录", theme.FolderOpenIcon(), func() {
		go func() {
			path, err := nativeSelectFolder()
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if path == "" {
				return
			}
			saveFolder = path
			folderLabel.SetText(saveFolder)
		}()
	})

	var convertBtn *widget.Button
	convertBtn = widget.NewButtonWithIcon("开始转换", theme.MediaPlayIcon(), func() {
		if pdfPath == "" {
			dialog.ShowError(fmt.Errorf("请先选择PDF文件"), w)
			return
		}
		if saveFolder == "" {
			dialog.ShowError(fmt.Errorf("请选择保存目录"), w)
			return
		}
		outName := strings.TrimSpace(fileName.Text)
		if outName == "" {
			outName = strings.TrimSuffix(filepath.Base(pdfPath), filepath.Ext(pdfPath))
		}
		ext := outputFormat
		if ext == "jpeg" {
			ext = "jpg"
		}
		outputPath := filepath.Join(saveFolder, outName+"."+ext)

		convertBtn.Disable()
		progressBar.SetValue(0)
		statusLabel.SetText("正在准备转换...")

		go runConversion(w, pdfPath, outputPath, outputFormat, convertBtn, progressBar, statusLabel)
	})
	convertBtn.Importance = widget.HighImportance

	// 顶部标题区
	titleIcon := canvas.NewImageFromResource(appIcon)
	titleIcon.FillMode = canvas.ImageFillContain
	titleIcon.SetMinSize(fyne.NewSize(34, 34))

	titleText := canvas.NewText("PDF 转图片工具", theme.Color(theme.ColorNameForeground))
	titleText.TextStyle = fyne.TextStyle{Bold: true}
	titleText.TextSize = 18

	subtitleText := canvas.NewText("将 PDF 页面合并为一张长图，自动去除页面间留白", ui.CardTitle)
	subtitleText.TextSize = 12

	header := container.NewBorder(nil, nil, titleIcon, nil,
		container.NewVBox(titleText, subtitleText))

	// PDF 文件区域
	pdfRow := container.NewBorder(nil, nil, pdfBtn, nil, pdfLabel)
	pdfCard := ui.NewCard("PDF 文件", pdfRow)

	// 保存位置区域
	folderRow := container.NewBorder(nil, nil, folderBtn, nil, folderLabel)
	folderCard := ui.NewCard("保存位置", folderRow)

	// 输出设置区域
	overwriteNote := widget.NewLabel("* 同名图片文件将被直接覆盖")
	overwriteNote.TextStyle = fyne.TextStyle{Italic: true}
	outputForm := widget.NewForm(
		widget.NewFormItem("格式", formatSelect),
		widget.NewFormItem("文件名", fileName),
	)
	outputContent := container.NewVBox(outputForm, overwriteNote)
	outputCard := ui.NewCard("输出设置", outputContent)

	// 底部操作区：转换按钮 + 进度 + 状态
	statusLabel.Alignment = fyne.TextAlignCenter
	progressContent := container.NewVBox(
		container.NewGridWithColumns(1, progressBar),
		ui.NewVSpacer(4),
		statusLabel,
	)
	progressCard := ui.NewCard("转换进度", progressContent)

	form := container.NewPadded(container.NewVBox(
		header,
		ui.NewVSpacer(14),
		pdfCard,
		ui.NewVSpacer(12),
		folderCard,
		ui.NewVSpacer(12),
		outputCard,
		layout.NewSpacer(),
		container.NewGridWithColumns(1, convertBtn),
		ui.NewVSpacer(12),
		progressCard,
	))

	w.SetContent(form)
	w.ShowAndRun()
}

func checkDLL(w fyne.Window) {
	exe, err := os.Executable()
	if err != nil {
		return
	}
	dllPath := filepath.Join(filepath.Dir(exe), "libmupdf.dll")
	if _, err := os.Stat(dllPath); os.IsNotExist(err) {
		dialog.ShowError(fmt.Errorf(
			"找不到 libmupdf.dll\n请将 libmupdf.dll 放到程序所在目录:\n%s",
			filepath.Dir(exe),
		), w)
	}
}

func runConversion(
	w fyne.Window,
	pdfPath, outputPath, format string,
	btn *widget.Button,
	bar *ui.CoolProgressBar,
	status *widget.Label,
) {
	defer btn.Enable()

	progressCh := make(chan converter.Progress, 10)
	doneCh := make(chan error, 1)

	go func() {
		doneCh <- converter.Convert(pdfPath, outputPath, format, progressCh)
	}()

	for {
		select {
		case p := <-progressCh:
			if p.Total > 0 {
				bar.SetValue(float64(p.Current) / float64(p.Total))
			}
			bar.SetText(p.Status)
			status.SetText(p.Status)
		case err := <-doneCh:
			if err != nil {
				dialog.ShowError(err, w)
				status.SetText("转换失败: " + err.Error())
			} else {
				bar.SetValue(1.0)
				bar.SetText("转换完成!")
				status.SetText("转换完成! 保存至: " + outputPath)
				dialog.ShowInformation("完成", "图片已保存至:\n"+outputPath, w)
			}
			return
		}
	}
}
