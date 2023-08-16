package font

import (
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/texture"
	color2 "github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
	"unicode"
)

var fonts map[string]*Font

func init() {
	fonts = make(map[string]*Font)
}

func GetFont(name string) *Font {
	return fonts[name]
}

func AddAlias(font *Font, name string) {
	fonts[name] = font
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
	ascent      float64
	flip        bool
	pixel       *texture.TextureRegion

	drawBg   bool
	bgBorder float64
	bgColor  color2.Color
}

func (font *Font) drawInternal(renderer *batch.QuadBatch, x, y, width, size, rotation float64, text string, monospaced bool, color color2.Color) {
	rotation += renderer.GetRotation()

	if font.drawBg && font.pixel != nil {
		pos2 := vector.NewVec2d(-font.bgBorder, -font.bgBorder).Rotate(rotation).AddS(x, y)
		vSize := vector.NewVec2d(width+2*font.bgBorder, size+2*font.bgBorder)

		renderer.DrawStObject(pos2, vector.TopLeft, vSize, false, false, rotation, font.bgColor, false, *font.pixel)
	}

	scale := size / font.initialSize

	scBase := scale * renderer.GetScale().Y * renderer.GetSubScale().Y

	scl := vector.NewVec2d(scale, scale)

	advance := 0.0

	cos := math.Cos(rotation)
	sin := math.Sin(rotation)

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

		pX := (advance + bearX) * scBase
		pY := char.offsetY * scBase

		pos := vector.NewVec2d(pX*cos-pY*sin+x, pX*sin+pY*cos+y)

		renderer.DrawStObject(pos, vector.TopLeft, scl, false, font.flip, rotation, color, false, *char.region)

		if monospaced && (unicode.IsDigit(c) || unicode.IsSpace(c)) {
			advance += font.biggest
		} else {
			advance += char.advance
		}
	}
}

func (font *Font) Draw(renderer *batch.QuadBatch, x, y, size float64, text string) {
	font.DrawOrigin(renderer, x, y, vector.BottomLeft, size, false, text)
}

func (font *Font) DrawMonospaced(renderer *batch.QuadBatch, x, y, size float64, text string) {
	font.DrawOrigin(renderer, x, y, vector.BottomLeft, size, true, text)
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

func (font *Font) GetAscent() float64 {
	return font.ascent
}

func (font *Font) DrawBg(v bool) {
	font.drawBg = v
}

func (font *Font) SetBgBorderSize(size float64) {
	font.bgBorder = size
}

func (font *Font) SetBgColor(color color2.Color) {
	font.bgColor = color
}

func (font *Font) DrawOrigin(renderer *batch.QuadBatch, x, y float64, origin vector.Vector2d, size float64, monospaced bool, text string) {
	font.DrawOriginRotation(renderer, x, y, origin, size, 0, monospaced, text)
}

func (font *Font) DrawOriginRotation(renderer *batch.QuadBatch, x, y float64, origin vector.Vector2d, size, rotation float64, monospaced bool, text string) {
	font.DrawOriginRotationColor(renderer, x, y, origin, size, rotation, monospaced, color2.NewL(1), text)
}

func (font *Font) DrawOriginRotationColor(renderer *batch.QuadBatch, x, y float64, origin vector.Vector2d, size, rotation float64, monospaced bool, color color2.Color, text string) {
	width := font.getWidthInternal(size, text, monospaced)
	align := origin.AddS(1, 1).Mult(vector.NewVec2d(-width/2, -(size/font.initialSize*font.ascent)/2)).Mult(renderer.GetScale()).Mult(renderer.GetSubScale()).Rotate(rotation)

	font.drawInternal(renderer, x+align.X, y+align.Y, width, size, rotation, text, monospaced, color)
}

func (font *Font) DrawOriginV(renderer *batch.QuadBatch, position vector.Vector2d, origin vector.Vector2d, size float64, monospaced bool, text string) {
	font.DrawOrigin(renderer, position.X, position.Y, origin, size, monospaced, text)
}

func (font *Font) DrawOriginRotationV(renderer *batch.QuadBatch, position vector.Vector2d, origin vector.Vector2d, size, rotation float64, monospaced bool, text string) {
	font.DrawOriginRotation(renderer, position.X, position.Y, origin, size, rotation, monospaced, text)
}

func (font *Font) DrawOriginRotationColorV(renderer *batch.QuadBatch, position vector.Vector2d, origin vector.Vector2d, size, rotation float64, monospaced bool, color color2.Color, text string) {
	font.DrawOriginRotationColor(renderer, position.X, position.Y, origin, size, rotation, monospaced, color, text)
}
