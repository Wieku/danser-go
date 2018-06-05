package render

import (
	"github.com/faiface/glhf"
	"github.com/wieku/danser/utils"
	"log"
)

var Circle *glhf.Texture
var CircleFull *glhf.Texture
var CircleOverlay *glhf.Texture
var SliderGradient *glhf.Texture
var SliderBall *glhf.Texture
var CursorTex *glhf.Texture
var CursorTop *glhf.Texture
var CursorTrail *glhf.Texture


func LoadTextures() {
	var err error
	log.Println("tyxxturs")
	Circle, err = utils.LoadTexture("assets/textures/hitcircle.png")
	log.Println(err)
	CircleFull, _ = utils.LoadTexture("assets/textures/hitcircle-full.png")
	CircleOverlay, _ = utils.LoadTexture("assets/textures/hitcircleoverlay.png")
	SliderGradient, _ = utils.LoadTexture("assets/textures/slidergradient.png")
	SliderBall, _ = utils.LoadTexture("assets/textures/sliderball.png")
	CursorTex, _ = utils.LoadTexture("assets/textures/cursor.png")
	CursorTop, _ = utils.LoadTexture("assets/textures/cursor-top.png")
	CursorTrail, _ = utils.LoadTexture("assets/textures/cursortrail.png")
}
