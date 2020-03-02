// rpgmap はRPGのマップに関する機能を提供するパッケージ。
package rpgmap

import (
	"fmt"
)

// SquareMap はスクエアマップを表す構造体。
type SquareMap struct {
	// Width はマップの幅。
	width int
	// Height はマップの高さ。
	height int
	// Chits はチットの配列。
	chits []*Chit
}

// NewSquareMap は新しいスクエアマップを返す。
func NewSquareMap(width int, height int) (*SquareMap, error) {
	if width <= 0 {
		return nil, fmt.Errorf("width must be greater than 0 (%d)", width)
	}

	if height <= 0 {
		return nil, fmt.Errorf("height must be greater than 0 (%d)", height)
	}

	return &SquareMap{
		width:  width,
		height: height,
	}, nil
}

// Width はマップの幅を返す。
func (m *SquareMap) Width() int {
	return m.width
}

// Height はマップの高さを返す。
func (m *SquareMap) Height() int {
	return m.height
}

// Chits はチットの配列のコピーを返す。
func (m *SquareMap) Chits() []*Chit {
	chits := make([]*Chit, len(m.chits))
	copy(chits, m.chits)

	return chits
}

// String はマップを表す文字列を返す。
func (m *SquareMap) String() string {
	return fmt.Sprintf("SquareMap (%s)", m.SizeStr())
}

// SizeStr はマップの大きさを表す文字列を返す。
func (m *SquareMap) SizeStr() string {
	return fmt.Sprintf("%d x %d", m.width, m.height)
}

func (m *SquareMap) AddChit(c *Chit) error {
	if c.X < 0 || c.X >= m.width {
		return fmt.Errorf("AddChit: X is out of range")
	}

	if c.Y < 0 || c.Y >= m.height {
		return fmt.Errorf("AddChit: Y is out of range")
	}

	m.chits = append(m.chits, c)

	return nil
}
