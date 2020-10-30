package storyboard

import (
	"github.com/wieku/danser-go/framework/math/vector"
)

var Origin = map[string]vector.Vector2d{
	"0":       vector.NewVec2d(-1, -1),
	"TopLeft": vector.NewVec2d(-1, -1),

	"1":      vector.NewVec2d(0, 0),
	"Centre": vector.NewVec2d(0, 0),

	"2":          vector.NewVec2d(-1, 0),
	"CentreLeft": vector.NewVec2d(-1, 0),

	"3":        vector.NewVec2d(1, -1),
	"TopRight": vector.NewVec2d(1, -1),

	"4":            vector.NewVec2d(0, 1),
	"BottomCentre": vector.NewVec2d(0, 1),

	"5":         vector.NewVec2d(0, -1),
	"TopCentre": vector.NewVec2d(0, -1),

	"7":           vector.NewVec2d(1, 0),
	"CentreRight": vector.NewVec2d(1, 0),

	"8":          vector.NewVec2d(-1, 1),
	"BottomLeft": vector.NewVec2d(-1, 1),

	"9":           vector.NewVec2d(1, 1),
	"BottomRight": vector.NewVec2d(1, 1),
}
