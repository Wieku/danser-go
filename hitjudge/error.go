package hitjudge

type Error struct {
	ReplayIndex	int
	ObjectIndex	int
	Result 		HitResult
	IsBreak		bool
	MaxComboOffset	int
	NowComboOffset	int
}