package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
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

	// 底部操作区：转换按钮 + 进度 + 状态（固定在窗口底部，始终可见）
	statusLabel.Alignment = fyne.TextAlignCenter
	progressContent := container.NewVBox(
		container.NewGridWithColumns(1, progressBar),
		ui.NewVSpacer(4),
		statusLabel,
	)
	progressCard := ui.NewCard("转换进度", progressContent)

	headerArea := container.NewVBox(header, ui.NewVSpacer(14))
	bottomArea := container.NewVBox(
		container.NewGridWithColumns(1, convertBtn),
		ui.NewVSpacer(12),
		progressCard,
	)

	// 中部卡片区放入滚动容器：窗口高度受限时卡片可滚动，头部与底部按钮不会被裁掉
	cardsColumn := container.NewVBox(
		pdfCard,
		ui.NewVSpacer(12),
		folderCard,
		ui.NewVSpacer(12),
		outputCard,
	)
	cardsScroll := container.NewVScroll(cardsColumn)

	form := container.NewPadded(container.NewBorder(headerArea, bottomArea, nil, nil, cardsScroll))

	w.SetContent(form)

	// 以内容最小高度为基准决定窗口高度，并夹在屏幕可用区域内
	desiredHeight := 2*theme.Padding() +
		headerArea.MinSize().Height +
		cardsColumn.MinSize().Height +
		bottomArea.MinSize().Height
	appliedHeight := fitWindowToContent(w, desiredHeight, 0)

	w.CenterOnScreen()
	w.Show()

	// 窗口创建后缩放比例、标题栏高度与实际布局尺寸才确定；再以实际卡区高度为准校正一次，
	// 让内容无需滚动即可完整显示，同时仍夹在屏幕可用区域内
	go func() {
		time.Sleep(300 * time.Millisecond)
		shortfall := cardsColumn.MinSize().Height - cardsScroll.Size().Height
		if shortfall > 0 {
			// 额外留 1 个逻辑像素余量，避免亚像素级溢出导致滚动条常驻
			desiredHeight += shortfall + 1
		}
		appliedHeight = fitWindowToContent(w, desiredHeight, appliedHeight)
		w.CenterOnScreen()
	}()

	a.Run()
}

const (
	windowWidth     float32 = 880 // 视觉基准宽度
	windowMinHeight float32 = 420
	// 窗口装饰（标题栏、边框）与任务栏重叠余量，计算可用高度时需扣除
	windowDecorationPixels float32 = 128
)

// fitWindowToContent 以内容最小高度为基准设置窗口高度，并夹在屏幕可用区域内。
// applied 为上一次请求的窗口外框高度：窗口创建后用它推算标题栏/边框占用的高度，
// 使内容区高度正好等于 contentHeight，避免把装饰高度硬编码成常量。
// 返回本次请求的外框高度。
func fitWindowToContent(w fyne.Window, contentHeight, applied float32) float32 {
	h := contentHeight
	if cs := w.Canvas().Size(); applied > 0 && cs.Height > 0 && applied > cs.Height {
		h += applied - cs.Height
	}
	if maxH := maxWindowHeight(w); maxH > 0 && h > maxH {
		h = maxH
	}
	if h < windowMinHeight {
		h = windowMinHeight
	}
	w.Resize(fyne.NewSize(windowWidth, h))
	return h
}

// maxWindowHeight 返回窗口外框可用的最大高度（逻辑像素），0 表示无法获取屏幕信息
func maxWindowHeight(w fyne.Window) float32 {
	_, screenH := usableScreenSize()
	if screenH <= 0 {
		return 0
	}
	scale := w.Canvas().Scale()
	if scale <= 1 {
		// 窗口尚未创建时 canvas scale 还是默认值 1，用系统 DPI 与用户缩放设置估算
		scale = systemDPIScale()
		if us := fyne.CurrentApp().Settings().Scale(); us > 0 {
			scale *= us
		}
	}
	if scale <= 0 {
		scale = 1
	}
	h := (float32(screenH) - windowDecorationPixels) / scale
	if h < windowMinHeight {
		return 0
	}
	return h
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
