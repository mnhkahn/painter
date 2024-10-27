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
	Table(startX, startY float64, font, fontStyle string, fontSize float64, table *Table) error
	MiShapeWithPinyin(text, font, fontStyle string, fontSize, x, y, w, hPinyin float64) error
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
