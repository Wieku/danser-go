package sprite

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/texture"
	"github.com/wieku/danser-go/framework/math/animation"
	color2 "github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
	"sort"
)

type Sprite struct {
	Texture *texture.TextureRegion

	transforms []*animation.Transformation

	startTime   float64
	endTime     float64
	showForever bool

	depth float64

	position vector.Vector2d
	origin   vector.Vector2d
	scale    vector.Vector2d
	flipX    bool
	flipY    bool
	rotation float64
	color    color2.Color
	additive bool

	cutX      float64
	cutY      float64
	cutOrigin vector.Vector2d
}

func NewSpriteSingle(tex *texture.TextureRegion, depth float64, position vector.Vector2d, origin vector.Vector2d) *Sprite {
	return &Sprite{
		Texture:     tex,
		transforms:  make([]*animation.Transformation, 0),
		depth:       depth,
		position:    position,
		origin:      origin,
		scale:       vector.NewVec2d(1, 1),
		color:       color2.NewL(1),
		showForever: true,
	}
}

func (sprite *Sprite) Update(time float64) {
	for i := 0; i < len(sprite.transforms); i++ {
		transform := sprite.transforms[i]
		if time < transform.GetStartTime() {
			break
		}

		sprite.updateTransform(transform, time)

		if time >= transform.GetEndTime() {
			copy(sprite.transforms[i:], sprite.transforms[i+1:])
			sprite.transforms = sprite.transforms[:len(sprite.transforms)-1]
			i--
		}
	}
}

func (sprite *Sprite) updateTransform(transform *animation.Transformation, time float64) { //nolint:gocyclo
	switch transform.GetType() {
	case animation.Fade, animation.Scale, animation.Rotate, animation.MoveX, animation.MoveY:
		value := transform.GetSingle(time)

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

		switch transform.GetType() {
		case animation.Additive:
			sprite.additive = value
		case animation.HorizontalFlip:
			sprite.flipX = value
		case animation.VerticalFlip:
			sprite.flipY = value
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
	if len(sprite.transforms) == 0 {
		return
	}

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
			sprite.updateTransform(t, t.GetStartTime()-1)

			applied[t.GetType()] = 1
		}
	}
}

func (sprite *Sprite) ShowForever(value bool) {
	sprite.showForever = value
}

func (sprite *Sprite) IsAlwaysVisible() bool {
	return sprite.showForever
}

func (sprite *Sprite) UpdateAndDraw(time float64, batch *batch.QuadBatch) {
	sprite.Update(time)
	sprite.Draw(time, batch)
}

func (sprite *Sprite) Draw(time float64, batch *batch.QuadBatch) {
	if (!sprite.showForever && time < sprite.startTime && time >= sprite.endTime) || sprite.color.A < 0.01 {
		return
	}

	if sprite.Texture == nil {
		return
	}

	alpha := sprite.color.A
	if alpha > 1.001 {
		alpha -= math32.Ceil(sprite.color.A) - 1 // HACK, some osu! storyboards use alpha higher than 1 to make flashing effect
	}

	region := *sprite.Texture
	position := sprite.position
	scale := sprite.scale.Abs()


	if sprite.cutX > 0.0 {
		if math.Abs(sprite.origin.X-sprite.cutOrigin.X) > 0 {
			position.X -= sprite.origin.X * float64(region.Width) * scale.X * sprite.cutX
		}

		ratio := float32(1 - sprite.cutX)
		middle := float32(sprite.cutOrigin.X)/2*math32.Abs(region.U2-region.U1) + (region.U1+region.U2)/2

		region.Width = region.Width * ratio
		region.U1 = (region.U1-middle)*ratio + middle
		region.U2 = (region.U2-middle)*ratio + middle
	}

	if sprite.cutY > 0.0 {
		if math.Abs(sprite.origin.Y-sprite.cutOrigin.Y) > 0 {
			position.Y -= sprite.origin.Y * float64(region.Height) * scale.Y * sprite.cutY
		}

		ratio := float32(1 - sprite.cutY)
		middle := float32(sprite.cutOrigin.Y)/2*math32.Abs(region.V2-region.V1) + (region.V1+region.V2)/2

		region.Height = region.Height * ratio
		region.V1 = (region.V1-middle)*ratio + middle
		region.V2 = (region.V2-middle)*ratio + middle
	}

	batch.DrawStObject(position, sprite.origin, scale, sprite.flipX, sprite.flipY, sprite.rotation, color2.NewRGBA(sprite.color.R, sprite.color.G, sprite.color.B, alpha), sprite.additive, region)
}

func (sprite *Sprite) GetOrigin() vector.Vector2d {
	return sprite.origin
}

func (sprite *Sprite) SetOrigin(origin vector.Vector2d) {
	sprite.origin = origin
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
	return mgl32.Vec3{sprite.color.R, sprite.color.G, sprite.color.B}
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

func (sprite *Sprite) SetStartTime(startTime float64) {
	sprite.startTime = startTime
}

func (sprite *Sprite) GetEndTime() float64 {
	return sprite.endTime
}

func (sprite *Sprite) SetEndTime(endTime float64) {
	sprite.endTime = endTime
}

func (sprite *Sprite) GetDepth() float64 {
	return sprite.depth
}
