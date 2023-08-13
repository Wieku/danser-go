package sprite

import (
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/texture"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
)

type Animation struct {
	*Sprite

	textures     []*texture.TextureRegion
	frameDelay   float64
	loopForever  bool
	currentFrame int
}

func NewAnimation(textures []*texture.TextureRegion, frameDelay float64, loopForever bool, depth float64, position vector.Vector2d, origin vector.Vector2d) *Animation {
	animation := &Animation{
		Sprite:      NewSpriteSingle(nil, depth, position, origin),
		textures:    textures,
		frameDelay:  frameDelay,
		loopForever: loopForever,
	}

	if animation.textures != nil && len(animation.textures) > 0 {
		animation.Texture = animation.textures[0]
	}

	return animation
}

func (animation *Animation) Update(time float64) {
	if animation.textures != nil && len(animation.textures) > 1 && time >= animation.startTime {
		frame := int(math.Floor((time - animation.startTime) / animation.frameDelay))
		if !animation.loopForever {
			frame = min(frame, len(animation.textures)-1)
		} else {
			frame = frame % len(animation.textures)
		}

		animation.currentFrame = frame
	} else {
		animation.currentFrame = 0
	}

	animation.Sprite.Update(time)
}

func (animation *Animation) Draw(time float64, batch *batch.QuadBatch) {
	if animation.textures == nil || len(animation.textures) == 0 || animation.textures[animation.currentFrame] == nil {
		return
	}

	animation.Texture = animation.textures[animation.currentFrame]

	animation.Sprite.Draw(time, batch)
}
