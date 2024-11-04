package font

import (
	"github.com/wieku/danser-go/app/utils"
	"github.com/wieku/danser-go/framework/graphics/texture"
	color2 "github.com/wieku/danser-go/framework/math/color"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
	"image"
	"io"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"unicode"
)

func LoadFont(reader io.Reader) *Font {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		panic("Error reading font: " + err.Error())
	}

	ttf, err := opentype.Parse(data)
	if err != nil {
		panic("Error reading font: " + err.Error())
	}

	fnt := new(Font)
	fnt.flip = true
	fnt.initialSize = 128.0
	fnt.glyphs = make(map[rune]*glyphData)
	fnt.kernTable = make(map[rune]map[rune]float64)

	fnt.atlas = texture.NewTextureAtlasCC(1024, 5, color2.NewLA(1, 0))
	fnt.atlas.SetManualMipmapping(true)

	buf := make([]byte, 25*4)
	for i := range buf {
		buf[i] = 0xff
	}

	fnt.pixel = fnt.atlas.AddTexture("pixel", 5, 5, buf)
	fnt.pixel.Width = 1
	fnt.pixel.Height = 1

	fc, err := opentype.NewFace(ttf, &opentype.FaceOptions{Size: fnt.initialSize, DPI: 72, Hinting: font.HintingFull})
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
				w = 2
				h = 2
			}

			if b.Min.X&((1<<6)-1) != 0 {
				w++
			}

			if b.Min.Y&((1<<6)-1) != 0 {
				h++
			}

			pixmap := texture.NewPixMapW(w, h)

			d := font.Drawer{
				Dst:  pixmap.NRGBA(),
				Src:  image.White,
				Face: fc,
			}

			x, y := fixed.I((-b.Min.X).Ceil()), fixed.I((-b.Min.Y).Ceil())
			d.Dot = fixed.Point26_6{X: x, Y: y}
			d.DrawString(string(i))

			region := fnt.atlas.AddTexture(string(i), pixmap.Width, pixmap.Height, pixmap.Data)

			region.V1, region.V2 = region.V2, region.V1

			pixmap.Dispose()

			//set w,h and adv, bearing V and bearing H in char
			advance := float64(gAdv) / 64
			offsetX := float64(b.Min.X) / 64
			ascent := float64(-b.Min.Y) / 64

			//Calculate real ascent from the tallest A-Z glyph, because glyphs like Ž may make it bigger
			if i >= 'A' && i <= 'Z' {
				fnt.ascent = max(fnt.ascent, float64(-b.Min.Y)/64)
			}

			fnt.glyphs[i] = &glyphData{region, advance, offsetX, ascent}
		}
	}

	fnt.atlas.GenerateMipmaps()

	for i, g := range fnt.glyphs {
		//Calculate real offset based on ascent calculated above
		g.offsetY = fnt.ascent - g.offsetY

		fnt.kernTable[i] = make(map[rune]float64)

		for j := range fnt.glyphs {
			if krn := fc.Kern(i, j); krn != 0 {
				fnt.kernTable[i][j] = float64(krn) / 64
			}
		}
	}

	fnt.biggest = fnt.glyphs['5'].advance

	name, _ := ttf.Name(buff, sfnt.NameIDFull)

	fonts[name] = fnt

	log.Println(name, "loaded!")

	return fnt
}

func LoadTextureFont(path, name string, minR, maxR rune, atlas *texture.TextureAtlas) *Font {
	font := new(Font)

	font.glyphs = make(map[rune]*glyphData)
	font.biggest = 40
	font.atlas = atlas

	extension := filepath.Ext(path)
	baseFile := strings.TrimSuffix(path, extension)

	for i := minR; i <= maxR; i++ {
		region, _ := utils.LoadTextureToAtlas(font.atlas, baseFile+string(i)+extension)

		font.initialSize = max(font.initialSize, float64(region.Height))

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

		font.initialSize = max(font.initialSize, float64(region.Height))

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

		font.initialSize = max(font.initialSize, float64(r.Height))

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

	font.ascent = font.initialSize
}
