package skin

import (
	"fmt"
	"github.com/faiface/mainthread"
	"github.com/wieku/danser-go/app/graphics/font"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/assets"
	"github.com/wieku/danser-go/framework/bass"
	"github.com/wieku/danser-go/framework/graphics/texture"
	"log"
	"path/filepath"
	"strconv"
	"strings"
)

type Source int

const (
	UNKNOWN = Source(0)
	LOCAL   = Source(1 << iota)
	SKIN
	BEATMAP
	ALL = LOCAL | SKIN | BEATMAP
)

const defaultName = "default"

var atlas *texture.TextureAtlas

var animationCache = make(map[string][]*texture.TextureRegion)

var skinCache = make(map[string]*texture.TextureRegion)
var defaultCache = make(map[string]*texture.TextureRegion)

var sourceCache = make(map[*texture.TextureRegion]Source)

//dead-locking single textures to not get swept by GC
var singleTextures = make(map[string]*texture.TextureSingle)

var fontCache = make(map[string]*font.Font)

var sampleCache = make(map[string]*bass.Sample)

var CurrentSkin = defaultName

var info *SkinInfo

func fallback() {
	CurrentSkin = defaultName

	var err error
	info, err = LoadInfo(filepath.Join("assets", "default-skin", "skin.ini"))

	if err != nil {
		log.Println("SkinManager: Default skin is corrupted! Please don't manipulate game's assets!")
		panic(err)
	}
}

func checkInit() {
	if info != nil {
		return
	}

	CurrentSkin = settings.Skin.CurrentSkin

	log.Println("SkinManager: Loading skin:", CurrentSkin)

	if CurrentSkin == defaultName {
		fallback()
	} else {
		var err error
		info, err = LoadInfo(filepath.Join(settings.General.OsuSkinsDir, CurrentSkin, "skin.ini"))
		if err != nil {
			log.Println("SkinManager:", CurrentSkin, "is corrupted, falling back to default...")
			fallback()
		}
	}

	log.Println(fmt.Sprintf("SkinManager: Skin \"%s\" loaded.", CurrentSkin))
}

func GetInfo() *SkinInfo {
	checkInit()
	return info
}

func GetFont(name string) *font.Font {
	checkInit()

	if fnt, exists := fontCache[name]; exists {
		return fnt
	}

	overlap := 0.0

	prefix := name

	switch name {
	case defaultName:
		prefix = info.HitCirclePrefix
		overlap = info.HitCircleOverlap
	case "score":
		prefix = info.ScorePrefix
		overlap = info.ScoreOverlap
	case "combo":
		prefix = info.ComboPrefix
		overlap = info.ComboOverlap
	}

	if name == "scoreentry" && GetTexture(prefix+"-0") == nil {
		return nil
	}

	chars := make(map[rune]*texture.TextureRegion)

	for i := '0'; i <= '9'; i++ {
		chars[i] = GetTexture(prefix + "-" + string(i))
	}

	chars[','] = GetTexture(prefix + "-comma")
	chars['.'] = GetTexture(prefix + "-dot")
	chars['%'] = GetTexture(prefix + "-percent")
	chars['x'] = GetTexture(prefix + "-x")

	fnt := font.LoadTextureFontMap2(chars, overlap)

	fontCache[name] = fnt

	return fnt
}

func GetTexture(name string) *texture.TextureRegion {
	return GetTextureSource(name, ALL)
}

func GetTextureSource(name string, source Source) *texture.TextureRegion {
	checkInit()

	source = source & (^BEATMAP)

	if CurrentSkin == defaultName {
		source = source & (^SKIN)
	}

	if source&SKIN > 0 {
		if rg, exists := skinCache[name]; exists {
			if rg != nil {
				return rg
			}
		} else {
			rg := loadTexture(filepath.Join(settings.General.OsuSkinsDir, CurrentSkin, name+".png"))
			skinCache[name] = rg

			if rg != nil {
				sourceCache[rg] = SKIN
				return rg
			}
		}
	}

	if source&LOCAL > 0 {
		if rg, exists := defaultCache[name]; exists {
			return rg
		}

		rg := loadTexture(filepath.Join("assets", "default-skin", name+".png"))
		defaultCache[name] = rg

		if rg != nil {
			sourceCache[rg] = LOCAL
		}

		return rg
	}

	return nil
}

func GetFrames(name string, useDash bool) []*texture.TextureRegion {
	if rg, exists := animationCache[name]; exists {
		return rg
	}

	dash := ""
	if useDash {
		dash = "-"
	}

	textures := make([]*texture.TextureRegion, 0)

	spTexture := GetTexture(name)
	frame := GetTexture(name + dash + "0")

	if frame != nil && frame == GetMostSpecific(frame, spTexture) {
		source := sourceCache[frame]

		for i := 1; frame != nil; i++ {
			textures = append(textures, frame)
			frame = GetTextureSource(name+dash+strconv.Itoa(i), source)
		}
	} else if spTexture != nil {
		textures = append(textures, spTexture)
	}

	animationCache[name] = textures

	return textures
}

func GetMostSpecific(rg1, rg2 *texture.TextureRegion) *texture.TextureRegion {
	if rg1 == nil {
		return rg2
	}

	if rg2 == nil {
		return rg1
	}

	rg1S := sourceCache[rg1]
	rg2S := sourceCache[rg2]

	if rg1S == BEATMAP ||
		rg1S == SKIN && rg2S != BEATMAP ||
		rg1S == LOCAL && rg2S != BEATMAP && rg2S != SKIN {
		return rg1
	}

	return rg2
}

func GetSource(name string) Source {
	tx := GetTexture(name)
	if tx == nil {
		return UNKNOWN
	}

	return sourceCache[tx]
}

func GetSourceFromTexture(rg *texture.TextureRegion) Source {
	if rg == nil {
		return UNKNOWN
	}

	return sourceCache[rg]
}

func checkAtlas() {
	if atlas == nil {
		atlas = texture.NewTextureAtlas(2048, 0)
		atlas.Bind(27)
	}
}

func getPixmap(name string) (*texture.Pixmap, error) {
	if strings.HasPrefix(name, "assets") {
		return assets.GetPixmap(name)
	}

	return texture.NewPixmapFileString(name)
}

func loadTexture(name string) *texture.TextureRegion {
	ext := filepath.Ext(name)

	x2Name := strings.TrimSuffix(name, ext) + "@2x" + ext

	var region *texture.TextureRegion

	image, err := getPixmap(x2Name)
	if err != nil {
		image, err = getPixmap(name)
		if err == nil {
			region = &texture.TextureRegion{}
			region.Width = int32(image.Width)
			region.Height = int32(image.Height)
		}
	} else {
		region = &texture.TextureRegion{}
		region.Width = int32(image.Width / 2)
		region.Height = int32(image.Height / 2)
	}

	if region != nil {
		// Upload this texture in GL thread
		mainthread.CallNonBlock(func() {
			checkAtlas()

			var rg *texture.TextureRegion

			if image.Width <= 1000 && image.Height <= 1000 {
				rg = atlas.AddTexture(name, image.Width, image.Height, image.Data)
			}

			// If texture is too big load it separately
			if rg == nil {
				tx := texture.NewTextureSingle(image.Width, image.Height, 0)
				tx.Bind(0)
				tx.SetData(0, 0, image.Width, image.Height, image.Data)
				reg := tx.GetRegion()

				singleTextures[name] = tx

				rg = &reg

				log.Println("SkinManager: Texture uploaded as single texture:", name)
			}

			image.Dispose()

			region.Texture = rg.Texture
			region.Layer = rg.Layer
			region.U1 = rg.U1
			region.U2 = rg.U2
			region.V1 = rg.V1
			region.V2 = rg.V2
		})
	}

	return region
}

func GetSample(name string) *bass.Sample {
	checkInit()
	if sample, exists := sampleCache[name]; exists {
		return sample
	}

	var sample *bass.Sample

	if CurrentSkin != defaultName {
		sample = tryLoad(filepath.Join(settings.General.OsuSkinsDir, CurrentSkin, name))
	}

	if sample == nil {
		sample = tryLoad(filepath.Join("assets", "default-skin", name))
	}

	sampleCache[name] = sample

	return sample
}

func getSample(name string) *bass.Sample {
	if strings.HasPrefix(name, "assets") {
		data, err := assets.GetBytes(name)
		if err != nil {
			return nil
		}

		return bass.NewSampleData(data)
	}

	return bass.NewSample(name)
}

func tryLoad(basePath string) *bass.Sample {
	if sam := getSample(basePath + ".wav"); sam != nil {
		return sam
	}

	if sam := getSample(basePath + ".ogg"); sam != nil {
		return sam
	}

	if sam := getSample(basePath + ".mp3"); sam != nil {
		return sam
	}

	return nil
}
