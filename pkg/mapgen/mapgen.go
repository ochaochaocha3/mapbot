package mapgen

import (
	"image"
	"image/color"
	"math"

	"github.com/llgcode/draw2d/draw2dimg"
	"github.com/llgcode/draw2d/draw2dkit"

	"github.com/ochaochaocha3/mapbot/pkg/colorutil"
	"github.com/ochaochaocha3/mapbot/pkg/rpgmap"
)

// SquareMapImage はスクエアマップの画像描画に必要な情報の構造体。
type SquareMapImage struct {
	// Map は描画対象のスクエアマップ。
	Map *rpgmap.SquareMap
	// GridWidth は1マスの幅。
	GridWidth int
	// GridHeight は1マスの高さ。
	GridHeight int
	// rect は描画領域の矩形。
	rect image.Rectangle
	// BackgroundColor はマップの背景色。
	BackgroundColor color.RGBA
	// GridColor はグリッドの線の色。
	GridColor color.RGBA
}

// NewSquareMapImage は新しいスクエアマップ描画情報を返す。
func NewSquareMapImage(m *rpgmap.SquareMap) *SquareMapImage {
	i := &SquareMapImage{
		Map:             m,
		GridWidth:       32,
		GridHeight:      32,
		BackgroundColor: colorutil.CSS3NameToRGBA("white"),
		GridColor:       colorutil.CSS3NameToRGBA("dimgray"),
	}

	i.updateRect()

	return i
}

// Width は画像の幅を返す。
func (img *SquareMapImage) Width() int {
	return img.rect.Dx()
}

// Height は画像の高さを返す。
func (img *SquareMapImage) Height() int {
	return img.rect.Dy()
}

// Render はマップを描画する。
func (img *SquareMapImage) Render() *image.RGBA {
	destRect := image.Rect(0, 0, img.Width(), img.Height())
	dest := image.NewRGBA(destRect)
	gc := draw2dimg.NewGraphicContext(dest)

	fillBackGround(gc, img)
	drawGrid(gc, img)
	drawChits(gc, img)

	return dest
}

// 描画領域の矩形を更新する。
func (i *SquareMapImage) updateRect() {
	width := i.GridWidth * i.Map.Width()
	height := i.GridHeight * i.Map.Height()

	i.rect = image.Rect(0, 0, width, height)
}

// fillBackGround はgcを背景色で塗りつぶす。
func fillBackGround(gc *draw2dimg.GraphicContext, i *SquareMapImage) {
	// 背景色で塗る
	gc.SetFillColor(i.BackgroundColor)
	draw2dkit.Rectangle(gc, 0, 0, float64(i.Width()), float64(i.Height()))
	gc.Fill()
}

// drawGrid はgcにグリッドを描画する。
func drawGrid(gc *draw2dimg.GraphicContext, img *SquareMapImage) {
	gc.SetStrokeColor(img.GridColor)

	for i := 1; i < img.Height(); i++ {
		y := float64(i * img.GridHeight)
		gc.MoveTo(0, y)
		gc.LineTo(float64(img.Width()), y)
		gc.Stroke()
	}

	for j := 1; j < img.Width(); j++ {
		x := float64(j * img.GridWidth)
		gc.MoveTo(x, 0)
		gc.LineTo(x, float64(img.Height()))
		gc.Stroke()
	}
}

// drawChits はgcにチットの集合を描画する。
//
// TODO: 同じ座標の場合にチットの位置をずらす。
func drawChits(gc *draw2dimg.GraphicContext, img *SquareMapImage) {
	chitSize := int(math.Min(float64(img.GridWidth), float64(img.GridHeight))) / 2
	offset := image.Point{X: 0, Y: 0}

	for _, c := range img.Map.Chits() {
		drawChit(gc, chitDrawing{
			Image:  img,
			Chit:   c,
			Size:   chitSize,
			Offset: offset,
		})
	}
}

// chitDrawing はチット描画の情報。
type chitDrawing struct {
	// Image はマップ描画情報。
	Image *SquareMapImage
	// Chit は描画対象チット。
	Chit *rpgmap.Chit
	// Size は描画するチットの大きさ（辺の長さ）。
	Size int
	// Offset は座標のずれ。
	Offset image.Point
}

// drawChit はgcにチットを描画する。
func drawChit(gc *draw2dimg.GraphicContext, d chitDrawing) {
	i := d.Image
	c := d.Chit

	x := float64(c.X*i.GridWidth) + float64(i.GridWidth)/2.0 + float64(d.Offset.X)
	y := float64(c.Y*i.GridHeight) + float64(i.GridHeight)/2.0 + float64(d.Offset.Y)
	r := float64(d.Size) / 2.0

	gc.SetFillColor(d.Chit.Color)
	draw2dkit.Circle(gc, x, y, r)
	gc.Fill()
}
