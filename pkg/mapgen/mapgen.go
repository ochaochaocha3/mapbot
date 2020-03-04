package mapgen

import (
	"image"
	"image/color"
	"image/draw"
	"math"

	"github.com/llgcode/draw2d"
	"github.com/llgcode/draw2d/draw2dimg"
	"github.com/llgcode/draw2d/draw2dkit"

	"github.com/ochaochaocha3/mapbot/pkg/colorutil"
	"github.com/ochaochaocha3/mapbot/pkg/rpgmap"
)

// SquareMapImage はスクエアマップの画像描画に必要な情報の構造体。
type SquareMapImage struct {
	// Map は描画対象のスクエアマップ。
	Map *rpgmap.SquareMap
	// FontCache はフォントの格納先。
	FontCache *FontCache
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
func NewSquareMapImage(m *rpgmap.SquareMap, fc *FontCache) *SquareMapImage {
	i := &SquareMapImage{
		Map:             m,
		FontCache:       fc,
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
func (i *SquareMapImage) Render() (*image.RGBA, error) {
	mapImg := image.NewRGBA(i.rect)
	mapGC := draw2dimg.NewGraphicContext(mapImg)

	i.fillBackGround(mapGC)
	i.drawGrid(mapGC)
	i.drawChits(mapGC)

	legend, legendErr := i.drawLegend()
	if legendErr != nil {
		return nil, legendErr
	}

	legendSP := image.Point{0, mapImg.Bounds().Dy()}
	legendRect := image.Rectangle{legendSP, legendSP.Add(legend.Bounds().Size())}

	r := image.Rectangle{image.ZP, legendRect.Max}

	dest := image.NewRGBA(r)

	draw.Draw(dest, mapImg.Bounds(), mapImg, image.ZP, draw.Src)
	draw.Draw(dest, legendRect, legend, image.ZP, draw.Src)

	return dest, nil
}

// 描画領域の矩形を更新する。
func (i *SquareMapImage) updateRect() {
	width := i.GridWidth * i.Map.Width()
	height := i.GridHeight * i.Map.Height()

	i.rect = image.Rect(0, 0, width, height)
}

// fillBackGround はgcを背景色で塗りつぶす。
func (i *SquareMapImage) fillBackGround(gc *draw2dimg.GraphicContext) {
	// 背景色で塗る
	gc.SetFillColor(i.BackgroundColor)
	draw2dkit.Rectangle(gc, 0, 0, float64(i.Width()), float64(i.Height()))
	gc.Fill()
}

// drawGrid はgcにグリッドを描画する。
func (img *SquareMapImage) drawGrid(gc *draw2dimg.GraphicContext) {
	gc.SetStrokeColor(img.GridColor)
	gc.SetLineWidth(1.0)

	for i := 0; i < img.Height(); i++ {
		y := float64(i * img.GridHeight)
		gc.MoveTo(0, y)
		gc.LineTo(float64(img.Width()), y)
		gc.Stroke()
	}

	for j := 0; j < img.Width(); j++ {
		x := float64(j * img.GridWidth)
		gc.MoveTo(x, 0)
		gc.LineTo(x, float64(img.Height()))
		gc.Stroke()
	}
}

// drawChits はgcにチットの集合を描画する。
//
// TODO: 同じ座標の場合にチットの位置をずらす。
func (i *SquareMapImage) drawChits(gc *draw2dimg.GraphicContext) {
	chitSize := int(math.Min(float64(i.GridWidth), float64(i.GridHeight))) / 2
	offset := image.Point{X: 0, Y: 0}

	i.Map.ForEachChit(func(_ int, c *rpgmap.Chit) {
		i.drawChit(gc, c, chitSize, offset)
	})
}

// chitDrawing はチット描画の情報。
type chitDrawing struct {
	// Chit は描画対象チット。
	Chit *rpgmap.Chit
	// Size は描画するチットの大きさ（辺の長さ）。
	Size int
	// Offset は座標のずれ。
	Offset image.Point
}

// drawChit はgcにチットを描画する。
func (i *SquareMapImage) drawChit(
	gc *draw2dimg.GraphicContext,
	chit *rpgmap.Chit,
	size int,
	offset image.Point,
) {
	x := float64(chit.X*i.GridWidth) + float64(i.GridWidth)/2.0 + float64(offset.X)
	y := float64(chit.Y*i.GridHeight) + float64(i.GridHeight)/2.0 + float64(offset.Y)
	r := float64(size) / 2.0

	gc.SetFillColor(chit.Color)
	draw2dkit.Circle(gc, x, y, r)
	gc.Fill()
}

// drawLegend は凡例を描画する。
func (mImg *SquareMapImage) drawLegend() (image.Image, error) {
	img := image.NewRGBA(image.Rect(0, 0, mImg.Width(), mImg.Map.NumOfChits()*mImg.GridHeight))
	gc := draw2dimg.NewGraphicContext(img)

	size := math.Min(float64(mImg.GridWidth), float64(mImg.GridHeight)) / 2.0
	fontSize := 0.8 * size

	gc.FontCache = mImg.FontCache
	gc.SetFontData(draw2d.FontData{Name: fontNameForMap})
	gc.SetFontSize(fontSize)

	// 背景色で塗る
	gc.SetFillColor(mImg.BackgroundColor)
	draw2dkit.Rectangle(gc, 0, 0, float64(img.Rect.Dx()), float64(img.Rect.Dy()))
	gc.Fill()

	// 凡例の各行を描画する
	x := float64(mImg.GridWidth) / 2.0
	xLabel := float64(mImg.GridWidth)
	r := size / 2.0
	mImg.Map.ForEachChit(func(i int, c *rpgmap.Chit) {
		y := float64(i*mImg.GridHeight) + float64(mImg.GridHeight)/2.0
		gc.SetFillColor(c.Color)
		draw2dkit.Circle(gc, x, y, r)
		gc.Fill()

		yLabel := y + fontSize/2
		gc.SetFillColor(colorutil.CSS3NameToRGBA("black"))
		gc.FillStringAt(c.Name, xLabel, yLabel)
	})

	return img, nil
}
