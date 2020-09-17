package storyboard

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/app/animation"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/app/render/batches"
	"github.com/wieku/danser-go/framework/graphics/texture"
	"math"
	"sort"
	"strings"
	"unicode"
)

const (
	storyboardArea = 640.0 * 480.0
	maxLoad        = 1.3328125 //480*480*(16/9)/(640*480)
)

type color struct {
	R, G, B, A float64
}

type Object interface {
	Update(time int64)
	Draw(time int64, batch *batches.SpriteBatch)
	GetLoad() float64
	GetStartTime() int64
	GetEndTime() int64
	GetZIndex() int64

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
	texture                    []*texture.TextureRegion
	frameDelay                 float64
	loopForever                bool
	currentFrame               int
	transform                  *Transformations
	transforms                 []*animation.Transformation
	startTime, endTime, zIndex int64
	position                   bmath.Vector2d
	origin                     bmath.Vector2d
	scale                      bmath.Vector2d
	flip                       bmath.Vector2d
	rotation                   float64
	color                      color
	dirty                      bool
	additive                   bool
	firstupdate                bool
	lastFinished               int
	subCommands                []string
}

func cutWhites(text string) (string, int) {
	for i, c := range text {
		if unicode.IsLetter(c) || unicode.IsNumber(c) {
			return text[i:], i
		}
	}

	return text, 0
}

func NewSprite(texture []*texture.TextureRegion, frameDelay float64, loopForever bool, zIndex int64, position bmath.Vector2d, origin bmath.Vector2d, subCommands []string) *Sprite {
	sprite := &Sprite{texture: texture, frameDelay: frameDelay, loopForever: loopForever, zIndex: zIndex, position: position, origin: origin, scale: bmath.NewVec2d(1, 1), flip: bmath.NewVec2d(1, 1), color: color{1, 1, 1, 1}}
	sprite.lastFinished = -1
	sprite.transform = NewTransformations(sprite)
	sprite.subCommands = subCommands
	sprite.startTime = math.MaxInt64
	sprite.endTime = math.MinInt64

	var currentLoop *LoopProcessor = nil
	loopDepth := -1

	for _, subCommand := range subCommands {
		command := strings.Split(subCommand, ",")
		var removed int
		command[0], removed = cutWhites(command[0])

		if command[0] == "T" {
			continue
		}

		if removed == 1 {
			if currentLoop != nil {
				sprite.transforms = append(sprite.transforms, currentLoop.Unwind()...)

				currentLoop = nil
				loopDepth = -1
			}
			if command[0] != "L" {
				sprite.transforms = append(sprite.transforms, NewCommand(command).GenerateTransformations()...)
			}
		}

		if command[0] == "L" {

			currentLoop = NewLoopProcessor(command)

			loopDepth = removed + 1

		} else if removed == loopDepth {
			currentLoop.Add(NewCommand(command))
		}
	}

	if currentLoop != nil {
		sprite.transforms = append(sprite.transforms, currentLoop.Unwind()...)

		currentLoop = nil
		loopDepth = -1
	}

	sort.SliceStable(sprite.transforms, func(i, j int) bool {
		return sprite.transforms[i].GetStartTime() < sprite.transforms[j].GetStartTime()
	})

	ex := make(map[animation.TransformationType]int)

	for _, t := range sprite.transforms {
		if int64(t.GetStartTime()) < sprite.startTime {
			sprite.startTime = int64(t.GetStartTime())
		}

		if int64(t.GetEndTime()) > sprite.endTime {
			sprite.endTime = int64(t.GetEndTime())
		}

		if t.GetType() == animation.Additive || t.GetType() == animation.HorizontalFlip || t.GetType() == animation.VerticalFlip {
			continue
		}

		if _, applied := ex[t.GetType()]; !applied {
			sprite.updateTransform(t, t.GetStartTime()-1)
			ex[t.GetType()] = 1
		}
	}

	return sprite
}

func (sprite *Sprite) Update(time int64) {
	sprite.currentFrame = 0

	if len(sprite.texture) > 1 {
		frame := int(math.Floor(float64(time-sprite.startTime) / sprite.frameDelay))
		if frame < 0 {
			frame = 0
		}

		if !sprite.loopForever {
			if frame >= len(sprite.texture) {
				frame = len(sprite.texture) - 1
			}
			sprite.currentFrame = frame
		} else {
			sprite.currentFrame = frame % len(sprite.texture)
		}
	}

	for i := sprite.lastFinished + 1; i < len(sprite.transforms); i++ {
		transform := sprite.transforms[i]
		if float64(time) < transform.GetStartTime() {
			break
		}

		sprite.updateTransform(transform, float64(time))

		if float64(time) >= transform.GetEndTime() {
			copy(sprite.transforms[i:], sprite.transforms[i+1:])
			sprite.transforms = sprite.transforms[:len(sprite.transforms)-1]
			i--
		}
	}

	sprite.firstupdate = true
}

func (sprite *Sprite) updateTransform(transform *animation.Transformation, time float64) {
	switch transform.GetType() {
	case animation.Fade, animation.Scale, animation.Rotate, animation.MoveX, animation.MoveY:
		value := transform.GetSingle(time)
		switch transform.GetType() {
		case animation.Fade:
			sprite.color.A = value
		case animation.Scale:
			sprite.scale.X = value
			sprite.scale.Y = value
		case animation.Rotate:
			sprite.rotation = value
		case animation.MoveX:
			sprite.position.X = value
		case animation.MoveY:
			sprite.position.Y = value
		}
	case animation.Move, animation.ScaleVector:
		x, y := transform.GetDouble(time)
		switch transform.GetType() {
		case animation.Move:
			sprite.position.X = x
			sprite.position.Y = y
		case animation.ScaleVector:
			sprite.scale.X = x
			sprite.scale.Y = y
		}
	case animation.Additive, animation.HorizontalFlip, animation.VerticalFlip:
		value := transform.GetBoolean(time)
		va1 := 1.0
		if value {
			va1 = -1
		}
		switch transform.GetType() {
		case animation.Additive:
			sprite.additive = value
		case animation.HorizontalFlip:
			sprite.flip.X = va1
		case animation.VerticalFlip:
			sprite.flip.Y = va1
		}

	case animation.Color3, animation.Color4:
		color := transform.GetColor(time)
		sprite.color.R = color.R
		sprite.color.G = color.G
		sprite.color.B = color.B
		if transform.GetType() == animation.Color4 {
			sprite.color.A = color.A
		}
	}
}

func (sprite *Sprite) Draw(time int64, batch *batches.SpriteBatch) {
	if !sprite.firstupdate || sprite.color.A < 0.01 {
		return
	}

	alpha := sprite.color.A
	if alpha > 1.001 {
		alpha -= math.Ceil(sprite.color.A) - 1
	}

	batch.DrawStObject(sprite.position, sprite.origin, sprite.scale.Abs(), sprite.flip, sprite.rotation, mgl32.Vec4{float32(sprite.color.R), float32(sprite.color.G), float32(sprite.color.B), float32(alpha)}, sprite.additive, *sprite.texture[sprite.currentFrame], true)
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
	j := 1.0
	if on {
		j = -1
	}
	sprite.flip.X = j
	sprite.dirty = true
}

func (sprite *Sprite) SetVFlip(on bool) {
	j := 1.0
	if on {
		j = -1
	}
	sprite.flip.Y = j
	sprite.dirty = true
}

func (sprite *Sprite) SetAdditive(on bool) {
	sprite.additive = on
	sprite.dirty = true
}

func (sprite *Sprite) GetStartTime() int64 {
	return sprite.startTime
}

func (sprite *Sprite) GetEndTime() int64 {
	return sprite.endTime
}

func (sprite *Sprite) GetZIndex() int64 {
	return sprite.zIndex
}

func (sprite *Sprite) GetLoad() float64 {
	if sprite.color.A >= 0.01 {
		return math.Min((float64(sprite.texture[0].Width)*sprite.scale.X*float64(sprite.texture[0].Height)*sprite.scale.Y)/storyboardArea, maxLoad)
	}
	return 0
}
