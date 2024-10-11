// Package painter
package painter

import (
	"bytes"
	"image/jpeg"
	"io"
	"math"
	"strings"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/code128"
	"github.com/boombuler/barcode/qr"
	"github.com/mnhkahn/gofpdf"
)

type PdfPainter struct {
	pdf            *gofpdf.Fpdf
	orientationStr string
	w              float64
	h              float64
	resource       string
}

func NewPdfPainter(w float64, h float64) *PdfPainter {
	return &PdfPainter{w: w, h: h}
}

func NewPdfPainterResource(w float64, h float64, resource string) (*PdfPainter, error) {
	p := NewPdfPainter(w, h)
	if err := p.Init(resource); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *PdfPainter) Init(resource string) error {
	p.orientationStr = "P"
	p.pdf = gofpdf.NewCustom(&gofpdf.InitType{
		OrientationStr: p.orientationStr,
		UnitStr:        "mm",
		Size: gofpdf.SizeType{
			p.w, p.h,
		},
	})
	p.resource = resource
	if err := initFontStyle(p.pdf, resource); err != nil {
		return err
	}
	return nil
}

func (p *PdfPainter) GetName() string {
	return "pdf"
}

func (p *PdfPainter) AddPage(wh ...float64) error {
	if len(wh) == 2 {
		p.pdf.AddPageFormat(p.orientationStr, gofpdf.SizeType{
			Wd: wh[0], Ht: wh[1],
		})
	} else {
		p.pdf.AddPageFormat(p.orientationStr, gofpdf.SizeType{
			Wd: p.w, Ht: p.h,
		})
	}

	p.pdf.SetAutoPageBreak(false, 0)
	return nil
}

func (p *PdfPainter) Line(x1, y1, x2, y2, width float64, isDash bool) error {
	if isDash {
		p.pdf.SetDashPattern([]float64{0.8, 0.8}, 0)
	} else {
		p.pdf.SetDashPattern([]float64{}, 0)
	}
	old := p.pdf.GetLineWidth()
	p.pdf.SetLineWidth(width)
	p.pdf.Line(x1, y1, x2, y2)
	p.pdf.SetLineWidth(old)
	return nil
}

func (p *PdfPainter) Text(text, font, fontStyle string, fontSize, x, y, w, h float64, align string, color *Color, border string) error {
	_fontSize := getTextWidthStyle(p.pdf, text, font, fontStyle, w, fontSize)
	p.pdf.SetFont(font, fontStyle, _fontSize)
	if color != nil {
		oldR, oldG, oldB := p.pdf.GetTextColor()
		p.pdf.SetTextColor(int(color.R()), int(color.G()), int(color.B()))
		cellText(p.pdf, text, border, gofpdf.LineBreakNormal, align, false, 0, "", x, y, w, h)
		p.pdf.SetTextColor(oldR, oldG, oldB)
	} else {
		cellText(p.pdf, text, border, gofpdf.LineBreakNormal, align, false, 0, "", x, y, w, h)
	}
	return nil
}

func (p *PdfPainter) TextWithTransform(text, font, fontStyle string, fontSize, x, y, w, h, h2, angle float64, align string, color *Color) error {
	p.pdf.TransformBegin()
	p.pdf.TransformRotate(angle, x+(h+h2)/2, y+(h+h2)/2)
	p.Text(text, font, fontStyle, fontSize, x, y+h, w, h2, align, color, "")
	p.pdf.TransformEnd()
	return nil
}

func (p *PdfPainter) RectText(text, font, fontStyle string, fontSize, hPerLine, x, y, w, h float64, color *Color) error {
	if hPerLine <= 0 {
		hPerLine = getLineHeight(p.pdf, fontSize)
	}
	p.pdf.SetFont(font, fontStyle, fontSize)
	oldR, oldG, oldB := p.pdf.GetTextColor()
	if color != nil {
		p.pdf.SetTextColor(int(color.R()), int(color.G()), int(color.B()))
	}
	texts := splitText(p.pdf, text, w)
	i := 0
	for _, text := range texts {
		_y := y + (hPerLine * float64(i))
		if _y > y+h-hPerLine {
			break
		}
		cellText(p.pdf, text, "", gofpdf.LineBreakNormal, AlignLeftMiddle, false, 0, "", x, _y, w, hPerLine)
		i++
	}
	p.pdf.SetTextColor(oldR, oldG, oldB)
	return nil
}

func (p *PdfPainter) Barcode(code string, x, y, w, h float64) error {
	textBarcodeName, err := registerCode128ImageReader(p.pdf, code, getBarcodeWidth(code))
	if err != nil {
		return err
	}
	p.pdf.Image(textBarcodeName, x, y, w, h, false, "", 0, "")
	return nil
}

func (p *PdfPainter) BarcodeWithTransform(code string, x, y, w, h, h2, angle float64) error {
	p.pdf.TransformBegin()
	p.pdf.TransformRotate(angle, x+(h+h2)/2, y+(h+h2)/2)
	p.Barcode(code, x, y, w, h)
	p.pdf.TransformEnd()
	return nil
}

func (p *PdfPainter) Picture(pic string, x, y, w, h float64) error {
	if pic == "" {
		//return errors.New("pic is nil")
		return nil
	}
	p.pdf.Image(p.resource+pic, x, y, w, h, false, "", 0, "")
	return nil
}

func (p *PdfPainter) Rect(styleStr string, x, y, w, h float64) error {
	p.pdf.Rect(x, y, w, h, styleStr)
	return nil
}

func (p *PdfPainter) QRCode(code string, x, y, w, h float64) error {
	bcode, err := qr.Encode(code, qr.H, qr.Unicode)
	if err != nil {
		return err
	}

	scaledBCode := bcode
	// 判断是否需要缩放
	orgBounds := bcode.Bounds()
	orgWidth := orgBounds.Max.X - orgBounds.Min.X
	orgHeight := orgBounds.Max.Y - orgBounds.Min.Y
	factor := int(math.Min(w/float64(orgWidth), h/float64(orgHeight)))
	if factor > 0 {
		scaleBCode, err := barcode.Scale(bcode, int(w), int(h))
		if err != nil {
			return err
		}
		scaledBCode = scaleBCode
	}

	name := barcodeKey(scaledBCode)
	buf := bytes.NewBuffer(nil)
	_ = jpeg.Encode(buf, scaledBCode, nil)
	p.pdf.RegisterImageReader(name, "jpg", buf)
	p.pdf.Image(name, x, y, w, h, false, "", 0, "")
	return nil
}

func (p *PdfPainter) Output(writer io.Writer) error {
	return p.pdf.Output(writer)
}

// initFontStyle 目前只支持黑体
func initFontStyle(pdf *gofpdf.Fpdf, resourceDir string) error {
	// 写字
	pdf.AddUTF8Font(FontSimhei, "", resourceDir+"/font/simhei.ttf")
	err := pdf.Error()
	if err != nil {
		return err
	}
	pdf.AddUTF8Font(FontSimhei, FontBold, resourceDir+"/font/simhei.ttf")
	err = pdf.Error()
	if err != nil {
		return err
	}

	return nil
}

func getTextWidthStyle(pdf *gofpdf.Fpdf, txt string, font, fontStyle string, w, fontSize float64) float64 {
	pdf.SetFont(font, fontStyle, fontSize)
	l := pdf.GetStringWidth(txt)

	_fontSize := fontSize
	for l > w {
		_fontSize -= 0.2
		pdf.SetFontSize(_fontSize)
		l = pdf.GetStringWidth(txt)
	}
	return _fontSize
}

func cellText(pdf *gofpdf.Fpdf, txtStr, borderStr string, ln int, alignStr string, fill bool, link int, linkStr string, x, y, w, h float64) {
	if h <= 0 {
		fontSizePt, _ := pdf.GetFontSize()
		h = getLineHeight(pdf, fontSizePt)
	}
	pdf.SetXY(x, y)
	//borderStr = gofpdf.BorderFull
	pdf.CellFormat(w, h, txtStr, borderStr, ln, alignStr, fill, link, linkStr)
}

func getLineHeight(pdf *gofpdf.Fpdf, fontSize float64) float64 {
	return fontSize / pdf.GetConversionRatio()
}

func splitText(pdf *gofpdf.Fpdf, txt string, w float64) (lines []string) {
	s := []rune(txt) // Return slice of UTF-8 runes
	if len(s) == 1 {
		return []string{txt}
	}

	j := 0
	minWidth := pdf.GetStringWidth("你")
	for i := 1; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, string(s[j:i]))
			j = i + 1
			continue
		}

		sw := pdf.GetStringWidth(string(s[j:i]))
		if w-sw <= minWidth {
			lines = append(lines, string(s[j:i]))
			j = i
		}
	}
	// 最后一条
	if j < len(s) {
		lines = append(lines, string(s[j:]))
	}

	return lines
}

func registerCode128ImageReader(pdf *gofpdf.Fpdf, code string, width int) (string, error) {
	bcode, err := code128.Encode(code)
	if err != nil {
		return "", err
	}
	scaleBCode, _ := barcode.Scale(bcode, width, 1)

	code128Name := barcodeKey(scaleBCode)

	buf := bytes.NewBuffer(nil)
	_ = jpeg.Encode(buf, scaleBCode, nil)
	pdf.RegisterImageReader(code128Name, "jpg", buf)

	return code128Name, nil
}

func barcodeKey(bcode barcode.Barcode) string {
	return bcode.Metadata().CodeKind + bcode.Content()
}

func getBarcodeWidth(code string) int {
	if strings.HasPrefix(strings.ToLower(code), "jd") {
		return 520
	} else {
		return 600
	}
}
