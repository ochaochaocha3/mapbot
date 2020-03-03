// colorutil は色についての操作を集めるパッケージ。
package colorutil

import (
	"image/color"
	"math/rand"

	"github.com/jyotiska/go-webcolors"
)

var (
	ChitColors []color.RGBA
)

func init() {
	ChitColors = append(ChitColors, CSS3NameToRGBA("deeppink"))
	ChitColors = append(ChitColors, CSS3NameToRGBA("red"))
	ChitColors = append(ChitColors, CSS3NameToRGBA("orange"))
	ChitColors = append(ChitColors, CSS3NameToRGBA("gold"))
	ChitColors = append(ChitColors, CSS3NameToRGBA("chocolate"))
	ChitColors = append(ChitColors, CSS3NameToRGBA("limegreen"))
	ChitColors = append(ChitColors, CSS3NameToRGBA("forestgreen"))
	ChitColors = append(ChitColors, CSS3NameToRGBA("dodgerblue"))
	ChitColors = append(ChitColors, CSS3NameToRGBA("darkorchid"))
	ChitColors = append(ChitColors, CSS3NameToRGBA("slategray"))
}

func RGBToRGBA(rgb []int) color.RGBA {
	return color.RGBA{uint8(rgb[0]), uint8(rgb[1]), uint8(rgb[2]), 0xFF}
}

func CSS3NameToRGBA(name string) color.RGBA {
	return RGBToRGBA(webcolors.NameToRGB(name, "css3"))
}

func RandomChitColor() color.RGBA {
	i := rand.Intn(len(ChitColors))
	return ChitColors[i]
}
