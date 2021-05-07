package font

import (
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/texture"
	"github.com/wieku/danser-go/framework/math/vector"
	font2 "golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
	"image"
	"io"
	"io/ioutil"
	"log"
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
}

func LoadFont(reader io.Reader) *Font {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		panic("Error reading font: " + err.Error())
	}

	ttf, err := opentype.Parse(data)
	if err != nil {
		panic("Error reading font: " + err.Error())
	}

	font := new(Font)
	font.initialSize = 64.0
	font.glyphs = make(map[rune]*glyphData)
	font.kernTable = make(map[rune]map[rune]float64)

	font.atlas = texture.NewTextureAtlas(1024, 4)
	font.atlas.SetManualMipmapping(true)

	fc, err := opentype.NewFace(ttf, &opentype.FaceOptions{Size: font.initialSize, DPI: 72, Hinting: font2.HintingFull})
	if err != nil {
		panic("Error reading font: " + err.Error())
	}

	defer fc.Close()

	buff := &sfnt.Buffer{}

	for i := rune(0); i <= unicode.MaxRune; i++ {
		if idx, _ := ttf.GlyphIndex(buff, i); idx > 0 {
			b, gAdv, _ := fc.GlyphBounds(i)
			w, h := (b.Max.X - b.Min.X).Ceil(), (b.Max.Y - b.Min.Y).Ceil()

			if w == 0 || h == 0 {
				b, _ = ttf.Bounds(buff, fixed.Int26_6(20), font2.HintingFull)

				w, h = (b.Max.X - b.Min.X).Ceil(), (b.Max.Y - b.Min.Y).Ceil()

				if w == 0 || h == 0 {
					w = 1
					h = 1
				}
			}

			if b.Min.X&((1<<6)-1) != 0 {
				w++
			}

			if b.Min.Y&((1<<6)-1) != 0 {
				h++
			}

			pixmap := texture.NewPixMap(w, h)

			d := font2.Drawer{
				Dst:  pixmap.NRGBA(),
				Src:  image.White,
				Face: fc,
			}

			x, y := fixed.I((-b.Min.X).Ceil()), fixed.I((-b.Min.Y).Ceil())
			d.Dot = fixed.Point26_6{X: x, Y: y}
			d.DrawString(string(i))

			region := font.atlas.AddTexture(string(i), pixmap.Width, pixmap.Height, pixmap.Data)

			region.V1, region.V2 = region.V2, region.V1

			pixmap.Dispose()

			//set w,h and adv, bearing V and bearing H in char
			advance := float64(gAdv) / 64
			bearingV := float64(b.Max.Y) / 64
			bearingH := float64(b.Min.X) / 64

			font.glyphs[i] = &glyphData{region, advance, bearingH, bearingV}
		}
	}

	font.atlas.GenerateMipmaps()

	for i := range font.glyphs {
		font.kernTable[i] = make(map[rune]float64)

		for j := range font.glyphs {
			if krn := fc.Kern(i, j); krn != 0 {
				font.kernTable[i][j] = float64(krn) / 64
			}
		}
	}

	font.biggest = font.glyphs['5'].advance

	name, _ := ttf.Name(buff, sfnt.NameIDFull)

	fonts[name] = font

	log.Println(name, "loaded!")

	return font
}

func (font *Font) drawInternal(renderer *batch.QuadBatch, x, y float64, size float64, text string, monospaced bool) {
	advance := x

	scale := size / font.initialSize

	for i, c := range text {
		char := font.glyphs[c]
		if char == nil {
			continue
		}

		if i > 0 && font.kernTable != nil && !(monospaced && (unicode.IsDigit(c) || unicode.IsDigit(rune(text[i-1])))) {
			advance += font.kernTable[rune(text[i-1])][c] * scale * renderer.GetScale().X
		}

		if i > 0 {
			advance -= font.Overlap * scale * renderer.GetScale().X
		}

		renderer.SetSubScale(scale, scale)

		bearX := char.bearingX

		if unicode.IsDigit(c) && monospaced {
			bearX = (font.biggest - float64(char.region.Width)) / 2
		}

		tr := vector.NewVec2d(bearX+float64(char.region.Width)/2, float64(char.region.Height)/2-char.bearingY).Scl(scale).Mult(renderer.GetScale()).AddS(advance, y)

		renderer.SetTranslation(tr)

		renderer.DrawTexture(*char.region)

		xAdv := char.advance

		if monospaced && (unicode.IsDigit(c) || unicode.IsSpace(c)) {
			xAdv = font.biggest
		}

		advance += scale * renderer.GetScale().X * xAdv
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
	align := origin.AddS(1, 1).Mult(vector.NewVec2d(-width/2, -size/2)).Mult(renderer.GetScale())

	font.drawInternal(renderer, x+align.X, y+align.Y, size, text, monospaced)
}

func (font *Font) DrawOriginV(renderer *batch.QuadBatch, position vector.Vector2d, origin vector.Vector2d, size float64, monospaced bool, text string) {
	font.DrawOrigin(renderer, position.X, position.Y, origin, size, monospaced, text)
}
