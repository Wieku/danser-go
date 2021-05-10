package font

import (
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/texture"
	color2 "github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/vector"
	"unicode"
)

var fonts map[string]*Font

func init() {
	fonts = make(map[string]*Font)
}

func GetFont(name string) *Font {
	return fonts[name]
}

type glyphData struct {
	region  *texture.TextureRegion
	advance float64
	offsetX float64
	offsetY float64
}

type Font struct {
	atlas       *texture.TextureAtlas
	glyphs      map[rune]*glyphData
	initialSize float64
	kernTable   map[rune]map[rune]float64
	biggest     float64
	Overlap     float64
	Ascent      float64
	Descent     float64
	flip        bool
}

func (font *Font) drawInternal(renderer *batch.QuadBatch, x, y float64, size float64, text string, monospaced bool) {
	scale := size / font.initialSize

	scBase := scale * renderer.GetScale().Y * renderer.GetSubScale().Y

	scl := vector.NewVec2d(scale, scale)

	col := color2.NewL(1)

	advance := 0.0

	for i, c := range text {
		char := font.glyphs[c]
		if char == nil {
			continue
		}

		if i > 0 && font.kernTable != nil && !(monospaced && (unicode.IsDigit(c) || unicode.IsDigit(rune(text[i-1])))) {
			advance += font.kernTable[rune(text[i-1])][c]
		}

		if i > 0 {
			advance -= font.Overlap
		}

		bearX := char.offsetX

		if unicode.IsDigit(c) && monospaced {
			bearX = (font.biggest - float64(char.region.Width)) / 2
		}

		pos := vector.NewVec2d((advance+bearX)*scBase+x, char.offsetY*scBase+y)

		renderer.DrawStObject(pos, bmath.Origin.TopLeft, scl, false, font.flip, 0, col, false, *char.region)

		if monospaced && (unicode.IsDigit(c) || unicode.IsSpace(c)) {
			advance += font.biggest
		} else {
			advance += char.advance
		}
	}
}

func (font *Font) Draw(renderer *batch.QuadBatch, x, y float64, size float64, text string) {
	font.DrawOrigin(renderer, x, y, bmath.Origin.BottomLeft, size, false, text)
}

func (font *Font) DrawMonospaced(renderer *batch.QuadBatch, x, y float64, size float64, text string) {
	font.DrawOrigin(renderer, x, y, bmath.Origin.BottomLeft, size, true, text)
}

func (font *Font) getWidthInternal(size float64, text string, monospaced bool) float64 {
	advance := 0.0

	scale := size / font.initialSize

	for i, c := range text {
		char := font.glyphs[c]
		if char == nil {
			continue
		}

		if i > 0 && font.kernTable != nil && !(monospaced && (unicode.IsDigit(c) || unicode.IsDigit(rune(text[i-1])))) {
			advance += font.kernTable[rune(text[i-1])][c]
		}

		if i > 0 {
			advance -= font.Overlap
		}

		if monospaced && (unicode.IsDigit(c) || unicode.IsSpace(c)) {
			advance += font.biggest
		} else {
			advance += char.advance
		}
	}

	return advance * scale
}

func (font *Font) GetWidth(size float64, text string) float64 {
	return font.getWidthInternal(size, text, false)
}

func (font *Font) GetWidthMonospaced(size float64, text string) float64 {
	return font.getWidthInternal(size, text, true)
}

func (font *Font) GetSize() float64 {
	return font.initialSize
}

func (font *Font) DrawOrigin(renderer *batch.QuadBatch, x, y float64, origin vector.Vector2d, size float64, monospaced bool, text string) {
	width := font.getWidthInternal(size, text, monospaced)
	align := origin.AddS(1, 1).Mult(vector.NewVec2d(-width/2, -size/2)).Mult(renderer.GetScale()).Mult(renderer.GetSubScale())

	font.drawInternal(renderer, x+align.X, y+align.Y, size, text, monospaced)
}

func (font *Font) DrawOriginV(renderer *batch.QuadBatch, position vector.Vector2d, origin vector.Vector2d, size float64, monospaced bool, text string) {
	font.DrawOrigin(renderer, position.X, position.Y, origin, size, monospaced, text)
}
