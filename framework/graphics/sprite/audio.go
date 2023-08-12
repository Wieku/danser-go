package sprite

import (
	"github.com/wieku/danser-go/framework/bass"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/math/vector"
)

type AudioSprite struct {
	*Sprite

	sample *bass.Sample

	sampleChannel *bass.SampleChannel

	volume float64
}

func NewAudioSprite(sample *bass.Sample, playAt, volume float64) *AudioSprite {
	aSprite := &AudioSprite{
		Sprite: NewSpriteSingle(nil, 0, vector.NewVec2d(0, 0), vector.NewVec2d(0, 0)),
		sample: sample,
		volume: volume,
	}

	aSprite.SetStartTime(playAt)
	aSprite.SetEndTime(playAt + 400)

	if sample != nil {
		length := sample.GetLength() * 1000

		aSprite.SetEndTime(playAt + max(100, length)) //some leeway for short samples
	}

	return aSprite
}

func (sprite *AudioSprite) Update(time float64) {
	if sprite.sample == nil {
		return
	}

	if time >= sprite.GetStartTime() && time <= sprite.GetEndTime() && sprite.sampleChannel == nil {
		sprite.sampleChannel = sprite.sample.PlayRV(sprite.volume)
	}

	if time >= sprite.GetEndTime() && sprite.sampleChannel != nil {
		bass.StopSample(sprite.sampleChannel)

		sprite.sampleChannel = nil
	}
}

func (sprite *AudioSprite) Draw(_ float64, _ *batch.QuadBatch) {}
