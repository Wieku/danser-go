package font

import (
	"github.com/wieku/danser/render"
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
	"math"
	"image/color"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser/bmath"
)

type glyphData struct {
	region texture.TextureRegion
	advance, bearingX, bearingY float64
}

type Font struct {
	face font2.Face
	atlas *texture.TextureAtlas
	glyphs []*glyphData
	min, max rune
	padding int
	renderer *render.SpriteBatch
}

func NewFont(reader io.Reader) *Font {
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
	font.padding = 12
	font.glyphs = make([]*glyphData, font.max-font.min+1)

	font.atlas = texture.NewTextureAtlas(4096, 4)

	font.atlas.Bind(20)

	fc := truetype.NewFace(ttf, &truetype.Options{Size: 128, DPI: 72, Hinting: font2.HintingFull})
	font.face = fc
	context := freetype.NewContext()
	context.SetFont(ttf)
	context.SetFontSize(float64(128))
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

		//set w,h and adv, bearing V and bearing H in char
		advance := float64(gAdv)/64
		bearingV := float64(gBnd.Max.Y)/64
		bearingH := float64(gBnd.Min.X)/64

		fg, bg := image.White, image.Transparent
		rect := image.Rect(0, 0, int(gw)+2*font.padding, int(gh)+2*font.padding)
		rgba := image.NewNRGBA(rect)
		draw.Draw(rgba, rgba.Bounds(), bg, image.ZP, draw.Src)

		context.SetClip(rgba.Bounds())
		context.SetDst(rgba)
		context.SetSrc(fg)

		px := 0 - (int(gBnd.Min.X) >> 6)
		py := (gAscent)
		pt := freetype.Pt(px+font.padding, py+font.padding)

		// Draw the text from mask to image
		_, err = context.DrawString(string(i), pt)
		if err != nil {
			log.Println(err)
			continue
		}

		region := font.atlas.AddTexture(string(i), rgba.Bounds().Dx(),rgba.Bounds().Dy(), font.toSDF(*rgba).Pix)
		font.glyphs[i-font.min] = &glyphData{*region, advance, bearingH, bearingV}
	}

	font.renderer = render.NewSpriteBatch(true)

	return font
}

func (font *Font) Draw(x, y float64, proj mgl32.Mat4, size float64, text string) {
	font.renderer.Begin()
	font.renderer.SetCamera(proj)
	xpad := x

	scale := size/128.0

	for i, c := range text {
		char := font.glyphs[c-font.min]
		if char == nil {
			continue
		}

		kerning := 0.0

		if i > 0 {
			kerning = float64(font.face.Kern(rune(text[i-1]), c))/64
		}

		font.renderer.SetScale(scale, scale)
		font.renderer.SetTranslation(bmath.NewVec2d(xpad+(-float64(font.padding)+char.bearingX-kerning+float64(char.region.Width-2*int32(font.padding))/2)*scale, y+(float64(char.region.Height-2*int32(font.padding))/2-char.bearingY)*scale))// y-(float64(char.region.Height-2*int32(font.padding))-char.bearingY)))
		font.renderer.DrawTexture(char.region)
		xpad += scale*(char.advance-kerning)

	}

	font.renderer.End()
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (font *Font) toSDF(nrgba image.NRGBA) image.NRGBA {
	bitmap := make([][]bool, nrgba.Bounds().Dx())
	dst := image.NewNRGBA(nrgba.Rect)
	for x:=0; x < nrgba.Bounds().Dx(); x++ {
		bitmap[x] = make([]bool, nrgba.Bounds().Dy())
		for y:=0; y < nrgba.Bounds().Dy(); y++ {
			if nrgba.NRGBAAt(x, y).A > 127 {
				bitmap[x][y] = true
			} else {
				bitmap[x][y] = false
			}
		}
	}

	spread := float64(font.padding)

	for x:=0; x < nrgba.Bounds().Dx(); x++ {
		for y:=0; y < nrgba.Bounds().Dy(); y++ {
			iSpread := int(math.Floor(spread))
			nearest := iSpread*iSpread
			for dx:= max(0, x - iSpread); dx < min(nrgba.Bounds().Dx()-1, x + iSpread); dx++ {
				for dy:= max(0, y - iSpread); dy < min(nrgba.Bounds().Dy()-1, y + iSpread); dy++ {
					if bitmap[x][y] == bitmap[dx][dy] {
						continue
					}

					dst := (dx-x)*(dx-x) + (dy-y)*(dy-y)
					if dst < nearest {
						nearest = dst
					}
				}
			}

			dsf := 0.5 * math.Sqrt(float64(nearest))/spread

			if !bitmap[x][y] {
				dsf *= -1
			}

			dsf += 0.5

			dst.Set(x, nrgba.Bounds().Dy()-y-1, color.NRGBA{255, 255, 255, uint8(dsf*255)})
		}
	}

	return *dst
}