// rpgmap はRPGのマップに関する機能を提供するパッケージ。
package rpgmap

import (
	"container/list"
	"fmt"
	"sync"
)

// stringChitMap は、文字列 -> チットの対応の型。
type stringListElementMap map[string]*list.Element

// SquareMap はスクエアマップを表す構造体。
type SquareMap struct {
	// Width はマップの幅。
	width int
	// Height はマップの高さ。
	height int
	// Chits はチットの配列。
	chits []*Chit
	// chitList はチットの連結リスト。
	chitList *list.List
	// nameToChitListElement はチットの名前とチットとの対応。
	nameToChitListElement stringListElementMap
	// mux は排他制御用のミューテックス。
	mux sync.Mutex
}

// NewSquareMap は新しいスクエアマップを返す。
func NewSquareMap(width int, height int) (*SquareMap, error) {
	if width < 2 {
		return nil, fmt.Errorf("width must be greater than or equal to 2 (%d)", width)
	}

	if height < 2 {
		return nil, fmt.Errorf("height must be greater than or equal to 2 (%d)", height)
	}

	return &SquareMap{
		width:                 width,
		height:                height,
		chits:                 []*Chit{},
		chitList:              list.New(),
		nameToChitListElement: stringListElementMap{},
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

// String はマップを表す文字列を返す。
func (m *SquareMap) String() string {
	return fmt.Sprintf("SquareMap (%s)", m.SizeStr())
}

// SizeStr はマップの大きさを表す文字列を返す。
func (m *SquareMap) SizeStr() string {
	return fmt.Sprintf("%d x %d", m.width, m.height)
}

// NumOfChits はマップに含まれるチット数を返す。
func (m *SquareMap) NumOfChits() int {
	return m.chitList.Len()
}

// FindChit は名前からチットを検索する。
func (m *SquareMap) FindChit(name string) (*Chit, bool) {
	e, found := m.nameToChitListElement[name]
	if !found {
		return nil, false
	}

	return e.Value.(*Chit), true
}

// 各チットに対して処理を行う。
func (m *SquareMap) ForEachChit(f func(i int, c *Chit)) {
	i := 0
	for e := m.chitList.Front(); e != nil; e = e.Next() {
		f(i, e.Value.(*Chit))
		i++
	}
}

// AddChit はチットを追加する。
func (m *SquareMap) AddChit(c *Chit) error {
	m.mux.Lock()
	defer m.mux.Unlock()

	if _, found := m.FindChit(c.Name); found {
		return fmt.Errorf(`chit "%s" already exists`, c.Name)
	}

	if !m.XIsInRange(c.X) {
		return fmt.Errorf("X is out of range: %d", c.X)
	}

	if !m.YIsInRange(c.Y) {
		return fmt.Errorf("Y is out of range: %d", c.Y)
	}

	e := m.chitList.PushBack(c)
	m.nameToChitListElement[c.Name] = e

	return nil
}

// DeleteChit はチットを削除する。
func (m *SquareMap) DeleteChit(name string) error {
	m.mux.Lock()
	defer m.mux.Unlock()

	e, found := m.nameToChitListElement[name]
	if !found {
		return fmt.Errorf("chit not found: %s", name)
	}

	m.chitList.Remove(e)
	delete(m.nameToChitListElement, name)

	return nil
}

// MoveChit はチットを移動する。
func (m *SquareMap) MoveChit(name string, newX int, newY int) (*Chit, error) {
	m.mux.Lock()
	defer m.mux.Unlock()

	c, ok := m.FindChit(name)
	if !ok {
		return nil, fmt.Errorf("chit not found: %s", name)
	}

	if !m.XIsInRange(newX) {
		return nil, fmt.Errorf("newX is out of range: %d", newX)
	}

	if !m.YIsInRange(newY) {
		return nil, fmt.Errorf("newY is out of range: %d", newY)
	}

	c.X = newX
	c.Y = newY

	return c, nil
}

// XIsInRange は、x座標がマップの範囲内かを返す。
func (m *SquareMap) XIsInRange(x int) bool {
	return x >= 0 && x < m.width
}

// YIsInRange は、y座標がマップの範囲内かを返す。
func (m *SquareMap) YIsInRange(y int) bool {
	return y >= 0 && y < m.height
}
