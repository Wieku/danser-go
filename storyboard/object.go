package storyboard

import (
	"github.com/Wieku/danser/bmath"
	"github.com/go-gl/mathgl/mgl32"
)

type Object interface {
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

	SetHFlip()
	SetVFlip()

	SetAdditive()
}
