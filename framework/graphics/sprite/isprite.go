package sprite

import (
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/math/animation"
	color2 "github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/vector"
)

type ISprite interface {
	Update(time float64)

	AddTransform(transformation *animation.Transformation)

	AddTransforms(transformations []*animation.Transformation)

	AddTransformUnordered(transformation *animation.Transformation)

	AddTransformsUnordered(transformations []*animation.Transformation)

	SortTransformations()

	ClearTransformations()

	ClearTransformationsOfType(transformationType animation.TransformationType)

	AdjustTimesToTransformations()

	ResetValuesToTransforms()

	ShowForever(value bool)

	IsAlwaysVisible() bool

	UpdateAndDraw(time float64, batch *batch.QuadBatch)

	Draw(time float64, batch *batch.QuadBatch)

	GetOrigin() vector.Vector2d

	SetOrigin(origin vector.Vector2d)

	GetPosition() vector.Vector2d

	SetPosition(vec vector.Vector2d)

	GetScale() vector.Vector2d

	SetScale(scale float64)

	SetScaleV(vec vector.Vector2d)

	GetRotation() float64

	SetRotation(rad float64)

	GetColor() color2.Color

	SetColor(color color2.Color)

	GetAlpha32() float32

	GetAlpha() float64

	SetAlpha(alpha float32)

	SetHFlip(on bool)

	SetVFlip(on bool)

	SetCutX(cutX float64)

	SetCutY(cutY float64)

	SetCutOrigin(origin vector.Vector2d)

	SetAdditive(on bool)

	GetStartTime() float64

	SetStartTime(startTime float64)

	GetEndTime() float64

	SetEndTime(endTime float64)

	GetDepth() float64
}
