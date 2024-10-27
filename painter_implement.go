// Package painter
package painter

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/code128"
	"github.com/boombuler/barcode/qr"
	"github.com/mnhkahn/gofpdf"
	"github.com/mnhkahn/gogogo/logger"
	"github.com/mozillazg/go-pinyin"
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
	err := filepath.Walk(resourceDir+"/font", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".ttf" {
			file_name := filepath.Base(path) //use this built-in function to obtain filename
			names := strings.Split(file_name, "_")
			if len(names) == 1 {
				nameWithoutExt := strings.TrimSuffix(file_name, ".ttf")
				logger.Info("init font", nameWithoutExt, "normal", path)
				pdf.AddUTF8Font(nameWithoutExt, "", path)
				err := pdf.Error()
				if err != nil {
					return err
				}
			} else if len(names) >= 2 {
				style := strings.TrimSuffix(names[len(names)-1], ".ttf")
				nameWithoutExt := strings.Join(names[:len(names)-1], "_")
				if strings.ToLower(style) == "bold" {
					logger.Info("init font", nameWithoutExt, FontBold, path)
					pdf.AddUTF8Font(nameWithoutExt, FontBold, path)
					err = pdf.Error()
					if err != nil {
						return err
					}
				} else if strings.ToLower(style) == "italic" {
					logger.Info("init font", nameWithoutExt, FontItalic, path)
					pdf.AddUTF8Font(nameWithoutExt, FontItalic, path)
					err = pdf.Error()
					if err != nil {
						return err
					}
				} else {
					logger.Info("init font", nameWithoutExt, "normal", path)
					pdf.AddUTF8Font(nameWithoutExt, FontNone, path)
					err := pdf.Error()
					if err != nil {
						return err
					}
				}
			}
		}

		return nil
	})

	return err
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

type lineFromTo struct {
	start, end float64
}

func delLineFromTo(l, o *lineFromTo) []*lineFromTo {
	if o == nil {
		return []*lineFromTo{l}
	} else if o.start == l.start {
		if o.end >= l.end {
			return nil
		} else {
			return []*lineFromTo{{o.end, l.end}}
		}
	} else if o.end == l.end {
		if l.start >= o.start {
			return nil
		} else {
			return []*lineFromTo{{l.start, o.start}}
		}
	} else if o.start >= l.end || o.end <= l.start {
		return []*lineFromTo{l}
	} else {
		return []*lineFromTo{{l.start, o.start}, {o.end, l.end}}
	}
}

func debugLineFromTo(a []*lineFromTo) string {
	buf := bytes.NewBuffer(nil)
	for _, aa := range a {
		buf.WriteString(fmt.Sprintf("[%v:%v]", aa.start, aa.end))
	}
	return buf.String()
}

func (p *PdfPainter) Table(startX, startY float64, font, fontStyle string, fontSize float64, table *Table) error {
	// 计算总宽度
	totalWidth := float64(0)
	for _, head := range table.heads {
		totalWidth += head.Width
	}
	totalHeight := table.rows.HeightPerLine * float64(table.rows.RowNums)

	// 初始化所有线
	colLines := map[int][]*lineFromTo{}
	for i := range table.heads {
		colLines[i] = append(colLines[i], &lineFromTo{start: startY, end: startY + totalHeight})
	}
	rowLines := map[int][]*lineFromTo{}
	for i := 0; i <= table.rows.RowNums; i += 1 {
		rowLines[i] = append(rowLines[i], &lineFromTo{start: startX, end: startX + totalWidth})
	}
	for _, span := range table.rows.Spans {
		for i := 1; i < span.Span; i++ {
			if span.Type == Colspan {
				// 从colLines里剔除
				tobeDel := &lineFromTo{start: startY + table.GetY(span.Y), end: startY + table.GetY(span.Y+1)}
				newColLines := make([]*lineFromTo, 0, len(colLines[span.X+i]))
				for _, colLine := range colLines[span.X+i] {
					newColLines = append(newColLines, delLineFromTo(colLine, tobeDel)...)
					// logger.Info("BBB", span.X+i, colLine, tobeDel, debugLineFromTo(newColLines))
				}
				colLines[span.X+i] = newColLines
			} else if span.Type == Rowspan {
				tobeDel := &lineFromTo{start: startX + table.GetX(span.X), end: startX + table.GetX(span.X+1)}
				newRowLines := make([]*lineFromTo, 0, len(rowLines[span.Y+i]))
				for _, rowLine := range rowLines[span.Y+i] {
					newRowLines = append(newRowLines, delLineFromTo(rowLine, tobeDel)...)
					// logger.Info("CCC", span.Y+i, rowLine, tobeDel, debugLineFromTo(newRowLines))
				}
				rowLines[span.Y+i] = newRowLines
			}
		}
		if span.Text != "" {
			textX := table.GetX(span.X)
			textY := table.GetY(span.Y)
			textWidth := table.GetX(span.Span+span.X) - textX
			if span.Span == 0 {
				textWidth = table.GetX(span.X+1) - textX
			}
			textHeight := table.rows.HeightPerLine
			if span.Type == Rowspan {
				textWidth = table.GetX(span.X+1) - textX
				textHeight = table.GetY(span.Span+span.Y) - textY
			}
			// logger.Info("text", span.Text, textX, textY, textWidth, textHeight)
			p.Text(span.Text, FontSimhei, "", fontSize, textX+startX, textY+startY, textWidth, textHeight, AlignCenterMiddle, nil, gofpdf.BorderNone)
		}
	}

	start := startX
	for i, head := range table.heads {
		// 竖线
		if span := colLines[i]; span != nil {
			for _, s := range span {
				p.Line(start, s.start, start, s.end, 0.1, false)
			}
		}

		p.Text(head.Text, font, fontStyle, fontSize, start, startY, head.Width, table.rows.HeightPerLine, AlignCenterMiddle, nil, gofpdf.BorderNone)
		start += head.Width
	}
	// 补最右竖线
	p.Line(startX+totalWidth, startY, startX+totalWidth, startY+totalHeight, 0.1, false)

	for i := 0; i <= table.rows.RowNums; i += 1 {
		// 横线
		// 如果设置了rowspan，横线需要处理，变短或者画两段
		if span := rowLines[i]; span != nil {
			for _, s := range span {
				// logger.Info("DDDD", s.start, s.end)
				p.Line(s.start, float64(i)*table.rows.HeightPerLine+startY, s.end, float64(i)*table.rows.HeightPerLine+startY, 0.1, false)
			}
		}
	}
	return nil
}

func (p *PdfPainter) MiShapeWithPinyin(text, font, fontStyle string, fontSize, x, y, w, hPinyin float64) error {
	a := pinyin.NewArgs()
	a.Style = pinyin.Tone
	pinyin := pinyin.Pinyin(text, a)
	for i, v := range pinyin {
		p.Text(v[0], font, fontStyle, fontSize, x+float64(i)*w, y, w, hPinyin, AlignCenterMiddle, nil, gofpdf.BorderNone)
	}

	p.MiShape(x, y+hPinyin, w, len(pinyin))
	return nil
}

func (p *PdfPainter) MiShape(x, y, w float64, wordNum int) error {
	totalWidth := w * float64(wordNum)
	totalHeight := w
	// 外框：上右下左
	p.Line(x, y, x+totalWidth, y, 0.1, false)
	p.Line(x+totalWidth, y, x+totalWidth, y+totalHeight, 0.1, false)
	p.Line(x, y+totalHeight, x+totalWidth, y+totalHeight, 0.1, false)
	p.Line(x, y, x, y+totalHeight, 0.1, false)
	// 字间隔竖线
	for i := 1; i < wordNum; i++ {
		p.Line(x+float64(i)*w, y, x+float64(i)*w, y+totalHeight, 0.1, false)
	}
	// 中间虚线
	p.Line(x, y+totalHeight/2, x+totalWidth, y+totalHeight/2, 0.1, true)
	// 每个字虚线
	for i := 0; i < wordNum; i++ {
		p.Line(x+(float64(i)+0.5)*w, y, x+(float64(i)+0.5)*w, y+totalHeight, 0.1, true)
	}

	return nil
}
