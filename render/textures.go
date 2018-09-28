package render

import (
	"github.com/wieku/danser/utils"
	"github.com/Wieku/danser/render/texture"
)

var Atlas *texture.TextureAtlas

var Circle *texture.TextureRegion
var ApproachCircle *texture.TextureRegion
var CircleFull *texture.TextureRegion
var CircleOverlay *texture.TextureRegion
var SliderGradient *texture.TextureSingle
var SliderTick *texture.TextureRegion
var SliderBall *texture.TextureRegion
var CursorTex *texture.TextureRegion
var CursorTop *texture.TextureRegion
var CursorTrail *texture.TextureSingle

func LoadTextures() {
	Atlas = texture.NewTextureAtlas(8192, 4)
	Atlas.Bind(16)
	Circle, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/hitcircle.png")
	ApproachCircle, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/approachcircle.png")
	CircleFull, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/hitcircle-full.png")
	CircleOverlay, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/hitcircleoverlay.png")
	SliderTick, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/sliderscorepoint.png")
	SliderBall, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/sliderball.png")
	CursorTex, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/cursor.png")
	CursorTop, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/cursor-top.png")
	SliderGradient, _ = utils.LoadTexture("assets/textures/slidergradient.png")
	CursorTrail, _ = utils.LoadTexture("assets/textures/cursortrail.png")
}
