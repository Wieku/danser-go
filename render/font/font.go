package font

import (
	"io"
	"github.com/golang/freetype/truetype"
	"io/ioutil"
	"golang.org/x/image/math/fixed"
	"github.com/golang/freetype"
	font2 "golang.org/x/image/font"
	"image"
	"image/draw"
	"log"
	"github.com/wieku/danser/render/texture"
	"github.com/wieku/danser/bmath"
	"github.com/wieku/danser/render/batches"
)

var fonts map[string]*Font

func init() {
	fonts = make(map[string]*Font)
}

func GetFont(name string) *Font {
	return fonts[name]
}

type glyphData struct {
	region                      texture.TextureRegion
	advance, bearingX, bearingY float64
}

type Font struct {
	atlas       *texture.TextureAtlas
	glyphs      []*glyphData
	min, max    rune
	initialSize float64
	lineDist    float64
	kernTable   map[rune]map[rune]float64
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
	font.min = rune(32)
	font.max = rune(127)
	font.initialSize = 64.0
	font.glyphs = make([]*glyphData, font.max-font.min+1)
	font.kernTable = make(map[rune]map[rune]float64)
	font.atlas = texture.NewTextureAtlas(4096, 4)

	font.atlas.Bind(20)

	fc := truetype.NewFace(ttf, &truetype.Options{Size: font.initialSize, DPI: 72, Hinting: font2.HintingFull})

	for i := font.min; i <= font.max; i++ {
		for j := i; j <= font.max; j++ {
			if font.kernTable[i] == nil {
				font.kernTable[i] = make(map[rune]float64)
			}

			if font.kernTable[j] == nil {
				font.kernTable[j] = make(map[rune]float64)
			}

			font.kernTable[i][j] = float64(fc.Kern(i, j)) / 64
			font.kernTable[j][i] = float64(fc.Kern(j, i)) / 64
		}
	}

	context := freetype.NewContext()
	context.SetFont(ttf)
	context.SetFontSize(font.initialSize)
	context.SetDPI(72)
	context.SetHinting(font2.HintingFull)

	for i := font.min; i <= font.max; i++ {

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
		bearingV := float64(gBnd.Max.Y) / 64
		bearingH := float64(gBnd.Min.X) / 64
		font.glyphs[i-font.min] = &glyphData{*region, advance, bearingH, bearingV}
	}

	log.Println(ttf.Name(truetype.NameIDFontFullName), "loaded!")
	fonts[ttf.Name(truetype.NameIDFontFullName)] = font
	return font
}

func (font *Font) Draw(renderer *batches.SpriteBatch, x, y float64, size float64, text string) {
	xpad := x

	scale := size / font.initialSize

	for i, c := range text {
		char := font.glyphs[c-font.min]
		if char == nil {
			continue
		}

		kerning := 0.0

		if i > 0 && font.kernTable != nil {
			kerning = font.kernTable[rune(text[i-1])][c]
		}

		renderer.SetSubScale(scale, scale)
		renderer.SetTranslation(bmath.NewVec2d(xpad+(char.bearingX-kerning+float64(char.region.Width)/2)*scale*renderer.GetScale().X, y+(float64(char.region.Height)/2-char.bearingY)*scale* renderer.GetScale().Y))
		renderer.DrawTexture(char.region)
		xpad += scale * renderer.GetScale().X * (char.advance - kerning)

	}
}

func (font *Font) GetWidth(size float64, text string) float64 {
	scale := size / font.initialSize
	xpad := 0.0
	for i, c := range text {
		char := font.glyphs[c-font.min]
		if char == nil {
			continue
		}

		kerning := 0.0
		if i > 0 && font.kernTable != nil {
			kerning = font.kernTable[rune(text[i-1])][c]
		}

		xpad += scale * (char.advance - kerning)
	}
	return xpad
}

func (font *Font) DrawCentered(renderer *batches.SpriteBatch, x, y float64, size float64, text string) {
	xpad := font.GetWidth(size, text) * renderer.GetScale().X
	font.Draw(renderer, x-(xpad)/2, y, size, text)
}
