package font

import (
	"github.com/wieku/danser-go/app/utils"
	"github.com/wieku/danser-go/framework/graphics/texture"
	"math"
	"path/filepath"
	"strings"
	"unicode"
)

func LoadTextureFont(path, name string, min, max rune, atlas *texture.TextureAtlas) *Font {
	font := new(Font)

	font.glyphs = make(map[rune]*glyphData)

	font.atlas = atlas

	extension := filepath.Ext(path)
	baseFile := strings.TrimSuffix(path, extension)

	for i := min; i <= max; i++ {
		region, _ := utils.LoadTextureToAtlas(font.atlas, baseFile+string(i)+extension)

		if float64(region.Height) > font.initialSize {
			font.initialSize = float64(region.Height)
		}

		font.glyphs[i] = &glyphData{region, float64(region.Width), 0, float64(region.Height) / 2}

		if unicode.IsDigit(i) {
			font.biggest = math.Max(font.biggest, float64(region.Width))
		}
	}

	return font
}

func LoadTextureFontMap(path, name string, chars map[string]rune, atlas *texture.TextureAtlas) *Font {
	font := new(Font)

	font.glyphs = make(map[rune]*glyphData) //, font.max-font.min+1)

	font.atlas = atlas

	extension := filepath.Ext(path)
	baseFile := strings.TrimSuffix(path, extension)

	for k, v := range chars {
		region, _ := utils.LoadTextureToAtlas(font.atlas, baseFile+k+extension)

		if float64(region.Height) > font.initialSize {
			font.initialSize = float64(region.Height)
		}

		font.glyphs[v] = &glyphData{region, float64(region.Width), 0, float64(region.Height) / 2}
		if unicode.IsDigit(v) {
			font.biggest = math.Max(font.biggest, float64(region.Width))
		}
	}

	return font
}

func LoadTextureFontMap2(chars map[rune]*texture.TextureRegion, overlap float64) *Font {
	font := new(Font)

	font.glyphs = make(map[rune]*glyphData)
	font.overlap = overlap

	for c, r := range chars {
		if r == nil {
			continue
		}

		if float64(r.Height) > font.initialSize {
			font.initialSize = float64(r.Height)
		}

		font.glyphs[c] = &glyphData{r, float64(r.Width), 0, float64(r.Height) / 2}

		if unicode.IsDigit(c) {
			font.biggest = math.Max(font.biggest, float64(r.Width))
		}
	}

	return font
}
