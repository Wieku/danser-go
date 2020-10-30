package scaling

import (
	"github.com/wieku/danser-go/framework/math/vector"
)

type Scaling int

const (
	// The source is not scaled.
	None = Scaling(iota)

	// Scales the source to fit the target while keeping the same aspect ratio. This may cause the source to be smaller than the
	// target in one direction.
	Fit

	// Scales the source to fill the target while keeping the same aspect ratio. This may cause the source to be larger than the
	// target in one direction.
	Fill

	// Scales the source to fill the target in the x direction while keeping the same aspect ratio. This may cause the source to be
	// smaller or larger than the target in the y direction.
	FillX

	// Scales the source to fill the target in the y direction while keeping the same aspect ratio. This may cause the source to be
	// smaller or larger than the target in the x direction.
	FillY

	// Scales the source to fill the target. This may cause the source to not keep the same aspect ratio.
	Stretch

	// Scales the source to fill the target in the x direction, without changing the y direction. This may cause the source to not
	// keep the same aspect ratio.
	StretchX

	// Scales the source to fill the target in the y direction, without changing the x direction. This may cause the source to not
	// keep the same aspect ratio.
	StretchY
)

// Returns the size of the source scaled to the target. Note the same Vector2 instance is always returned and should never be
// cached.
func (s Scaling) Apply(sourceX, sourceY, targetX, targetY float32) vector.Vector2f {
	var res vector.Vector2f

	switch s {
	case Fit:
		targetRatio := targetY / targetX
		sourceRatio := sourceY / sourceX

		scale := targetY / sourceY
		if targetRatio > sourceRatio {
			scale = targetX / sourceX
		}

		res.X = sourceX * scale
		res.Y = sourceY * scale
	case Fill:
		targetRatio := targetY / targetX
		sourceRatio := sourceY / sourceX

		scale := targetX / sourceX
		if targetRatio > sourceRatio {
			scale = targetY / sourceY
		}

		res.X = sourceX * scale
		res.Y = sourceY * scale
	case FillX:
		scale := targetX / sourceX
		res.X = sourceX * scale
		res.Y = sourceY * scale
	case FillY:
		scale := targetY / sourceY
		res.X = sourceX * scale
		res.Y = sourceY * scale
	case Stretch:
		res.X = targetX
		res.Y = targetY
	case StretchX:
		res.X = targetX
		res.Y = sourceY
	case StretchY:
		res.X = sourceX
		res.Y = targetY
	default:
		res.X = sourceX
		res.Y = sourceY
	}

	return res
}
