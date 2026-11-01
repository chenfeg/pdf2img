package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// cjkFontPaths 按优先级尝试的系统中文字体（.ttc 合集 Fyne 无法解析，跳过）
var cjkFontPaths = []string{
	`C:\Windows\Fonts\simhei.ttf`,
	`C:\Windows\Fonts\simkai.ttf`,
	`C:\Windows\Fonts\simfang.ttf`,
	`C:\Windows\Fonts\simsunb.ttf`,
}

var cachedCJKFont fyne.Resource

// loadCJKFont 加载首个可用的中文 TTF 字体；失败返回 nil（回退默认字体）
func loadCJKFont() fyne.Resource {
	if cachedCJKFont != nil {
		return cachedCJKFont
	}
	for _, p := range cjkFontPaths {
		if r, err := fyne.LoadResourceFromPath(p); err == nil {
			cachedCJKFont = r
			return cachedCJKFont
		}
	}
	return nil
}

// 现代深色主题配色
var (
	colorBackground = color.NRGBA{R: 0x14, G: 0x17, B: 0x1D, A: 0xFF}
	colorButton     = color.NRGBA{R: 0x2A, G: 0x2F, B: 0x3A, A: 0xFF}
	colorPrimary    = color.NRGBA{R: 0x4C, G: 0x8D, B: 0xFF, A: 0xFF}
	colorInputBg    = color.NRGBA{R: 0x22, G: 0x26, B: 0x2E, A: 0xFF}
	colorHover      = color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0x14}
	colorFocus      = color.NRGBA{R: 0x4C, G: 0x8D, B: 0xFF, A: 0x66}
	colorSelection  = color.NRGBA{R: 0x4C, G: 0x8D, B: 0xFF, A: 0x59}
)

// modernTheme 在深色主题基础上覆盖配色与间距
type modernTheme struct {
	fyne.Theme
}

func newModernTheme() fyne.Theme {
	return &modernTheme{Theme: theme.DarkTheme()}
}

func (t *modernTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return colorBackground
	case theme.ColorNameButton:
		return colorButton
	case theme.ColorNamePrimary:
		return colorPrimary
	case theme.ColorNameInputBackground:
		return colorInputBg
	case theme.ColorNameHover:
		return colorHover
	case theme.ColorNameFocus:
		return colorFocus
	case theme.ColorNameSelection:
		return colorSelection
	}
	return t.Theme.Color(name, variant)
}

func (t *modernTheme) Size(name fyne.ThemeSizeName) float32 {
	if name == theme.SizeNamePadding {
		return 6
	}
	return t.Theme.Size(name)
}

// Font 优先返回中文字体，避免部分 Windows 环境下 CJK 渲染为 "?"
func (t *modernTheme) Font(style fyne.TextStyle) fyne.Resource {
	if f := loadCJKFont(); f != nil {
		return f
	}
	return t.Theme.Font(style)
}
