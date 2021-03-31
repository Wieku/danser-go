package audio

import (
	"github.com/wieku/danser-go/framework/bass"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/sprite"
	"github.com/wieku/danser-go/framework/math/vector"
)

type AudioSprite struct {
	*sprite.Sprite

	sample *bass.Sample

	played bool
	playAt float64
}

func NewAudioSprite(sample *bass.Sample, playAt float64) *AudioSprite {
	aSprite := &AudioSprite{
		Sprite: sprite.NewSpriteSingle(nil, 0, vector.NewVec2d(0,0), vector.NewVec2d(0,0)),
		playAt: playAt,
		sample: sample,
	}
	aSprite.Sprite.ShowForever(true)

	return aSprite
}

func (sprite *AudioSprite) Update(time float64) {
	if time >= sprite.playAt && !sprite.played {
		if sprite.sample != nil {
			sprite.sample.Play()
		}

		sprite.played = true
		sprite.Sprite.ShowForever(false)
	}
}


func (sprite *AudioSprite) Draw(_ float64, _ *batch.QuadBatch) {

}