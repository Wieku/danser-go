package render

import (
	"github.com/wieku/danser-go/utils"
	"github.com/wieku/danser-go/render/texture"
	"github.com/wieku/danser-go/render/font"
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
var CursorTrail *texture.TextureSingle

var SpinnerMiddle *texture.TextureRegion
var SpinnerMiddle2 *texture.TextureRegion
var SpinnerAC *texture.TextureRegion

var Pixel *texture.TextureSingle
var Triangle *texture.TextureRegion
var Combo *font.Font

var RankingD *texture.TextureRegion
var RankingC *texture.TextureRegion
var RankingB *texture.TextureRegion
var RankingA *texture.TextureRegion
var RankingS *texture.TextureRegion
var RankingSH *texture.TextureRegion
var RankingSS *texture.TextureRegion
var RankingSSH *texture.TextureRegion

var GradeTexture map[int64]*texture.TextureRegion

var Hit0 *texture.TextureRegion
var Hit50 *texture.TextureRegion
var Hit100 *texture.TextureRegion
var OvButton *texture.TextureRegion
var OvButtonE *texture.TextureRegion

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

	SpinnerMiddle, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/spinner-middle.png")
	SpinnerMiddle2, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/spinner-middle2.png")
	SpinnerAC, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/spinner-approachcircle.png")

	RankingD, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/ranking-d-small.png")
	RankingC, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/ranking-c-small.png")
	RankingB, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/ranking-b-small.png")
	RankingA, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/ranking-a-small.png")
	RankingS, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/ranking-s-small.png")
	RankingSH, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/ranking-sh-small.png")
	RankingSS, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/ranking-x-small.png")
	RankingSSH, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/ranking-xh-small.png")

	GradeTexture = map[int64]*texture.TextureRegion{0: RankingD, 1: RankingC, 2: RankingB, 3: RankingA, 4: RankingS, 5: RankingSH, 6: RankingSS, 7: RankingSSH}

	Hit0, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/hit0-0.png")
	Hit50, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/hit50.png")
	Hit100, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/hit100.png")
	OvButton, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/ovbutton.png")
	OvButtonE, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/ovbutton-e.png")

	Triangle, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/triangle.png")

	SliderGradient, _ = utils.LoadTexture("assets/textures/slidergradient.png")
	CursorTrail, _ = utils.LoadTexture("assets/textures/cursortrail.png")
	Pixel = texture.NewTextureSingle(1, 1, 0)
	Pixel.SetData(0, 0, 1, 1, []byte{0xFF, 0xFF, 0xFF, 0xFF})
	Combo = font.LoadTextureFont("assets/textures/default-.png", "Numbers", '0', '9', Atlas)
}
