package converter

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/gen2brain/go-fitz"
)

type Progress struct {
	Current int
	Total   int
	Status  string
}

func Convert(pdfPath, outputPath, format string, progressCh chan<- Progress) error {
	send := func(p Progress) {
		if progressCh != nil {
			progressCh <- p
		}
	}

	send(Progress{Status: "正在打开PDF文件...", Current: 0, Total: 0})

	doc, err := fitz.New(pdfPath)
	if err != nil {
		return fmt.Errorf("无法打开PDF文件: %w", err)
	}
	defer doc.Close()

	numPages := doc.NumPage()
	if numPages == 0 {
		return fmt.Errorf("PDF文件没有页面")
	}

	send(Progress{Status: fmt.Sprintf("共 %d 页，正在渲染...", numPages), Current: 0, Total: numPages})

	trimmedImgs := make([]*image.RGBA, 0, numPages)

	for i := 0; i < numPages; i++ {
		img, err := doc.ImageDPI(i, 300)
		if err != nil {
			return fmt.Errorf("无法渲染第 %d 页: %w", i+1, err)
		}

		trimmed := TrimWhitespace(img)
		trimmedImgs = append(trimmedImgs, trimmed)

		send(Progress{
			Current: i + 1,
			Total:   numPages,
			Status:  fmt.Sprintf("正在处理第 %d/%d 页...", i+1, numPages),
		})
	}

	send(Progress{Status: "正在拼接页面...", Current: numPages, Total: numPages})

	result := StitchImages(trimmedImgs)

	send(Progress{Status: "正在编码图片...", Current: numPages, Total: numPages})

	data, err := EncodeImage(result, format)
	if err != nil {
		return fmt.Errorf("编码图片失败: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("保存文件失败: %w", err)
	}

	send(Progress{Status: "转换完成!", Current: numPages, Total: numPages})
	return nil
}

func EncodeImage(img image.Image, format string) ([]byte, error) {
	var buf bytes.Buffer
	switch strings.ToLower(format) {
	case "jpg", "jpeg":
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 95}); err != nil {
			return nil, err
		}
	case "png":
		if err := png.Encode(&buf, img); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("不支持的格式: %s", format)
	}
	return buf.Bytes(), nil
}

func TrimWhitespace(img image.Image) *image.RGBA {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w == 0 || h == 0 {
		return imageToRGBA(img)
	}

	bgColor := getBackgroundColor(img)
	if !isLight(bgColor) {
		return imageToRGBA(img)
	}

	top := scanFromTop(img, bgColor, w, h)
	bottom := scanFromBottom(img, bgColor, w, h)

	if top > bottom {
		return imageToRGBA(img)
	}

	margin := 10
	top = maxInt(0, top-margin)
	bottom = minInt(h-1, bottom+margin)

	cropped := image.NewRGBA(image.Rect(0, 0, w, bottom-top+1))
	draw.Draw(cropped, cropped.Bounds(), img, image.Point{bounds.Min.X, top + bounds.Min.Y}, draw.Src)
	return cropped
}

func getBackgroundColor(img image.Image) color.RGBA {
	b := img.Bounds()
	corners := []image.Point{
		b.Min,
		{b.Max.X - 1, b.Min.Y},
		{b.Min.X, b.Max.Y - 1},
		{b.Max.X - 1, b.Max.Y - 1},
	}
	var rSum, gSum, bSum uint32
	for _, p := range corners {
		r, g, b, _ := img.At(p.X, p.Y).RGBA()
		rSum += r >> 8
		gSum += g >> 8
		bSum += b >> 8
	}
	n := uint32(len(corners))
	return color.RGBA{uint8(rSum / n), uint8(gSum / n), uint8(bSum / n), 255}
}

func isLight(c color.RGBA) bool {
	gray := 0.299*float64(c.R) + 0.587*float64(c.G) + 0.114*float64(c.B)
	return gray > 200
}

func scanFromTop(img image.Image, bg color.RGBA, w, h int) int {
	for y := 0; y < h; y++ {
		if !isRowMostlyBG(img, bg, y, w, 0.998) {
			return y
		}
	}
	return 0
}

func scanFromBottom(img image.Image, bg color.RGBA, w, h int) int {
	for y := h - 1; y >= 0; y-- {
		if !isRowMostlyBG(img, bg, y, w, 0.998) {
			return y
		}
	}
	return h - 1
}

func scanFromLeft(img image.Image, bg color.RGBA, w, h int) int {
	for x := 0; x < w; x++ {
		if !isColMostlyBG(img, bg, x, h, 0.998) {
			return x
		}
	}
	return 0
}

func scanFromRight(img image.Image, bg color.RGBA, w, h int) int {
	for x := w - 1; x >= 0; x-- {
		if !isColMostlyBG(img, bg, x, h, 0.998) {
			return x
		}
	}
	return w - 1
}

func isRowMostlyBG(img image.Image, bg color.RGBA, rowY, w int, threshold float64) bool {
	bounds := img.Bounds()
	count := 0
	for x := 0; x < w; x++ {
		if colorClose(img.At(bounds.Min.X+x, bounds.Min.Y+rowY), bg, 12) {
			count++
		}
	}
	return float64(count)/float64(w) >= threshold
}

func isColMostlyBG(img image.Image, bg color.RGBA, colX, h int, threshold float64) bool {
	bounds := img.Bounds()
	count := 0
	for y := 0; y < h; y++ {
		if colorClose(img.At(bounds.Min.X+colX, bounds.Min.Y+y), bg, 12) {
			count++
		}
	}
	return float64(count)/float64(h) >= threshold
}

func colorClose(c color.Color, bg color.RGBA, tolerance int) bool {
	r, g, b, _ := c.RGBA()
	dr := absDiff(int(r>>8), int(bg.R))
	dg := absDiff(int(g>>8), int(bg.G))
	db := absDiff(int(b>>8), int(bg.B))
	return dr <= tolerance && dg <= tolerance && db <= tolerance
}

func absDiff(a, b int) int {
	if a > b {
		return a - b
	}
	return b - a
}

func StitchImages(images []*image.RGBA) *image.RGBA {
	if len(images) == 0 {
		return image.NewRGBA(image.Rect(0, 0, 1, 1))
	}

	totalHeight := 0
	maxWidth := 0
	for _, img := range images {
		b := img.Bounds()
		totalHeight += b.Dy()
		if b.Dx() > maxWidth {
			maxWidth = b.Dx()
		}
	}

	canvas := image.NewRGBA(image.Rect(0, 0, maxWidth, totalHeight))
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	y := 0
	for _, img := range images {
		b := img.Bounds()
		iw := b.Dx()
		ih := b.Dy()
		x := (maxWidth - iw) / 2
		dstRect := image.Rect(x, y, x+iw, y+ih)
		draw.Draw(canvas, dstRect, img, b.Min, draw.Src)
		y += ih
	}

	return canvas
}

func imageToRGBA(img image.Image) *image.RGBA {
	if rgba, ok := img.(*image.RGBA); ok {
		return rgba
	}
	b := img.Bounds()
	rgba := image.NewRGBA(b)
	draw.Draw(rgba, b, img, b.Min, draw.Src)
	return rgba
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func OutputPath(folder, pdfPath, format string) string {
	base := strings.TrimSuffix(filepath.Base(pdfPath), filepath.Ext(pdfPath))
	ext := "." + strings.ToLower(format)
	if format == "jpg" {
		ext = ".jpg"
	}
	return filepath.Join(folder, base+ext)
}
