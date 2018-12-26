package hitjudge

import "danser/bmath"

type ObjectResult struct {
	JudgePos  bmath.Vector2d
	JudgeTime int64
	Result    HitResult
	IsBreak   bool
}