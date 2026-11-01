package ui

import (
	"fmt"
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	segmentCount = 24
	barHeight    = 20
	barPaddingV  = 6
)

var (
	colorStart = color.NRGBA{R: 0, G: 200, B: 255, A: 255}
	colorEnd   = color.NRGBA{R: 160, G: 60, B: 255, A: 255}
	colorGlow  = color.NRGBA{R: 80, G: 130, B: 255, A: 60}
)

type CoolProgressBar struct {
	widget.BaseWidget

	progress float64

	bg      *canvas.Rectangle
	fill    *canvas.Rectangle
	glow    *canvas.Rectangle
	shimmer []*canvas.Rectangle
	label   *canvas.Text
	content *fyne.Container
}

func NewCoolProgressBar() *CoolProgressBar {
	p := &CoolProgressBar{}
	p.ExtendBaseWidget(p)

	p.bg = canvas.NewRectangle(theme.Color(theme.ColorNameInputBackground))
	p.bg.CornerRadius = barHeight / 2

	p.glow = canvas.NewRectangle(color.Transparent)
	p.glow.CornerRadius = barHeight/2 + 2

	p.fill = canvas.NewRectangle(colorStart)
	p.fill.CornerRadius = barHeight / 2

	p.shimmer = make([]*canvas.Rectangle, segmentCount)
	for i := range p.shimmer {
		p.shimmer[i] = canvas.NewRectangle(color.NRGBA{R: 255, G: 255, B: 255, A: 0})
	}

	p.label = canvas.NewText("0%", color.White)
	p.label.Alignment = fyne.TextAlignCenter
	p.label.TextSize = 12
	p.label.TextStyle = fyne.TextStyle{Bold: true}

	shimmerObjs := make([]fyne.CanvasObject, len(p.shimmer))
	for i, s := range p.shimmer {
		shimmerObjs[i] = s
	}

	p.content = container.NewStack(
		p.bg,
		p.glow,
		p.fill,
		container.NewWithoutLayout(shimmerObjs...),
		container.NewCenter(p.label),
	)
	p.content.Resize(fyne.NewSize(300, barHeight+barPaddingV*2))

	return p
}

func (p *CoolProgressBar) MinSize() fyne.Size {
	return fyne.NewSize(300, barHeight+barPaddingV*2)
}

func (p *CoolProgressBar) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(p.content)
}

func (p *CoolProgressBar) SetValue(v float64) {
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}
	p.progress = v
	p.updateVisuals()
}

func (p *CoolProgressBar) SetText(s string) {
	p.label.Text = s
	p.label.Refresh()
}

func (p *CoolProgressBar) updateVisuals() {
	size := p.Size()
	if size.Width <= 0 {
		size = p.MinSize()
	}

	barW := size.Width - 4
	barH := float32(barHeight)
	fillW := barW * float32(p.progress)
	pct := int(p.progress * 100)

	p.label.Text = fmt.Sprintf("%d%%", pct)
	p.label.Refresh()

	p.bg.Move(fyne.NewPos(2, barPaddingV))
	p.bg.Resize(fyne.NewSize(barW, barH))
	p.bg.Refresh()

	p.glow.Move(fyne.NewPos(0, barPaddingV-2))
	p.glow.Resize(fyne.NewSize(fillW+4, barH+4))
	p.glow.FillColor = interpolateColor(colorGlow, color.NRGBA{R: 160, G: 60, B: 255, A: 80}, p.progress)
	p.glow.Refresh()

	p.fill.Move(fyne.NewPos(2, barPaddingV))
	p.fill.Resize(fyne.NewSize(fillW, barH))
	p.fill.FillColor = interpolateColor(colorStart, colorEnd, p.progress)
	p.fill.Refresh()

	if fillW > 0 {
		segW := fillW / float32(segmentCount)
		for i := 0; i < segmentCount; i++ {
			t := float64(i) / float64(segmentCount-1)
			sx := 2 + float32(i)*segW
			p.shimmer[i].Move(fyne.NewPos(sx, barPaddingV))
			p.shimmer[i].Resize(fyne.NewSize(segW+1, barH))
			c := interpolateColor(colorStart, colorEnd, t*p.progress)
			alpha := uint8(20 + 20*math.Sin(t*math.Pi*2+p.progress*math.Pi*4))
			p.shimmer[i].FillColor = color.NRGBA{R: c.R, G: c.G, B: c.B, A: alpha}
			p.shimmer[i].Refresh()
		}
	}
}

func interpolateColor(a, b color.NRGBA, t float64) color.NRGBA {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	return color.NRGBA{
		R: uint8(float64(a.R) + (float64(b.R)-float64(a.R))*t),
		G: uint8(float64(a.G) + (float64(b.G)-float64(a.G))*t),
		B: uint8(float64(a.B) + (float64(b.B)-float64(a.B))*t),
		A: uint8(float64(a.A) + (float64(b.A)-float64(a.A))*t),
	}
}
