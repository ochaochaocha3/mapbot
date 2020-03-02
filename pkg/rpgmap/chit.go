package rpgmap

import (
	"fmt"
)

// Chit は駒を表す構造体。
type Chit struct {
	// Name は駒の名前。
	Name string
	X    int
	Y    int
}

// String は駒を表す文字列を返す。
func (c *Chit) String() string {
	return fmt.Sprintf("%s %s", c.Name, c.CoordStr())
}

// CoordStr は駒の座標を表す文字列を返す。
func (c *Chit) CoordStr() string {
	return fmt.Sprintf("(%d, %d)", c.X+1, c.Y+1)
}
