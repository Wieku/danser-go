package font

import (
	"github.com/wieku/danser-go/render/texture"
	"github.com/wieku/danser-go/utils"
	"path/filepath"
	"strings"
)

func LoadTextureFont(path, name string, min, max rune, atlas *texture.TextureAtlas) *Font {
	font := new(Font)
	font.min = min
	font.max = max
	font.glyphs = make(map[rune]*glyphData) //, font.max-font.min+1)

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

func LoadTextureFontMap(path, name string, chars map[string]rune, atlas *texture.TextureAtlas) *Font {
	font := new(Font)

	min := rune(10000)
	max := rune(0)

	for _, v := range chars {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	font.min = min
	font.max = max
	font.glyphs = make(map[rune]*glyphData) //, font.max-font.min+1)

	font.atlas = atlas

	extension := filepath.Ext(path)
	baseFile := strings.TrimSuffix(path, extension)

	for k, v := range chars {
		region, _ := utils.LoadTextureToAtlas(font.atlas, baseFile+k+extension)

		if float64(region.Height) > font.initialSize {
			font.initialSize = float64(region.Height)
		}

		font.glyphs[v-font.min] = &glyphData{*region, float64(region.Width), 0, float64(region.Height) / 2}
	}

	return font
}
