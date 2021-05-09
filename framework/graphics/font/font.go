package font

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/texture"
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
	region   *texture.TextureRegion
	advance  float64
	bearingX float64
	bearingY float64
}

type Font struct {
	atlas       *texture.TextureAtlas
	glyphs      map[rune]*glyphData
	initialSize float64
	lineDist    float64
	kernTable   map[rune]map[rune]float64
	biggest     float64
	Overlap     float64
	Ascent      float64
	Descent     float64
}

func (font *Font) drawInternal(renderer *batch.QuadBatch, x, y float64, size float64, text string, monospaced bool) {
	scale := size / font.initialSize
	scl := vector.NewVec2d(scale, scale)

	batchScl := renderer.GetScale().Mult(renderer.GetSubScale())
	col := mgl32.Vec4{1, 1, 1, 1}

	posX := 0.0

	for i, c := range text {
		char := font.glyphs[c]
		if char == nil {
			continue
		}

		if i > 0 && font.kernTable != nil && !(monospaced && (unicode.IsDigit(c) || unicode.IsDigit(rune(text[i-1])))) {
			posX += font.kernTable[rune(text[i-1])][c] * scale * batchScl.X
		}

		if i > 0 {
			posX -= font.Overlap * scale * batchScl.X
		}

		bearX := char.bearingX

		if unicode.IsDigit(c) && monospaced {
			bearX = (font.biggest - float64(char.region.Width)) / 2
		}

		tr := vector.NewVec2d(bearX, -char.bearingY).Scl(scale).Mult(batchScl).AddS(posX+x, y)

		renderer.DrawStObject(tr, bmath.Origin.TopLeft, scl, false, false, 0, col, false, *char.region)

		xAdv := char.advance

		if monospaced && (unicode.IsDigit(c) || unicode.IsSpace(c)) {
			xAdv = font.biggest
		}

		posX += scale * renderer.GetScale().X * xAdv
	}
}

func (font *Font) Draw(renderer *batch.QuadBatch, x, y float64, size float64, text string) {
	font.drawInternal(renderer, x, y, size, text, false)
}

func (font *Font) DrawCentered(renderer *batch.QuadBatch, x, y float64, size float64, text string) {
	xpad := font.GetWidth(size, text) * renderer.GetScale().X
	font.Draw(renderer, x-(xpad)/2, y, size, text)
}

func (font *Font) DrawMonospaced(renderer *batch.QuadBatch, x, y float64, size float64, text string) {
	font.drawInternal(renderer, x, y, size, text, true)
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

		xAdv := char.advance

		if monospaced && (unicode.IsDigit(c) || unicode.IsSpace(c)) {
			xAdv = font.biggest
		}

		advance += xAdv
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
