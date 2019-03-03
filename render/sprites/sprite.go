package sprites

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser/animation"
	"github.com/wieku/danser/bmath"
	"github.com/wieku/danser/render/batches"
	"github.com/wieku/danser/render/texture"
	"math"
	"sort"
)

const (
	storyboardArea = 640.0 * 480.0
	maxLoad        = 1.3328125 //480*480*(16/9)/(640*480)
)

type Sprite struct {
	texture      []*texture.TextureRegion
	frameDelay   float64
	loopForever  bool
	currentFrame int
	transforms   []*animation.Transformation

	startTime, endTime, depth float64

	position         bmath.Vector2d
	positionRelative bmath.Vector2d
	origin           bmath.Vector2d
	scale            bmath.Vector2d
	flip             bmath.Vector2d
	rotation         float64
	color            bmath.Color
	dirty            bool
	additive         bool
	showForever      bool
}

func NewSpriteSingle(tex *texture.TextureRegion, depth float64, position bmath.Vector2d, origin bmath.Vector2d) *Sprite {
	textures := []*texture.TextureRegion{tex}
	sprite := &Sprite{texture: textures, frameDelay: 0.0, loopForever: true, depth: depth, position: position, origin: origin, scale: bmath.NewVec2d(1, 1), flip: bmath.NewVec2d(1, 1), color: bmath.Color{1, 1, 1, 1}, showForever:true}
	sprite.transforms = make([]*animation.Transformation, 0)
	return sprite
}

/*
func NewSpriteSingle(texture []*texture.TextureRegion, frameDelay float64, loopForever bool, depth float64, position bmath.Vector2d, origin bmath.Vector2d) *Sprite {
	sprite := &Sprite{texture: texture, frameDelay: frameDelay, loopForever: loopForever, zIndex: zIndex, position: position, origin: origin, scale: bmath.NewVec2d(1, 1), flip: bmath.NewVec2d(1, 1), color: color{1, 1, 1, 1}}
	sprite.transform = NewTransformations(sprite)

	var currentLoop *Loop = nil
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
				sprite.loopQueue = append(sprite.loopQueue, currentLoop)
				loopDepth = -1
			}
			if command[0] != "L" {
				sprite.transform.Add(NewCommand(command))
			}
		}

		if command[0] == "L" {

			currentLoop = NewLoop(command, sprite)

			loopDepth = removed + 1

		} else if removed == loopDepth {
			currentLoop.Add(NewCommand(command))
		}
	}

	if currentLoop != nil {
		sprite.loopQueue = append(sprite.loopQueue, currentLoop)
		loopDepth = -1
	}

	sprite.transform.Finalize()

	sprite.startTime = sprite.transform.startTime
	sprite.endTime = sprite.transform.endTime

	for _, loop := range sprite.loopQueue {
		if loop.start < sprite.startTime {
			sprite.startTime = loop.start
		}

		if loop.end > sprite.endTime {
			sprite.endTime = loop.end
		}
	}

	return sprite
}
 */

func (sprite *Sprite) Update(time int64) {
	sprite.currentFrame = 0

	if len(sprite.texture) > 1 {
		frame := int(math.Floor((float64(time) - sprite.startTime) / sprite.frameDelay))
		if !sprite.loopForever {
			if frame >= len(sprite.texture) {
				frame = len(sprite.texture) - 1
			}
			sprite.currentFrame = frame
		} else {
			sprite.currentFrame = frame % len(sprite.texture)
		}
	}

	for i := 0; i < len(sprite.transforms); i++ {
		transform := sprite.transforms[i]
		if float64(time) < transform.GetStartTime() {
			break
		}

		switch transform.GetType() {
		case animation.Fade, animation.Scale, animation.Rotate, animation.MoveX, animation.MoveY:
			value := transform.GetSingle(float64(time))
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
			x, y := transform.GetDouble(float64(time))
			switch transform.GetType() {
			case animation.Move:
				sprite.position.X = x
				sprite.position.Y = y
			case animation.ScaleVector:
				sprite.scale.X = x
				sprite.scale.Y = y
			}
		case animation.Additive, animation.HorizontalFlip, animation.VerticalFlip:
			value := transform.GetBoolean(float64(time))
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
			color := transform.GetColor(float64(time))
			sprite.color.R = color.R
			sprite.color.G = color.G
			sprite.color.B = color.B
			if transform.GetType() == animation.Color4 {
				sprite.color.A = color.A
			}
		}

		if float64(time) >= transform.GetEndTime() {
			sprite.transforms = append(sprite.transforms[:i], sprite.transforms[i+1:]...)
			i--
		}
	}
}

func (sprite *Sprite) AddTransform(transformation *animation.Transformation, sortAfter bool) {
	sprite.transforms = append(sprite.transforms, transformation)

	if sortAfter {
		sprite.SortTransformations()
	}
}

func (sprite *Sprite) SortTransformations() {
	sort.Slice(sprite.transforms, func(i, j int) bool {
		return sprite.transforms[i].GetStartTime() < sprite.transforms[j].GetStartTime()
	})
}

func (sprite *Sprite) AdjustTimesToTransformations() {
	startTime := math.MaxFloat64
	endTime := -math.MaxFloat64
	for _, t := range sprite.transforms {
		startTime = math.Min(startTime, t.GetStartTime())
		endTime = math.Max(endTime, t.GetEndTime())
	}
	sprite.startTime = startTime
	sprite.endTime = endTime
}

func (sprite *Sprite) ShowForever(value bool) {
	sprite.showForever = value
}

func (sprite *Sprite) UpdateAndDraw(time int64, batch *batches.SpriteBatch) {
	sprite.Update(time)
	sprite.Draw(time, batch)
}

func (sprite *Sprite) Draw(time int64, batch *batches.SpriteBatch) {
	if (!sprite.showForever && float64(time) < sprite.startTime && float64(time) >= sprite.endTime) || sprite.color.A < 0.01 {
		return
	}

	alpha := sprite.color.A
	if alpha > 1.001 {
		alpha -= math.Ceil(sprite.color.A) - 1
	}
	batch.DrawStObject(sprite.position, sprite.origin, sprite.scale.Abs(), sprite.flip, sprite.rotation, mgl32.Vec4{float32(sprite.color.R), float32(sprite.color.G), float32(sprite.color.B), float32(alpha)}, sprite.additive, *sprite.texture[sprite.currentFrame], false)
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

func (sprite *Sprite) SetScale(scale float64) {
	sprite.scale.X = scale
	sprite.scale.Y = scale
	sprite.dirty = true
}

func (sprite *Sprite) SetScaleV(vec bmath.Vector2d) {
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

func (sprite *Sprite) SetColor(color bmath.Color) {
	sprite.color.R, sprite.color.G, sprite.color.B = color.R, color.G, color.B
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

func (sprite *Sprite) GetStartTime() float64 {
	return sprite.startTime
}

func (sprite *Sprite) GetEndTime() float64 {
	return sprite.endTime
}

func (sprite *Sprite) GetDepth() float64 {
	return sprite.depth
}

func (sprite *Sprite) GetLoad() float64 {
	if sprite.color.A >= 0.01 {
		return math.Min((float64(sprite.texture[0].Width)*sprite.scale.X*float64(sprite.texture[0].Height)*sprite.scale.Y)/storyboardArea, maxLoad)
	}
	return 0
}
