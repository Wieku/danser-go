package hitjudge

type HitResult int

const (
	Hit300	HitResult = 0
	Hit100	HitResult = 1
	Hit50	HitResult = 2
	HitMiss	HitResult = 3
)