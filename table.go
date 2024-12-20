package painter

type Table struct {
	heads *TableHead
	rows  *TableRow
}

func NewTable(heads *TableHead, rows *TableRow) *Table {
	return &Table{
		heads: heads, rows: rows,
	}
}

type TableHead struct {
	Font          string
	FontStyle     string
	FontSize      float64
	HeightPerLine float64
	Heads         []*HeadCell
}

type HeadCell struct {
	Text  string
	Width float64
}

type Cell struct {
	X, Y, Span int
	Type       uint8 // 0 colspan, 1 rowspan
	Text       string
	Font       string
	FontStyle  string
	FontSize   float64
}

const (
	Nonespan = 0
	Colspan  = 1
	Rowspan  = 2
)

type TableRow struct {
	Font          string
	FontStyle     string
	FontSize      float64
	HeightPerLine float64
	RowNums       int
	Spans         []*Cell
}

func (t *Table) GetX(col int) float64 {
	w := float64(0)
	for i, head := range t.heads.Heads {
		if i < col {
			w += head.Width
		}
	}
	return w
}

func (t *Table) GetY(row int) float64 {
	return float64(row) * t.rows.HeightPerLine
}
