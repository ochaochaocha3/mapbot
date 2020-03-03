package rpgmap

import (
	"fmt"
	"image/color"
)

// Chit は駒を表す構造体。
type Chit struct {
	// Name は駒の名前。
	Name string
	// X は駒のx座標。
	X int
	// Y は駒のy座標。
	Y int
	// Color は駒の色。
	Color color.RGBA
}

// String は駒を表す文字列を返す。
func (c *Chit) String() string {
	return fmt.Sprintf("%s %s", c.Name, c.CoordStr())
}

// CoordStr は駒の座標を表す文字列を返す。
func (c *Chit) CoordStr() string {
	return fmt.Sprintf("(%d, %d)", c.X+1, c.Y+1)
}
