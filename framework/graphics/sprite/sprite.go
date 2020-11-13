package sprite

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/texture"
	"github.com/wieku/danser-go/framework/math/animation"
	color2 "github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
	"sort"
)

const (
	storyboardArea = 640.0 * 480.0
	maxLoad        = 1.3328125 //480*480*(16/9)/(640*480)
)

type Sprite struct {
	Textures     []*texture.TextureRegion
	frameDelay   float64
	loopForever  bool
	currentFrame int
	transforms   []*animation.Transformation

	startTime, endTime, depth float64

	position         vector.Vector2d
	positionRelative vector.Vector2d
	origin           vector.Vector2d
	scale            vector.Vector2d
	flipX            bool
	flipY            bool
	rotation         float64
	color            color2.Color
	additive         bool
	showForever      bool

	cutX      float64
	cutY      float64
	cutOrigin vector.Vector2d

	scaleTo vector.Vector2d
}

func NewSpriteSingle(tex *texture.TextureRegion, depth float64, position vector.Vector2d, origin vector.Vector2d) *Sprite {
	textures := []*texture.TextureRegion{tex}
	sprite := &Sprite{Textures: textures, frameDelay: 0.0, loopForever: true, depth: depth, position: position, origin: origin, scale: vector.NewVec2d(1, 1), color: color2.Color{1, 1, 1, 1}, showForever: true}
	sprite.transforms = make([]*animation.Transformation, 0)
	return sprite
}

func NewSpriteSingleCentered(tex *texture.TextureRegion, size vector.Vector2d) *Sprite {
	textures := []*texture.TextureRegion{tex}
	sprite := &Sprite{Textures: textures, frameDelay: 0.0, loopForever: true, depth: 0, origin: vector.NewVec2d(0, 0), scale: vector.NewVec2d(1, 1), color: color2.Color{1, 1, 1, 1}, showForever: true}
	sprite.transforms = make([]*animation.Transformation, 0)
	sprite.scaleTo = vector.NewVec2d(size.X/float64(tex.Width), size.Y/float64(tex.Height))
	return sprite
}

func NewSpriteSingleOrigin(tex *texture.TextureRegion, size vector.Vector2d, origin vector.Vector2d) *Sprite {
	textures := []*texture.TextureRegion{tex}
	sprite := &Sprite{Textures: textures, frameDelay: 0.0, loopForever: true, depth: 0, origin: origin, scale: vector.NewVec2d(1, 1), color: color2.Color{1, 1, 1, 1}, showForever: true}
	sprite.transforms = make([]*animation.Transformation, 0)
	sprite.scaleTo = vector.NewVec2d(size.X/float64(tex.Width), size.Y/float64(tex.Height))
	return sprite
}

func NewAnimation(textures []*texture.TextureRegion, frameDelay float64, loopForever bool, depth float64, position vector.Vector2d, origin vector.Vector2d) *Sprite {
	sprite := &Sprite{Textures: textures, frameDelay: frameDelay, loopForever: loopForever, depth: depth, position: position, origin: origin, scale: vector.NewVec2d(1, 1), color: color2.Color{1, 1, 1, 1}, showForever: true}
	sprite.transforms = make([]*animation.Transformation, 0)
	return sprite
}

func (sprite *Sprite) Update(time int64) {
	sprite.currentFrame = 0

	if sprite.Textures != nil && len(sprite.Textures) > 1 && float64(time) >= sprite.startTime {
		frame := int(math.Floor((float64(time) - sprite.startTime) / sprite.frameDelay))
		if !sprite.loopForever {
			frame = bmath.MinI(frame, len(sprite.Textures)-1)
		} else {
			frame = frame % len(sprite.Textures)
		}

		sprite.currentFrame = frame
	}

	for i := 0; i < len(sprite.transforms); i++ {
		transform := sprite.transforms[i]
		if float64(time) < transform.GetStartTime() {
			break
		}

		sprite.updateTransform(transform, time)

		if float64(time) >= transform.GetEndTime() {
			copy(sprite.transforms[i:], sprite.transforms[i+1:])
			sprite.transforms = sprite.transforms[:len(sprite.transforms)-1]
			i--
		}
	}
}

func (sprite *Sprite) updateTransform(transform *animation.Transformation, time int64) {
	switch transform.GetType() {
	case animation.Fade, animation.Scale, animation.Rotate, animation.MoveX, animation.MoveY:
		value := transform.GetSingle(float64(time))
		switch transform.GetType() {
		case animation.Fade:
			sprite.color.A = float32(value)
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
		switch transform.GetType() {
		case animation.Additive:
			sprite.additive = value
		case animation.HorizontalFlip:
			sprite.flipX = value
		case animation.VerticalFlip:
			sprite.flipY = value
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
}

func (sprite *Sprite) AddTransform(transformation *animation.Transformation) {
	sprite.transforms = append(sprite.transforms, transformation)

	sprite.SortTransformations()
}

func (sprite *Sprite) AddTransforms(transformations []*animation.Transformation) {
	sprite.transforms = append(sprite.transforms, transformations...)

	sprite.SortTransformations()
}

func (sprite *Sprite) AddTransformUnordered(transformation *animation.Transformation) {
	sprite.transforms = append(sprite.transforms, transformation)
}

func (sprite *Sprite) AddTransformsUnordered(transformations []*animation.Transformation) {
	sprite.transforms = append(sprite.transforms, transformations...)
}

func (sprite *Sprite) SortTransformations() {
	sort.SliceStable(sprite.transforms, func(i, j int) bool {
		return sprite.transforms[i].GetStartTime() < sprite.transforms[j].GetStartTime()
	})
}

func (sprite *Sprite) ClearTransformations() {
	sprite.transforms = make([]*animation.Transformation, 0)
}

func (sprite *Sprite) ClearTransformationsOfType(transformationType animation.TransformationType) {
	for i := 0; i < len(sprite.transforms); i++ {
		t := sprite.transforms[i]
		if t.GetType() == transformationType {
			copy(sprite.transforms[i:], sprite.transforms[i+1:])
			sprite.transforms = sprite.transforms[:len(sprite.transforms)-1]
			i--
		}
	}
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

func (sprite *Sprite) ResetValuesToTransforms() {
	applied := make(map[animation.TransformationType]int)

	for _, t := range sprite.transforms {
		if _, exists := applied[t.GetType()]; !exists {
			sprite.updateTransform(t, int64(t.GetStartTime()-1))

			applied[t.GetType()] = 1
		}
	}
}

func (sprite *Sprite) ShowForever(value bool) {
	sprite.showForever = value
}

func (sprite *Sprite) UpdateAndDraw(time int64, batch *batch.QuadBatch) {
	sprite.Update(time)
	sprite.Draw(time, batch)
}

func (sprite *Sprite) Draw(time int64, batch *batch.QuadBatch) {
	if (!sprite.showForever && float64(time) < sprite.startTime && float64(time) >= sprite.endTime) || sprite.color.A < 0.01 {
		return
	}

	if sprite.Textures == nil || len(sprite.Textures) == 0 || sprite.Textures[sprite.currentFrame] == nil {
		return
	}

	alpha := sprite.color.A
	if alpha > 1.001 {
		alpha -= math32.Ceil(sprite.color.A) - 1
	}

	scaleX := 1.0
	if sprite.scaleTo.X > 0 {
		scaleX = sprite.scaleTo.X
	}

	scaleY := 1.0
	if sprite.scaleTo.Y > 0 {
		scaleY = sprite.scaleTo.Y
	}

	region := *sprite.Textures[sprite.currentFrame]
	position := sprite.position

	if sprite.cutX > 0.0 {
		if math.Abs(sprite.origin.X-sprite.cutOrigin.X) > 0 {
			position.X -= sprite.origin.X * float64(region.Width) * sprite.cutX
		}

		ratio := float32(1 - sprite.cutX)
		middle := float32(sprite.cutOrigin.X)/2*math32.Abs(region.U2-region.U1) + (region.U1+region.U2)/2

		region.Width = int32(float32(region.Width) * ratio)
		region.U1 = (region.U1-middle)*ratio + middle
		region.U2 = (region.U2-middle)*ratio + middle
	}

	if sprite.cutY > 0.0 {
		if math.Abs(sprite.origin.Y-sprite.cutOrigin.Y) > 0 {
			position.Y -= sprite.origin.Y * float64(region.Height) * sprite.cutY
		}

		ratio := float32(1 - sprite.cutY)
		middle := float32(sprite.cutOrigin.Y)/2*math32.Abs(region.V2-region.V1) + (region.V1+region.V2)/2

		region.Height = int32(float32(region.Height) * ratio)
		region.V1 = (region.V1-middle)*ratio + middle
		region.V2 = (region.V2-middle)*ratio + middle
	}

	batch.DrawStObject(position, sprite.origin, sprite.scale.Abs().Mult(vector.NewVec2d(scaleX, scaleY)), sprite.flipX, sprite.flipY, sprite.rotation, mgl32.Vec4{float32(sprite.color.R), float32(sprite.color.G), float32(sprite.color.B), float32(alpha)}, sprite.additive, region)
}

func (sprite *Sprite) GetPosition() vector.Vector2d {
	return sprite.position
}

func (sprite *Sprite) SetPosition(vec vector.Vector2d) {
	sprite.position = vec
}

func (sprite *Sprite) GetScale() vector.Vector2d {
	return sprite.scale
}

func (sprite *Sprite) SetScale(scale float64) {
	sprite.scale.X = scale
	sprite.scale.Y = scale
}

func (sprite *Sprite) SetScaleV(vec vector.Vector2d) {
	sprite.scale = vec
}

func (sprite *Sprite) GetRotation() float64 {
	return sprite.rotation
}

func (sprite *Sprite) SetRotation(rad float64) {
	sprite.rotation = rad
}

func (sprite *Sprite) GetColor() mgl32.Vec3 {
	return mgl32.Vec3{float32(sprite.color.R), float32(sprite.color.G), float32(sprite.color.B)}
}

func (sprite *Sprite) SetColor(color color2.Color) {
	sprite.color.R, sprite.color.G, sprite.color.B = color.R, color.G, color.B
}

func (sprite *Sprite) GetAlpha32() float32 {
	return sprite.color.A
}

func (sprite *Sprite) GetAlpha() float64 {
	return float64(sprite.color.A)
}

func (sprite *Sprite) SetAlpha(alpha float32) {
	sprite.color.A = alpha
}

func (sprite *Sprite) SetHFlip(on bool) {
	sprite.flipX = on
}

func (sprite *Sprite) SetVFlip(on bool) {
	sprite.flipY = on
}

func (sprite *Sprite) SetCutX(cutX float64) {
	sprite.cutX = cutX
}

func (sprite *Sprite) SetCutY(cutY float64) {
	sprite.cutY = cutY
}

func (sprite *Sprite) SetCutOrigin(origin vector.Vector2d) {
	sprite.cutOrigin = origin
}

func (sprite *Sprite) SetAdditive(on bool) {
	sprite.additive = on
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
		return math.Min((float64(sprite.Textures[0].Width)*sprite.scale.X*float64(sprite.Textures[0].Height)*sprite.scale.Y)/storyboardArea, maxLoad)
	}
	return 0
}
