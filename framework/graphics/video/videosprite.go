package video

import (
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/sprite"
	"github.com/wieku/danser-go/framework/graphics/texture"
	"github.com/wieku/danser-go/framework/math/vector"
	"sync"
)

type Video struct {
	*sprite.Sprite

	texture *texture.TextureSingle
	decoder *VideoDecoder

	lastTime float64

	mutex *sync.Mutex
	data  []byte
	dirty bool
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

	video := &Video{
		Sprite:  sp,
		texture: tex,
		decoder: decoder,
		mutex:   &sync.Mutex{},
		data:    make([]byte, decoder.Metadata.Width*decoder.Metadata.Height*3),
	}

	video.SetStartTime(0)

	return video
}

func (video *Video) SetStartTime(startTime float64) {
	video.Sprite.SetStartTime(startTime)
	video.Sprite.SetEndTime(startTime + video.decoder.Metadata.Duration*1000)
}

func (video *Video) Update(time float64) {
	video.Sprite.Update(time)

	if video.decoder == nil || video.decoder.HasFinished() {
		return
	}

	time -= video.GetStartTime()
	if time < 0 {
		return
	}

	delta := 1000.0 / video.decoder.Metadata.FPS

	if time < video.lastTime || video.lastTime+1000 < time {
		video.decoder.StartFFmpeg(int64(time))
		video.lastTime = time - delta
	}

	if video.lastTime+delta < time {
		video.lastTime += delta

		frame := video.decoder.GetFrame()

		video.mutex.Lock()
		copy(video.data, frame)
		video.dirty = true
		video.mutex.Unlock()

		video.decoder.Free(frame)
	}
}

func (video *Video) Draw(time float64, batch *batch.QuadBatch) {
	video.mutex.Lock()
	if video.dirty {
		gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)

		video.texture.SetData(0, 0, video.decoder.Metadata.Width, video.decoder.Metadata.Height, video.data)

		video.dirty = false
	}
	video.mutex.Unlock()

	video.Sprite.Draw(time, batch)
}
