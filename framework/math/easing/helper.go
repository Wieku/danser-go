package easing

var easings = []Easing{
	Linear,
	OutQuad,
	InQuad,
	InQuad,
	OutQuad,
	InOutQuad,
	InCubic,
	OutCubic,
	InOutCubic,
	InQuart,
	OutQuart,
	InOutQuart,
	InQuint,
	OutQuint,
	InOutQuint,
	InSine,
	OutSine,
	InOutSine,
	InExpo,
	OutExpo,
	InOutExpo,
	InCirc,
	OutCirc,
	InOutCirc,
	InElastic,
	OutElastic,
	OutHalfElastic,
	OutQuartElastic,
	InOutElastic,
	InBack,
	OutBack,
	InOutBack,
	InBounce,
	OutBounce,
	InOutBounce,
}

func GetEasing(easingID int64) Easing {
	if easingID < 0 || easingID >= int64(len(easings)) {
		easingID = 0
	}
	return easings[easingID]
}
