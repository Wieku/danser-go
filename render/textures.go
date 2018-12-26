package render

import (
	"danser/utils"
	"danser/render/texture"
)

var Atlas *texture.TextureAtlas

var Circle *texture.TextureRegion
var Spinner *texture.TextureRegion
var ApproachCircle *texture.TextureRegion
var CircleFull *texture.TextureRegion
var CircleOverlay *texture.TextureRegion
var SliderGradient *texture.TextureSingle
var SliderTick *texture.TextureRegion
var SliderBall *texture.TextureRegion
var CursorTex *texture.TextureRegion
var CursorTop *texture.TextureRegion
var CursorTrail *texture.TextureSingle
var PressKey *texture.TextureRegion

var Hit300 *texture.TextureRegion
var Hit100 *texture.TextureRegion
var Hit50 *texture.TextureRegion
var Hit0 *texture.TextureRegion

var RankX *texture.TextureRegion
var RankS *texture.TextureRegion
var RankA *texture.TextureRegion
var RankB *texture.TextureRegion
var RankC *texture.TextureRegion
var RankD *texture.TextureRegion

func LoadTextures() {
	Atlas = texture.NewTextureAtlas(8192, 4)
	Atlas.Bind(16)
	Circle, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/hitcircle.png")
	Spinner, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/spinner.png")
	ApproachCircle, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/approachcircle.png")
	CircleFull, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/hitcircle-full.png")
	CircleOverlay, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/hitcircleoverlay.png")
	SliderTick, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/sliderscorepoint.png")
	SliderBall, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/sliderball.png")
	CursorTex, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/cursor.png")
	CursorTop, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/cursor-top.png")
	SliderGradient, _ = utils.LoadTexture("assets/textures/slidergradient.png")
	CursorTrail, _ = utils.LoadTexture("assets/textures/cursortrail.png")
	PressKey, _ = utils.LoadTextureToAtlas(Atlas,"assets/textures/presskey.png")

	Hit300, _ = utils.LoadTextureToAtlas(Atlas,"assets/textures/hit-300.png")
	Hit100, _ = utils.LoadTextureToAtlas(Atlas,"assets/textures/hit-100.png")
	Hit50, _ = utils.LoadTextureToAtlas(Atlas,"assets/textures/hit-50.png")
	Hit0, _ = utils.LoadTextureToAtlas(Atlas,"assets/textures/hit-0.png")

	RankX, _ = utils.LoadTextureToAtlas(Atlas,"assets/textures/ranking-x.png")
	RankS, _ = utils.LoadTextureToAtlas(Atlas,"assets/textures/ranking-s.png")
	RankA, _ = utils.LoadTextureToAtlas(Atlas,"assets/textures/ranking-a.png")
	RankB, _ = utils.LoadTextureToAtlas(Atlas,"assets/textures/ranking-b.png")
	RankC, _ = utils.LoadTextureToAtlas(Atlas,"assets/textures/ranking-c.png")
	RankD, _ = utils.LoadTextureToAtlas(Atlas,"assets/textures/ranking-d.png")
}
