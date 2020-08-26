package bmath

var Origin = struct {
	TopLeft,
	Centre,
	CentreLeft,
	TopRight,
	BottomCentre,
	TopCentre,
	CentreRight,
	BottomLeft,
	BottomRight Vector2d
}{
	NewVec2d(-1, -1),
	NewVec2d(0, 0),
	NewVec2d(-1, 0),
	NewVec2d(1, -1),
	NewVec2d(0, 1),
	NewVec2d(0, -1),
	NewVec2d(1, 0),
	NewVec2d(-1, 1),
	NewVec2d(1, 1),
}

type Transform struct {
	Position, Origin, Scale Vector2d
	Rotation                float64
}
