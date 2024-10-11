// Package painter
package painter

import (
	"os"
	"testing"

	"github.com/issue9/assert"
)

func TestNewPdfPainterResource(t *testing.T) {
	p, err := NewPdfPainterResource(210, 297, "./resource")
	assert.Nil(t, err)
	assert.Nil(t, p.AddPage())
	assert.Nil(t, p.Line(0, 0, 100, 100, 10, false))
	f, err := os.Create("dat2.pdf")
	assert.Nil(t, err)
	assert.Nil(t, p.Output(f))
}
