package video

import (
	"github.com/faiface/mainthread"
	"github.com/wieku/danser-go/framework/graphics/sprite"
	"github.com/wieku/danser-go/framework/graphics/texture"
	"github.com/wieku/danser-go/framework/math/vector"
)

type Video struct {
	*sprite.Sprite

	texture *texture.TextureSingle
	decoder *VideoDecoder

	lastTime float64
	Offset   int64
}

func NewVideo(path string, depth float64, position vector.Vector2d, origin vector.Vector2d) *Video {
	decoder := NewVideoDecoder(path)

	if decoder == nil {
		return nil
	}

	tex := texture.NewTextureSingleFormat(decoder.Metadata.Width, decoder.Metadata.Height, texture.RGB, 0)
	region := tex.GetRegion()

	sp := sprite.NewSpriteSingle(&region, depth, position, origin)

	decoder.StartFFmpeg(0)

	return &Video{
		Sprite:  sp,
		texture: tex,
		decoder: decoder,
	}
}

func (video *Video) Update(time int64) {
	if video.decoder == nil || video.decoder.HasFinished() {
		return
	}

	time -= video.Offset
	if time < 0 {
		return
	}

	delta := 1000.0 / video.decoder.Metadata.FPS

	if float64(time) < video.lastTime || video.lastTime+delta*10 < float64(time) {
		video.decoder.StartFFmpeg(time)
		video.lastTime = float64(time) - delta
	}

	for video.lastTime+delta < float64(time) {
		video.lastTime += delta

		frame := video.decoder.GetFrame()

		data := make([]byte, len(frame))
		copy(data, frame)
		video.decoder.Free(frame)

		mainthread.CallNonBlock(func() {
			video.texture.SetData(0, 0, video.decoder.Metadata.Width, video.decoder.Metadata.Height, data)
		})
	}
}
