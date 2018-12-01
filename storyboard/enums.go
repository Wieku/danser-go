package storyboard

import "danser/bmath"

var Origin = map[string]bmath.Vector2d{
	"0":            bmath.NewVec2d(-1, -1),
	"TopLeft":      bmath.NewVec2d(-1, -1),

	"1":            bmath.NewVec2d(0, 0),
	"Centre":       bmath.NewVec2d(0, 0),

	"2":            bmath.NewVec2d(-1, 0),
	"CentreLeft":   bmath.NewVec2d(-1, 0),

	"3":            bmath.NewVec2d(1, -1),
	"TopRight":     bmath.NewVec2d(1, -1),

	"4":            bmath.NewVec2d(0, 1),
	"BottomCentre": bmath.NewVec2d(0, 1),

	"5":            bmath.NewVec2d(0, -1),
	"TopCentre":    bmath.NewVec2d(0, -1),


	"7":            bmath.NewVec2d(1, 0),
	"CentreRight":  bmath.NewVec2d(1, 0),

	"8":            bmath.NewVec2d(-1, 1),
	"BottomLeft":   bmath.NewVec2d(-1, 1),

	"9":            bmath.NewVec2d(1, 1),
	"BottomRight":  bmath.NewVec2d(1, 1),
}

