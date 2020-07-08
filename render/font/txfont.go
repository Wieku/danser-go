package font

import (
	"github.com/wieku/danser-go/render/texture"
	"path/filepath"
	"strings"
	"github.com/wieku/danser-go/utils"
)

func LoadTextureFont(path, name string, min, max rune, atlas *texture.TextureAtlas) *Font {
	font := new(Font)
	font.min = min
	font.max = max
	font.glyphs = make(map[rune]*glyphData)//, font.max-font.min+1)

	font.atlas = atlas

	extension := filepath.Ext(path)
	baseFile := strings.TrimSuffix(path, extension)

	for i := min; i <= max; i++ {
		region, _ := utils.LoadTextureToAtlas(font.atlas, baseFile+string(i)+extension)

		if float64(region.Height) > font.initialSize {
			font.initialSize = float64(region.Height)
		}

		font.glyphs[i-font.min] = &glyphData{*region, float64(region.Width), 0, float64(region.Height) / 2}
	}

	return font
}
