package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

// 界面配色与主题（theme.go）保持一致
var (
	CardBackground = color.NRGBA{R: 0x1E, G: 0x22, B: 0x2A, A: 0xFF}
	CardBorder     = color.NRGBA{R: 0x32, G: 0x38, B: 0x43, A: 0xFF}
	CardTitle      = color.NRGBA{R: 0x8E, G: 0x98, B: 0xAB, A: 0xFF}
)

// NewCard 返回一个带圆角背景、细边框和小标题的卡片容器
func NewCard(title string, content fyne.CanvasObject) fyne.CanvasObject {
	bg := canvas.NewRectangle(CardBackground)
	bg.CornerRadius = 12

	border := canvas.NewRectangle(color.Transparent)
	border.StrokeColor = CardBorder
	border.StrokeWidth = 1
	border.CornerRadius = 12

	titleText := canvas.NewText(title, CardTitle)
	titleText.TextStyle = fyne.TextStyle{Bold: true}
	titleText.TextSize = 13

	gap := canvas.NewRectangle(color.Transparent)
	gap.SetMinSize(fyne.NewSize(1, 6))

	inner := container.NewVBox(titleText, gap, content)
	padded := container.NewPadded(container.NewPadded(inner))

	return container.NewStack(bg, border, padded)
}

// NewVSpacer 返回固定高度的垂直间距
func NewVSpacer(h float32) fyne.CanvasObject {
	r := canvas.NewRectangle(color.Transparent)
	r.SetMinSize(fyne.NewSize(1, h))
	return r
}
