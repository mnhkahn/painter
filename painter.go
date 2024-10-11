// Package painter
package painter

import (
	"io"

	"github.com/mnhkahn/gofpdf"
)

type Painter interface {
	Init(resource string) error
	GetName() string
	AddPage(wh ...float64) error
	Line(x1, y1, x2, y2, width float64, isDash bool) error
	Text(text, font, fontStyle string, fontSize, x, y, w, h float64, align string, color *Color, border string) error
	TextWithTransform(text, font, fontStyle string, fontSize, x, y, w, h, h2, angle float64, align string, color *Color) error
	RectText(text, font, fontStyle string, fontSize, hPerLine, x, y, w, h float64, color *Color) error
	Barcode(code string, x, y, w, h float64) error
	BarcodeWithTransform(code string, x, y, w, h, h2, angle float64) error
	Picture(pic string, x, y, w, h float64) error
	Rect(styleStr string, x, y, w, h float64) error
	QRCode(code string, x, y, w, h float64) error
	Output(writer io.Writer) error
}

// FontSimhei ...
const (
	FontSimhei = "simhei"

	FontNone   = ""
	FontBold   = "B"
	FontItalic = "I"

	AlignNone         = ""
	AlignCenterMiddle = gofpdf.AlignCenter + gofpdf.AlignMiddle
	AlignLeftMiddle   = gofpdf.AlignLeft + gofpdf.AlignMiddle
	AlignRightMiddle  = gofpdf.AlignRight + gofpdf.AlignMiddle
)

type Color struct {
	r, g, b uint8
}

func NewColor(r, g, b uint8) *Color {
	return &Color{
		r: r,
		g: g,
		b: b,
	}
}

func (c *Color) RGB() (uint8, uint8, uint8) {
	return c.r, c.g, c.b
}

func (c *Color) B() uint8 {
	return c.b
}

func (c *Color) G() uint8 {
	return c.g
}

func (c *Color) R() uint8 {
	return c.r
}
