package storyboard

import (
	"github.com/wieku/danser/bmath"
	"github.com/go-gl/mathgl/mgl32"
	"sort"
	"github.com/wieku/danser/render"
	"github.com/wieku/glhf"
	"github.com/go-gl/gl/v3.3-core/gl"
	"unicode"
	"strings"
	"math"
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

func NewTransformations(obj Object) *Transformations {
	return &Transformations{object: obj, startTime: math.MaxInt64, endTime: math.MinInt64}
}

func (trans *Transformations) Add(command *Command) {

	if command.command != "P" {
		exists := false

		for _, e := range trans.queue {
			if e.command == command.command && e.start < command.start {
				exists = true
				break
			}
		}

		if !exists {
			command.Update(command.start)
			command.Apply(trans.object)
		}
	}

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
	for i := 0; i < len(trans.queue); i++ {
		c := trans.queue[i]
		if c.start <= time {
			trans.processed = append(trans.processed, c)
			trans.queue = append(trans.queue[:i], trans.queue[i+1:]...)
			i--
		}
	}

	for i := 0; i < len(trans.processed); i++ {
		c := trans.processed[i]
		c.Update(time)
		c.Apply(trans.object)

		if time > c.end {
			trans.processed = append(trans.processed[:i], trans.processed[i+1:]...)
			i--
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
	firstupdate            bool
}

func cutWhites(text string) (string, int) {
	for i, c := range text {
		if unicode.IsLetter(c) || unicode.IsNumber(c) {
			return text[i:], i
		}
	}

	return text, 0
}

func NewSprite(texture *glhf.Texture, position bmath.Vector2d, origin bmath.Vector2d, subCommands []string) *Sprite {
	sprite := &Sprite{texture: texture, position: position, origin: origin, scale: bmath.NewVec2d(1, 1), color: color{1, 1, 1, 1}}
	sprite.transform = NewTransformations(sprite)

	for _, subCommand := range subCommands {
		command := strings.Split(subCommand, ",")
		var removed int
		command[0], removed = cutWhites(command[0])
		if removed == 1 && command[0] != "L" && command[0] != "T" {
			sprite.transform.Add(NewCommand(command))
		}
	}

	return sprite
}

func (sprite *Sprite) Update(time int64) {
	sprite.transform.Update(time)
	sprite.firstupdate = true
}

func (sprite *Sprite) Draw(time int64, batch *render.SpriteBatch) {
	if !sprite.firstupdate {
		return
	}
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

func (sprite *Sprite) GetTransform() *Transformations {
	return sprite.transform
}
