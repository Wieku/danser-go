package objects

import (
	"github.com/wieku/danser-go/app/audio"
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
)

type IHitObject interface {
	Update(time float64) bool
	SetTiming(timings *Timings, beatmapVersion int, diffCalcOnly bool) //diffCalcOnly skips stables' path generation which is quite memory consuming
	SetDifficulty(difficulty *difficulty.Difficulty)

	GetStartTime() float64
	GetEndTime() float64
	GetDuration() float64

	GetPositionAt(float64) vector.Vector2f
	GetStackedPositionAtMod(time float64, diff *difficulty.Difficulty) vector.Vector2f

	GetStartPosition() vector.Vector2f
	GetStackedStartPositionMod(diff *difficulty.Difficulty) vector.Vector2f

	GetEndPosition() vector.Vector2f
	GetStackedEndPositionMod(diff *difficulty.Difficulty) vector.Vector2f

	GetID() int64
	SetID(int64)
	SetComboNumber(cn int64)
	GetComboSet() int64
	SetComboSet(set int64)
	GetComboSetHax() int64
	SetComboSetHax(set int64)

	GetStackIndex(stackThreshold int64) int64
	GetStackIndexMod(diff *difficulty.Difficulty) int64
	SetStackIndex(stackThreshold, index int64)
	SetStackLeniency(leniency float64)

	GetColorOffset() int64
	IsLastCombo() bool
	SetLastInCombo(b bool)
	IsNewCombo() bool
	SetNewCombo(b bool)

	GetType() Type

	DisableAudioSubmission(value bool)

	Finalize()
}

type ILongObject interface {
	IHitObject

	GetStartAngleMod(diff *difficulty.Difficulty) float32

	GetEndAngleMod(diff *difficulty.Difficulty) float32

	GetPartLen() float32
}

type HitObject struct {
	StartPosRaw vector.Vector2f
	EndPosRaw   vector.Vector2f

	StartTime float64
	EndTime   float64

	StackLeniency float64
	StackIndexMap map[int64]int64

	PositionDelegate func(time float64) vector.Vector2f

	HitObjectID int64

	LastInCombo bool
	NewCombo    bool
	ComboNumber int64
	ComboSet    int64
	ComboSetHax int64
	ColorOffset int64

	BasicHitSound           audio.HitSoundInfo
	audioSubmissionDisabled bool
}

func (hitObject *HitObject) Update(_ float64) bool { return true }

func (hitObject *HitObject) SetTiming(_ *Timings, _ int, _ bool) {}

func (hitObject *HitObject) UpdateStacking() {}

func (hitObject *HitObject) SetDifficulty(_ *difficulty.Difficulty) {}

func (hitObject *HitObject) GetStartTime() float64 {
	return hitObject.StartTime
}

func (hitObject *HitObject) GetEndTime() float64 {
	return hitObject.EndTime
}

func (hitObject *HitObject) GetDuration() float64 {
	return hitObject.EndTime - hitObject.StartTime
}

func (hitObject *HitObject) GetPositionAt(time float64) vector.Vector2f {
	if hitObject.PositionDelegate != nil {
		return hitObject.PositionDelegate(time)
	}

	return hitObject.StartPosRaw
}

func (hitObject *HitObject) GetStackedPositionAtMod(time float64, diff *difficulty.Difficulty) vector.Vector2f {
	return ModifyPosition(hitObject, hitObject.GetPositionAt(time), diff)
}

func (hitObject *HitObject) GetStartPosition() vector.Vector2f {
	return hitObject.StartPosRaw
}

func (hitObject *HitObject) GetStackedStartPositionMod(diff *difficulty.Difficulty) vector.Vector2f {
	return ModifyPosition(hitObject, hitObject.GetStartPosition(), diff)
}

func (hitObject *HitObject) GetEndPosition() vector.Vector2f {
	return hitObject.EndPosRaw
}

func (hitObject *HitObject) GetStackedEndPositionMod(diff *difficulty.Difficulty) vector.Vector2f {
	return ModifyPosition(hitObject, hitObject.GetEndPosition(), diff)
}

func (hitObject *HitObject) GetID() int64 {
	return hitObject.HitObjectID
}

func (hitObject *HitObject) SetID(id int64) {
	hitObject.HitObjectID = id
}

func (hitObject *HitObject) SetComboNumber(cn int64) {
	hitObject.ComboNumber = cn
}

func (hitObject *HitObject) GetComboSet() int64 {
	return hitObject.ComboSet
}

func (hitObject *HitObject) SetComboSet(set int64) {
	hitObject.ComboSet = set
}

func (hitObject *HitObject) GetComboSetHax() int64 {
	return hitObject.ComboSetHax
}

func (hitObject *HitObject) SetComboSetHax(set int64) {
	hitObject.ComboSetHax = set
}

func (hitObject *HitObject) GetStackIndex(stackThreshold int64) int64 {
	return hitObject.StackIndexMap[stackThreshold]
}

func (hitObject *HitObject) GetStackIndexMod(diff *difficulty.Difficulty) int64 {
	stackThreshold := int64(math.Floor(diff.Preempt * hitObject.StackLeniency))

	return hitObject.StackIndexMap[stackThreshold]
}

func (hitObject *HitObject) SetStackIndex(stackThreshold, index int64) {
	hitObject.StackIndexMap[stackThreshold] = index
}

func (hitObject *HitObject) SetStackLeniency(leniency float64) {
	hitObject.StackLeniency = leniency
}

func (hitObject *HitObject) GetColorOffset() int64 {
	return hitObject.ColorOffset
}

func (hitObject *HitObject) IsLastCombo() bool {
	return hitObject.LastInCombo
}

func (hitObject *HitObject) SetLastInCombo(b bool) {
	hitObject.LastInCombo = b
}

func (hitObject *HitObject) IsNewCombo() bool {
	return hitObject.NewCombo
}

func (hitObject *HitObject) SetNewCombo(b bool) {
	hitObject.NewCombo = b
}

func (hitObject *HitObject) DisableAudioSubmission(value bool) {
	hitObject.audioSubmissionDisabled = value
}

func ModifyPosition(hitObject *HitObject, basePosition vector.Vector2f, diff *difficulty.Difficulty) vector.Vector2f {
	if diff.CheckModActive(difficulty.HardRock) {
		basePosition.Y = 384 - basePosition.Y
	}

	stackIndex := hitObject.GetStackIndexMod(diff)
	stackOffset := float32(stackIndex) * float32(diff.CircleRadius) / 10

	return basePosition.SubS(stackOffset, stackOffset)
}

func (hitObject *HitObject) Finalize() {}
