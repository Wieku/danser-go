package sliderrenderer

import (
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/math/vector"
)

func createUnitCircle(segments int) ([]float32, []uint16) {
	points := make([]float32, (segments+1)*3)
	indices := make([]uint16, segments*3)

	for i := 0; i < segments; i++ {
		p := vector.NewVec2fRad(float32(i)/float32(segments)*2*math32.Pi, 1)

		j := i * 3

		points[j+3], points[j+4], points[j+5] = p.X, p.Y, 1.0

		indices[j], indices[j+1], indices[j+2] = 0, uint16(i+1), uint16(i+2)

		if i == segments-1 { // loop
			indices[j+2] = 1
		}
	}

	return points, indices
}

func createUnitLine() ([]float32, []uint16) {
	vertices := make([]float32, 3*6)
	indices := make([]uint16, 12)

	vertices[0], vertices[1], vertices[2] = 0, 1, 1
	vertices[3], vertices[4], vertices[5] = 1, 1, 1

	vertices[6], vertices[7], vertices[8] = 0, 0, 0
	vertices[9], vertices[10], vertices[11] = 1, 0, 0

	vertices[12], vertices[13], vertices[14] = 0, -1, 1
	vertices[15], vertices[16], vertices[17] = 1, -1, 1

	indices[0], indices[1], indices[2] = 2, 0, 1
	indices[3], indices[4], indices[5] = 1, 3, 2

	indices[6], indices[7], indices[8] = 4, 2, 3
	indices[9], indices[10], indices[11] = 3, 5, 4

	return vertices, indices
}
