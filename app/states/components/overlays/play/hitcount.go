package play

import (
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/app/skin"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/sprite"
	"github.com/wieku/danser-go/framework/math/vector"
)

type HitCounter struct {
	*sprite.Sprite

	hit   *sprite.Sprite
	value string

	yOffset float64
}

func NewHitCounter(hitText string, value string, position vector.Vector2d) *HitCounter {
	aSprite := &HitCounter{
		Sprite: sprite.NewSpriteSingle(nil, 3, position, vector.NewVec2d(0,0)),
		hit: sprite.NewSpriteSingle(skin.GetTexture(hitText), 0, vector.NewVec2d(0,0), bmath.Origin.Centre),
		value: value,
	}

	aSprite.hit.SetScale(0.5)

	if skin.GetInfo().Version >= 2 {
		aSprite.yOffset = -16/0.625
	} else {
		aSprite.yOffset = -25/0.625
	}

	return aSprite
}

func (sprite *HitCounter) Update(time float64) {
	sprite.hit.Update(time)
	sprite.hit.SetPosition(sprite.Sprite.GetPosition())
}


func (sprite *HitCounter) Draw(time float64, batch *batch.QuadBatch) {
	sprite.hit.Draw(time, batch)

	pos := sprite.Sprite.GetPosition()

	fnt := skin.GetFont("score")

	fnt.DrawOriginV(batch, pos.AddS(40/0.625, sprite.yOffset), bmath.Origin.TopLeft, fnt.GetSize()*1.12, false, sprite.value)
	batch.ResetTransform()
}
