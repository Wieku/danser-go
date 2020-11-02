package graphics

import (
	"github.com/wieku/danser-go/app/utils"
	"github.com/wieku/danser-go/framework/graphics/texture"
)

var Atlas *texture.TextureAtlas

var CursorTex *texture.TextureRegion
var CursorTop *texture.TextureRegion
var CursorTrail *texture.TextureSingle

var Pixel *texture.TextureSingle
var Triangle *texture.TextureRegion
var TriangleShadowed *texture.TextureRegion
var TriangleSmall *texture.TextureRegion

var Hit0 *texture.TextureRegion
var Hit50 *texture.TextureRegion
var Hit100 *texture.TextureRegion
var OvButton *texture.TextureRegion
var OvButtonE *texture.TextureRegion

func LoadTextures() {
	Atlas = texture.NewTextureAtlas(4096, 4)
	Atlas.Bind(16)

	CursorTex, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/cursor.png")
	CursorTop, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/cursor-top.png")

	Hit0, _ = utils.LoadTextureToAtlas(Atlas, "assets/default-skin/hit0-0.png")
	Hit50, _ = utils.LoadTextureToAtlas(Atlas, "assets/default-skin/hit50.png")
	Hit100, _ = utils.LoadTextureToAtlas(Atlas, "assets/default-skin/hit100.png")
	OvButton, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/ovbutton.png")
	OvButtonE, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/ovbutton-e.png")

	Triangle, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/triangle.png")
	TriangleShadowed, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/triangle-shadow.png")
	TriangleSmall, _ = utils.LoadTextureToAtlas(Atlas, "assets/textures/triangle-small.png")

	CursorTrail, _ = utils.LoadTexture("assets/textures/cursortrail.png")
	Pixel = texture.NewTextureSingle(1, 1, 0)
	Pixel.SetData(0, 0, 1, 1, []byte{0xFF, 0xFF, 0xFF, 0xFF})
}
