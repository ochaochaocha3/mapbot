package mapgen

import (
	"fmt"
	"io/ioutil"

	"github.com/golang/freetype/truetype"
	"github.com/llgcode/draw2d"
)

const (
	// fontNameForMap は、マップ用に使用するフォントの名前。
	fontNameForMap = "normal"
)

// stringFontMap は文字列 -> フォントデータの対応の型。
type stringFontMap map[string]*truetype.Font

// FontCache はフォントデータを格納する構造体。
type FontCache struct {
	// fontMap はフォント名 -> フォントデータの対応。
	fontMap stringFontMap
}

// NewFontCache は新しいフォントデータの格納先を返す。
func NewFontCache() *FontCache {
	return &FontCache{
		fontMap: stringFontMap{},
	}
}

// Store はフォントキャッシュにフォントデータを格納する。
func (fc *FontCache) Store(fd draw2d.FontData, font *truetype.Font) {
	fc.fontMap[fd.Name] = font
}

// Load はフォントキャッシュからフォントデータを読み出す。
func (fc *FontCache) Load(fd draw2d.FontData) (*truetype.Font, error) {
	font, stored := fc.fontMap[fd.Name]
	if !stored {
		return nil, fmt.Errorf("font not found: %s", fd.Name)
	}

	return font, nil
}

// フォントファイルからデータを読み、格納する。
func (fc *FontCache) StoreFontDataFromFile(fontPath string) error {
	b, err := ioutil.ReadFile(fontPath)
	if err != nil {
		return err
	}

	font, err := truetype.Parse(b)
	if err != nil {
		return err
	}

	fc.Store(draw2d.FontData{Name: fontNameForMap}, font)

	return nil
}
