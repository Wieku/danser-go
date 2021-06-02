package storyboard

import (
	"github.com/wieku/danser-go/framework/math/vector"
)

func parseOrigin(v string) vector.Vector2d {
	switch v {
	case "0":
		return vector.TopLeft
	case "1":
		return vector.Centre
	case "2":
		return vector.CentreLeft
	case "3":
		return vector.TopRight
	case "4":
		return vector.BottomCentre
	case "5":
		return vector.TopCentre
	case "7":
		return vector.CentreRight
	case "8":
		return vector.BottomLeft
	case "9":
		return vector.BottomRight
	default:
		return vector.ParseOrigin(v)
	}
}
