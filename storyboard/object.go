package storyboard

import (
	"github.com/Wieku/danser/bmath"
	"github.com/go-gl/mathgl/mgl32"
	"sort"
	"github.com/Wieku/danser/render"
	"github.com/Wieku/glhf"
	"github.com/go-gl/gl/v3.3-core/gl"
)

type color struct {
	R, G, B, A float64
}

type Transformations struct {
	object             Object
	queue              []*Command
	processed          []*Command
	startTime, endTime int64
}

func (trans *Transformations) Add(command *Command) {
	trans.queue = append(trans.queue, command)
	sort.Slice(trans.queue, func(i, j int) bool {
		return trans.queue[i].start < trans.queue[j].start
	})

	if command.start < trans.startTime {
		trans.startTime = command.start
	}

	if command.end > trans.endTime {
		trans.endTime = command.end
	}
}

func (trans *Transformations) Update(time int64) {
	for i, c := range trans.queue {
		if c.start <= time {
			trans.processed = append(trans.processed, c)
			trans.queue = append(trans.queue[:i], trans.queue[i+1:]...)
		}
	}

	for i, c := range trans.processed {
		c.Update(time)
		c.Apply(trans.object)

		if time > c.end {
			trans.processed = append(trans.processed[:i], trans.processed[i+1:]...)
		}
	}

}

type Object interface {
	Update(time int64)
	Draw(time int64, batch *render.SpriteBatch)
	GetTransform() *Transformations

	GetPosition() bmath.Vector2d
	SetPosition(vec bmath.Vector2d)

	GetScale() bmath.Vector2d
	SetScale(vec bmath.Vector2d)

	GetRotation() float64
	SetRotation(rad float64)

	GetColor() mgl32.Vec3
	SetColor(color mgl32.Vec3)

	GetAlpha() float64
	SetAlpha(alpha float64)

	SetHFlip(on bool)
	SetVFlip(on bool)

	SetAdditive(on bool)
}

type Sprite struct {
	texture                *glhf.Texture
	transform              *Transformations
	position               bmath.Vector2d
	origin                 bmath.Vector2d
	scale                  bmath.Vector2d
	rotation               float64
	color                  color
	dirty                  bool
	hflip, vflip, additive bool
}

func (sprite *Sprite) Update(time int64) {
	sprite.transform.Update(time)
}

func (sprite *Sprite) Draw(time int64, batch *render.SpriteBatch) {
	if sprite.additive {
		gl.BlendFunc(gl.SRC_ALPHA, gl.ONE)
	}

	batch.DrawStObject(sprite.position, sprite.origin, sprite.scale, sprite.rotation, mgl32.Vec4{float32(sprite.color.R), float32(sprite.color.G), float32(sprite.color.B), float32(sprite.color.A)}, sprite.texture)

	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
}

func (sprite *Sprite) GetPosition() bmath.Vector2d {
	return sprite.position
}

func (sprite *Sprite) SetPosition(vec bmath.Vector2d) {
	sprite.position = vec
	sprite.dirty = true
}

func (sprite *Sprite) GetScale() bmath.Vector2d {
	return sprite.scale
}

func (sprite *Sprite) SetScale(vec bmath.Vector2d) {
	sprite.scale = vec
	sprite.dirty = true
}

func (sprite *Sprite) GetRotation() float64 {
	return sprite.rotation
}

func (sprite *Sprite) SetRotation(rad float64) {
	sprite.rotation = rad
	sprite.dirty = true
}

func (sprite *Sprite) GetColor() mgl32.Vec3 {
	return mgl32.Vec3{float32(sprite.color.R), float32(sprite.color.G), float32(sprite.color.B)}
}

func (sprite *Sprite) SetColor(color mgl32.Vec3) {
	sprite.color.R, sprite.color.G, sprite.color.B = float64(color[0]), float64(color[1]), float64(color[2])
	sprite.dirty = true
}

func (sprite *Sprite) GetAlpha() float64 {
	return sprite.color.A
}

func (sprite *Sprite) SetAlpha(alpha float64) {
	sprite.color.A = alpha
	sprite.dirty = true
}

func (sprite *Sprite) SetHFlip(on bool) {
	sprite.hflip = on
	sprite.dirty = true
}

func (sprite *Sprite) SetVFlip(on bool) {
	sprite.vflip = on
	sprite.dirty = true
}

func (sprite *Sprite) SetAdditive(on bool) {
	sprite.additive = on
	sprite.dirty = true
}
