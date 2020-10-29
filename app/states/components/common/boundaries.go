package common

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/graphics/shape"
	color2 "github.com/wieku/danser-go/framework/math/color"
)

type Boundaries struct {
	shapeRenderer *shape.Renderer
}

func NewBoundaries() *Boundaries {
	boundaries := new(Boundaries)
	boundaries.shapeRenderer = shape.NewRendererSize(18)

	return boundaries
}

func (boundaries *Boundaries) Draw(projection mgl32.Mat4, circleRadius, alpha float32) {
	if !settings.Gameplay.Boundaries.Enabled {
		return
	}

	thickness := float32(settings.Gameplay.Boundaries.BorderThickness)
	cs := circleRadius

	boundaries.shapeRenderer.SetCamera(projection)
	boundaries.shapeRenderer.Begin()

	if bAlpha := settings.Gameplay.Boundaries.BackgroundOpacity; bAlpha > 0.001 {
		colHSV := settings.Gameplay.Boundaries.BackgroundColor
		color := color2.NewHSVA(float32(colHSV.Hue), float32(colHSV.Saturation), float32(colHSV.Value), float32(bAlpha)*alpha)

		boundaries.shapeRenderer.SetColorM(color)
		boundaries.shapeRenderer.DrawQuad(-cs, -cs, bmath.OsuWidth+cs, -cs, bmath.OsuWidth+cs, bmath.OsuHeight+cs, -cs, bmath.OsuHeight+cs)
	}

	if bAlpha := settings.Gameplay.Boundaries.BorderOpacity; bAlpha > 0.001 {
		colHSV := settings.Gameplay.Boundaries.BorderColor
		color := color2.NewHSVA(float32(colHSV.Hue), float32(colHSV.Saturation), float32(colHSV.Value), float32(bAlpha)*alpha)

		boundaries.shapeRenderer.SetColorM(color)

		half := thickness / 2

		csH := cs + half

		if settings.Gameplay.Boundaries.BorderFill > 0.99 {
			boundaries.shapeRenderer.DrawLine(-csH-half, -csH, bmath.OsuWidth+csH+half, -csH, thickness)
			boundaries.shapeRenderer.DrawLine(-csH-half, bmath.OsuHeight+csH, bmath.OsuWidth+csH+half, bmath.OsuHeight+csH, thickness)
		} else {
			dx := (bmath.OsuWidth + cs*2) / 2 * float32(settings.Gameplay.Boundaries.BorderFill)

			// top
			boundaries.shapeRenderer.DrawLine(-cs-thickness, -csH, -cs+dx, -csH, thickness)
			boundaries.shapeRenderer.DrawLine(bmath.OsuWidth+cs-dx, -csH, bmath.OsuWidth+cs+thickness, -csH, thickness)

			// bottom
			boundaries.shapeRenderer.DrawLine(-cs-thickness, bmath.OsuHeight+csH, -cs+dx, bmath.OsuHeight+csH, thickness)
			boundaries.shapeRenderer.DrawLine(bmath.OsuWidth+cs-dx, bmath.OsuHeight+csH, bmath.OsuWidth+cs+thickness, bmath.OsuHeight+csH, thickness)
		}

		if settings.Gameplay.Boundaries.BorderFill > bmath.OsuHeight/bmath.OsuWidth {
			boundaries.shapeRenderer.DrawLine(-csH, -csH+half, -csH, bmath.OsuHeight+csH-half, thickness)
			boundaries.shapeRenderer.DrawLine(bmath.OsuWidth+csH, -csH+half, bmath.OsuWidth+csH, bmath.OsuHeight+csH-half, thickness)
		} else {
			dy := (bmath.OsuWidth + cs*2) / 2 * float32(settings.Gameplay.Boundaries.BorderFill)

			// left
			boundaries.shapeRenderer.DrawLine(-csH, -cs, -csH, -cs+dy, thickness)
			boundaries.shapeRenderer.DrawLine(-csH, bmath.OsuHeight+cs-dy, -csH, bmath.OsuHeight+cs, thickness)

			// right
			boundaries.shapeRenderer.DrawLine(bmath.OsuWidth+csH, -cs, bmath.OsuWidth+csH, -cs+dy, thickness)
			boundaries.shapeRenderer.DrawLine(bmath.OsuWidth+csH, bmath.OsuHeight+cs-dy, bmath.OsuWidth+csH, bmath.OsuHeight+cs, thickness)
		}
	}

	boundaries.shapeRenderer.End()
}
