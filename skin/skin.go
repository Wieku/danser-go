package skin

import "github.com/wieku/danser-go/render/texture"

type Source int

const (
	LOCAL = Source(iota)
	SKIN
	BEATMAP
)

var atlases = make(map[Source]*texture.TextureAtlas)
var textures = make(map[Source]map[string]*texture.TextureRegion)

var CurrentSkin string

