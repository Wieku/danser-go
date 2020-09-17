package bmath

import "github.com/wieku/danser-go/framework/math/vector"

var Origin = struct {
	TopLeft,
	Centre,
	CentreLeft,
	TopRight,
	BottomCentre,
	TopCentre,
	CentreRight,
	BottomLeft,
	BottomRight vector.Vector2d
}{
	vector.NewVec2d(-1, -1),
	vector.NewVec2d(0, 0),
	vector.NewVec2d(-1, 0),
	vector.NewVec2d(1, -1),
	vector.NewVec2d(0, 1),
	vector.NewVec2d(0, -1),
	vector.NewVec2d(1, 0),
	vector.NewVec2d(-1, 1),
	vector.NewVec2d(1, 1),
}

type Transform struct {
	Position, Origin, Scale vector.Vector2d
	Rotation                float64
}
