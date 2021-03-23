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
	//Offset   float64
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

func (video *Video) Update(time float64) {
	if video.decoder == nil || video.decoder.HasFinished() {
		video.SetEndTime(time)
		return
	}

	time -= video.GetStartTime()
	if time < 0 {
		return
	}

	delta := 1000.0 / video.decoder.Metadata.FPS

	if time < video.lastTime || video.lastTime+delta*10 < time {
		video.decoder.StartFFmpeg(int64(time))
		video.lastTime = time - delta
	}

	for video.lastTime+delta < time {
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
