package font

import (
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/texture"
	"github.com/wieku/danser-go/framework/math/vector"
	font2 "golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"image"
	"image/draw"
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
	region                      *texture.TextureRegion
	advance, bearingX, bearingY float64
}

type Font struct {
	atlas       *texture.TextureAtlas
	glyphs      map[rune]*glyphData
	initialSize float64
	lineDist    float64
	kernTable   map[rune]map[rune]float64
	biggest     float64
	overlap     float64
}

func LoadFont(reader io.Reader) *Font {
	data, err := ioutil.ReadAll(reader)

	if err != nil {
		panic("Error reading font: " + err.Error())
	}

	ttf, err := truetype.Parse(data)
	if err != nil {
		panic("Error reading font: " + err.Error())
	}

	font := new(Font)
	font.initialSize = 100.0
	font.glyphs = make(map[rune]*glyphData) //, font.max-font.min+1)
	font.kernTable = make(map[rune]map[rune]float64)
	font.atlas = texture.NewTextureAtlas(1024, 4)

	font.atlas.Bind(20)

	fc := truetype.NewFace(ttf, &truetype.Options{Size: font.initialSize, DPI: 72, Hinting: font2.HintingFull})

	context := freetype.NewContext()
	context.SetFont(ttf)
	context.SetFontSize(font.initialSize)
	context.SetDPI(72)
	context.SetHinting(font2.HintingFull)

	for i := rune(0); i <= unicode.MaxRune; i++ {
		if ttf.Index(i) > 0 {
			gBnd, gAdv, ok := fc.GlyphBounds(i)
			if ok != true {
				continue
			}

			gh := int32((gBnd.Max.Y - gBnd.Min.Y) >> 6)
			gw := int32((gBnd.Max.X - gBnd.Min.X) >> 6)

			//if gylph has no diamensions set to a max value
			if gw == 0 || gh == 0 {
				gBnd = ttf.Bounds(fixed.Int26_6(20))
				gw = int32((gBnd.Max.X - gBnd.Min.X) >> 6)
				gh = int32((gBnd.Max.Y - gBnd.Min.Y) >> 6)

				//above can sometimes yield 0 for font smaller than 48pt, 1 is minimum
				if gw == 0 || gh == 0 {
					gw = 1
					gh = 1
				}
			}

			//The glyph's ascent and descent equal -bounds.Min.Y and +bounds.Max.Y.
			gAscent := int(-gBnd.Min.Y) >> 6
			//gdescent := int(gBnd.Max.Y) >> 6

			fg, bg := image.White, image.Transparent
			rect := image.Rect(0, 0, int(gw), int(gh))
			rgba := image.NewNRGBA(rect)
			draw.Draw(rgba, rgba.Bounds(), bg, image.ZP, draw.Src)

			context.SetClip(rect)
			context.SetDst(rgba)
			context.SetSrc(fg)

			px := 0 - (int(gBnd.Min.X) >> 6)
			py := (gAscent)
			pt := freetype.Pt(px, py)

			// Draw the text from mask to image
			_, err = context.DrawString(string(i), pt)
			if err != nil {
				log.Println(err)
				continue
			}

			//res := font.toSDF(rgba)

			newPix := make([]uint8, len(rgba.Pix))
			height := rgba.Bounds().Dy()

			for i := 0; i < height; i++ {
				copy(newPix[i*rgba.Stride:(i+1)*rgba.Stride], rgba.Pix[(height-1-i)*rgba.Stride:(height-i)*rgba.Stride])
			}

			region := font.atlas.AddTexture(string(i), rgba.Bounds().Dx(), rgba.Bounds().Dy(), newPix)

			//set w,h and adv, bearing V and bearing H in char
			advance := float64(gAdv) / 64
			/*font.biggest = math.Max(font.biggest, float64(advance))*/
			bearingV := float64(gBnd.Max.Y) / 64
			bearingH := float64(gBnd.Min.X) / 64
			font.glyphs[i] = &glyphData{region, advance, bearingH, bearingV}
		}
	}

	for i := range font.glyphs {
		for j := range font.glyphs {
			if font.kernTable[i] == nil {
				font.kernTable[i] = make(map[rune]float64)
			}

			font.kernTable[i][j] = float64(fc.Kern(i, j)) / 64
		}
	}

	font.biggest = font.glyphs['9'].advance

	fonts[ttf.Name(truetype.NameIDFontFullName)] = font

	log.Println(ttf.Name(truetype.NameIDFontFullName), "loaded!")

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

		if i > 0 || monospaced {
			advance -= font.overlap * scale * renderer.GetScale().X
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

		if i > 0 || monospaced {
			advance -= font.overlap
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
