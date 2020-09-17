package sliderrenderer

import (
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/math/vector"
)

func createUnitCircle(segments int) []float32 {
	points := make([]vector.Vector2f, segments+1)

	for i := 0; i <= segments; i++ {
		points[i] = vector.NewVec2fRad(float32(i)/float32(segments)*2*math32.Pi, 1)
	}

	unitCircle := make([]float32, 9*segments)

	base := 0
	for j := 1; j < len(points); j++ {
		p1, p2 := points[j-1], points[j]

		unitCircle[base], unitCircle[base+1], unitCircle[base+2] = p1.X, p1.Y, 1.0

		base += 3

		unitCircle[base], unitCircle[base+1], unitCircle[base+2] = p2.X, p2.Y, 1.0

		base += 6
	}

	return unitCircle
}
