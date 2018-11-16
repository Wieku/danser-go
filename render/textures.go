package render

import (
	"github.com/wieku/danser/utils"
	"github.com/wieku/danser/render/texture"
	"github.com/wieku/danser/render/font"
)

var Atlas *texture.TextureAtlas

var Circle *texture.TextureRegion
var ApproachCircle *texture.TextureRegion
var CircleFull *texture.TextureRegion
var CircleOverlay *texture.TextureRegion
var SliderGradient *texture.TextureSingle
var SliderTick *texture.TextureRegion
var SliderBall *texture.TextureRegion
var SliderReverse *texture.TextureRegion
var SliderFollow *texture.TextureRegion
var CursorTex *texture.TextureRegion
var CursorTop *texture.TextureRegion
var SpinnerMiddle *texture.TextureRegion
var SpinnerMiddle2 *texture.TextureRegion
var SpinnerAC *texture.TextureRegion
var CursorTrail *texture.TextureSingle
var Pixel *texture.TextureSingle
var Combo *font.Font

func LoadTextures() {
	Atlas = texture.NewTextureAtlas(8192, 4)
	Atlas.Bind(16)
	Circle, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/hitcircle.png")
	ApproachCircle, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/approachcircle.png")
	CircleFull, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/hitcircle-full.png")
	CircleOverlay, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/hitcircleoverlay.png")
	SliderTick, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/sliderscorepoint.png")
	SliderBall, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/sliderball.png")
	SliderReverse, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/reversearrow.png")
	SliderFollow, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/sliderfollowcircle.png")
	CursorTex, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/cursor.png")
	CursorTop, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/cursor-top.png")
	SpinnerMiddle, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/spinner/spinner-middle.png")
	SpinnerMiddle2, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/spinner/spinner-middle2.png")
	SpinnerAC, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/spinner/spinner-approachcircle.png")
	SliderGradient, _ = utils.LoadTexture("assets/textures/slidergradient.png")
	CursorTrail, _ = utils.LoadTexture("assets/textures/cursortrail.png")
	Pixel = texture.NewTextureSingle(1, 1, 0)
	Pixel.SetData(0, 0, 1, 1, []byte{0xFF, 0xFF, 0xFF, 0xFF})
	Combo = font.LoadTextureFont("assets/textures/numbers/default-.png", "Numbers", '0', '9', Atlas)
}
