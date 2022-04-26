package vector

// DON'T TOUCH THIS
var (
	TopLeft      = Vector2d{-1, -1}
	Centre       = Vector2d{0, 0} //nolint:misspell
	CentreLeft   = Vector2d{-1, 0}
	TopRight     = Vector2d{1, -1}
	BottomCentre = Vector2d{0, 1}
	TopCentre    = Vector2d{0, -1}
	CentreRight  = Vector2d{1, 0}
	BottomLeft   = Vector2d{-1, 1}
	BottomRight  = Vector2d{1, 1}
)

func ParseOrigin(v string) Vector2d {
	switch v {
	case "TopLeft":
		return TopLeft
	case "Centre": //nolint:misspell
		return Centre //nolint:misspell
	case "CentreLeft", "Left":
		return CentreLeft
	case "TopRight":
		return TopRight
	case "BottomCentre", "Bottom":
		return BottomCentre
	case "TopCentre", "Top":
		return TopCentre
	case "CentreRight", "Right":
		return CentreRight
	case "BottomLeft":
		return BottomLeft
	case "BottomRight":
		return BottomRight
	default:
		return TopLeft
	}
}
