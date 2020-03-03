package mapgen

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io/ioutil"
	"math"

	"github.com/golang/freetype/truetype"
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

// MyFontCache は自前のフォントキャッシュの型。
type MyFontCache map[string]*truetype.Font

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
func (i *SquareMapImage) Render() (*image.RGBA, error) {
	mapImg := image.NewRGBA(i.rect)
	mapGC := draw2dimg.NewGraphicContext(mapImg)

	fillBackGround(mapGC, i)
	drawGrid(mapGC, i)
	drawChits(mapGC, i)

	legend, legendErr := drawLegend(i)
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
func fillBackGround(gc *draw2dimg.GraphicContext, i *SquareMapImage) {
	// 背景色で塗る
	gc.SetFillColor(i.BackgroundColor)
	draw2dkit.Rectangle(gc, 0, 0, float64(i.Width()), float64(i.Height()))
	gc.Fill()
}

// drawGrid はgcにグリッドを描画する。
func drawGrid(gc *draw2dimg.GraphicContext, img *SquareMapImage) {
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

// Store はフォントキャッシュにフォントデータを格納する。
func (fc MyFontCache) Store(fd draw2d.FontData, font *truetype.Font) {
	fc[fd.Name] = font
}

// Load はフォントキャッシュからフォントデータを読み出す。
func (fc MyFontCache) Load(fd draw2d.FontData) (*truetype.Font, error) {
	font, stored := fc[fd.Name]
	if !stored {
		return nil, fmt.Errorf("font not found: %s", fd.Name)
	}

	return font, nil
}

// drawLegend は凡例を描画する。
func drawLegend(mImg *SquareMapImage) (image.Image, error) {
	img := image.NewRGBA(image.Rect(0, 0, mImg.Width(), mImg.Map.NumOfChits()*mImg.GridHeight))
	gc := draw2dimg.NewGraphicContext(img)

	size := math.Min(float64(mImg.GridWidth), float64(mImg.GridHeight)) / 2.0
	fontSize := 0.8 * size

	fontCache, err := setupFontCache()
	if err != nil {
		return nil, err
	}

	gc.FontCache = fontCache
	gc.SetFontData(draw2d.FontData{Name: "gothic"})
	gc.SetFontSize(fontSize)

	// 背景色で塗る
	gc.SetFillColor(mImg.BackgroundColor)
	draw2dkit.Rectangle(gc, 0, 0, float64(img.Rect.Dx()), float64(img.Rect.Dy()))
	gc.Fill()

	// 凡例の各行を描画する
	x := float64(mImg.GridWidth) / 2.0
	xLabel := float64(mImg.GridWidth)
	r := size / 2.0
	for i, c := range mImg.Map.Chits() {
		y := float64(i*mImg.GridHeight) + float64(mImg.GridHeight)/2.0
		gc.SetFillColor(c.Color)
		draw2dkit.Circle(gc, x, y, r)
		gc.Fill()

		yLabel := y + fontSize/2
		gc.SetFillColor(colorutil.CSS3NameToRGBA("black"))
		gc.FillStringAt(c.Name, xLabel, yLabel)
	}

	return img, nil
}

var fontFile = "TakaoPGothic.ttf"

// setupFontCache はフォントキャッシュを用意する。
func setupFontCache() (MyFontCache, error) {
	b, err := ioutil.ReadFile(fontFile)
	if err != nil {
		return nil, err
	}

	font, err := truetype.Parse(b)
	if err != nil {
		return nil, err
	}

	fontCache := MyFontCache{}
	fontCache.Store(draw2d.FontData{Name: "gothic"}, font)

	return fontCache, nil
}
