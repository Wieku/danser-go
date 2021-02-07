package oppai

import (
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/framework/math/vector"
)

type DiffObject struct {
	Data    objects.IHitObject
	Normpos vector.Vector2d
	Angle   float64
	Strains []float64
	//IsSingle bool
	DeltaTime float64
	DDistance float64
}
