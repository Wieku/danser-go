package sprite

import (
	"cmp"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/texture"
	"github.com/wieku/danser-go/framework/math/animation"
	color2 "github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
	"slices"
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

	cutX      vector.Vector2d
	cutY      vector.Vector2d
	cutOrigin vector.Vector2d

	nextTransformID int64
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
			if transform.IsLoop() {
				transform.UpdateLoop()

				n := sort.Search(len(sprite.transforms)-i-1, func(f int) bool {
					b := sprite.transforms[f+i+1]

					r := cmp.Compare(transform.GetStartTime(), b.GetStartTime())
					return r == -1 || (r == 0 && transform.GetID() < b.GetID())
				})

				if n != 0 {
					copy(sprite.transforms[i:], sprite.transforms[i+1:n+i+1])
					sprite.transforms[n+i] = transform

					i--
				}
			} else {
				copy(sprite.transforms[i:], sprite.transforms[i+1:])
				sprite.transforms = sprite.transforms[:len(sprite.transforms)-1]
				i--
			}
		}
	}
}

func (sprite *Sprite) updateTransform(transform *animation.Transformation, time float64) { //nolint:gocyclo
	switch transform.GetType() {
	case animation.Fade:
		sprite.color.A = float32(transform.GetSingle(time))
	case animation.Scale:
		s := transform.GetSingle(time)

		sprite.scale.X = s
		sprite.scale.Y = s
	case animation.Rotate:
		sprite.rotation = transform.GetSingle(time)
	case animation.MoveX:
		sprite.position.X = transform.GetSingle(time)
	case animation.MoveY:
		sprite.position.Y = transform.GetSingle(time)
	case animation.Move:
		sprite.position.X, sprite.position.Y = transform.GetDouble(time)
	case animation.ScaleVector:
		sprite.scale.X, sprite.scale.Y = transform.GetDouble(time)
	case animation.Additive:
		sprite.additive = transform.GetBoolean(time)
	case animation.HorizontalFlip:
		sprite.flipX = transform.GetBoolean(time)
	case animation.VerticalFlip:
		sprite.flipY = transform.GetBoolean(time)
	case animation.Color3:
		color := transform.GetColor(time)

		sprite.color.R = color.R
		sprite.color.G = color.G
		sprite.color.B = color.B
	case animation.Color4:
		sprite.color = transform.GetColor(time)
	}
}

func (sprite *Sprite) AddTransform(transformation *animation.Transformation) {
	sprite.AddTransformUnordered(transformation)
	sprite.SortTransformations()
}

func (sprite *Sprite) AddTransforms(transformations []*animation.Transformation) {
	sprite.AddTransformsUnordered(transformations)
	sprite.SortTransformations()
}

func (sprite *Sprite) AddTransformUnordered(transformation *animation.Transformation) {
	transformation.SetID(sprite.nextTransformID)
	sprite.nextTransformID++

	sprite.transforms = append(sprite.transforms, transformation)
}

func (sprite *Sprite) AddTransformsUnordered(transformations []*animation.Transformation) {
	for _, t := range transformations {
		t.SetID(sprite.nextTransformID)
		sprite.nextTransformID++

		sprite.transforms = append(sprite.transforms, t)
	}
}

func (sprite *Sprite) SortTransformations() {
	slices.SortFunc(sprite.transforms, func(a, b *animation.Transformation) int {
		r := cmp.Compare(a.GetStartTime(), b.GetStartTime())

		if r != 0 {
			return r
		}

		return cmp.Compare(a.GetID(), b.GetID())
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
		startTime = min(startTime, t.GetStartTime())
		endTime = max(endTime, t.GetTotalEndTime())
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
	flipX := sprite.flipX != (sprite.scale.X < 0) // XOR, flip again if scale is negative
	flipY := sprite.flipY != (sprite.scale.Y < 0) // XOR, flip again if scale is negative

	if sprite.cutX.X > 0.0 || sprite.cutX.Y > 0.0 {
		ratioL := float32(1 - sprite.cutX.X)
		ratioR := float32(1 - sprite.cutX.Y)

		oldU := math32.Abs(region.U2 - region.U1)

		middle := float32(sprite.cutOrigin.X)/2*oldU + (region.U1+region.U2)/2

		//region.Width = region.Width * ratio
		region.U1 = (region.U1-middle)*ratioL + middle
		region.U2 = (region.U2-middle)*ratioR + middle
		region.Width *= math32.Abs(region.U2-region.U1) / oldU
	}

	if sprite.cutY.X > 0.0 || sprite.cutY.Y > 0.0 {
		ratioT := float32(1 - sprite.cutY.X)
		ratioB := float32(1 - sprite.cutY.Y)

		oldV := math32.Abs(region.V2 - region.V1)

		middle := float32(sprite.cutOrigin.Y)/2*oldV + (region.V1+region.V2)/2

		//region.Height = region.Height * ratio
		region.V1 = (region.V1-middle)*ratioT + middle
		region.V2 = (region.V2-middle)*ratioB + middle
		region.Height *= math32.Abs(region.V2-region.V1) / oldV
	}

	batch.DrawStObject(position, sprite.origin, sprite.scale.Abs(), flipX, flipY, sprite.rotation, color2.NewRGBA(sprite.color.R, sprite.color.G, sprite.color.B, alpha), sprite.additive, region)
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

func (sprite *Sprite) GetColor() color2.Color {
	return sprite.color
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

func (sprite *Sprite) SetCutX(left, right float64) {
	sprite.cutX.X = left
	sprite.cutX.Y = right
}

func (sprite *Sprite) SetCutY(top, bottom float64) {
	sprite.cutY.X = top
	sprite.cutY.Y = bottom
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
