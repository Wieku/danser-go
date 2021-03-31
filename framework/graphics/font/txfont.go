package font

import (
	"github.com/wieku/danser-go/app/utils"
	"github.com/wieku/danser-go/framework/graphics/texture"
	"math"
	"path/filepath"
	"strings"
)

func LoadTextureFont(path, name string, min, max rune, atlas *texture.TextureAtlas) *Font {
	font := new(Font)

	font.glyphs = make(map[rune]*glyphData)
	font.biggest = 40
	font.atlas = atlas

	extension := filepath.Ext(path)
	baseFile := strings.TrimSuffix(path, extension)

	for i := min; i <= max; i++ {
		region, _ := utils.LoadTextureToAtlas(font.atlas, baseFile+string(i)+extension)

		font.initialSize = math.Max(font.initialSize, float64(region.Height))

		font.glyphs[i] = &glyphData{region, float64(region.Width), 0, float64(region.Height) / 2}
	}

	setMeasures(font)

	return font
}

func LoadTextureFontMap(path, name string, chars map[string]rune, atlas *texture.TextureAtlas) *Font {
	font := new(Font)

	font.glyphs = make(map[rune]*glyphData)
	font.biggest = 40
	font.atlas = atlas

	extension := filepath.Ext(path)
	baseFile := strings.TrimSuffix(path, extension)

	for k, v := range chars {
		region, _ := utils.LoadTextureToAtlas(font.atlas, baseFile+k+extension)

		font.initialSize = math.Max(font.initialSize, float64(region.Height))

		font.glyphs[v] = &glyphData{region, float64(region.Width), 0, float64(region.Height) / 2}
	}

	setMeasures(font)

	return font
}

func LoadTextureFontMap2(chars map[rune]*texture.TextureRegion, overlap float64) *Font {
	font := new(Font)

	font.glyphs = make(map[rune]*glyphData)
	font.Overlap = overlap
	font.biggest = 40

	for c, r := range chars {
		if r == nil {
			continue
		}

		font.initialSize = math.Max(font.initialSize, float64(r.Height))

		font.glyphs[c] = &glyphData{r, float64(r.Width), 0, 0}
	}

	setMeasures(font)

	return font
}

func setMeasures(font *Font) {
	if glyph, exists := font.glyphs['5']; exists {
		font.biggest = glyph.advance
		font.initialSize = float64(glyph.region.Height)
	}
}
