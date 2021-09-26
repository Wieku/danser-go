package common

import (
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/skin"
	"github.com/wieku/danser-go/framework/bass"
)

const barsPerSegment = 4
const nightCoreDivisor = 2

type NightcoreProcessor struct {
	*BeatSynced

	hasFirstValue bool
	firstValue    int

	pProgress int

	hatSample    *bass.Sample
	clapSample   *bass.Sample
	kickSample   *bass.Sample
	finishSample *bass.Sample
}

func NewNightcoreProcessor() *NightcoreProcessor {
	proc := &NightcoreProcessor{
		BeatSynced:   NewBeatSynced(),
		hatSample:    skin.GetSample("nightcore-hat"),
		clapSample:   skin.GetSample("nightcore-clap"),
		kickSample:   skin.GetSample("nightcore-kick"),
		finishSample: skin.GetSample("nightcore-finish"),
	}

	proc.Divisor = nightCoreDivisor

	return proc
}

func (bs *NightcoreProcessor) Update(time float64) {
	bs.BeatSynced.Update(time)

	segLength := bs.timingPoint.Signature * nightCoreDivisor * barsPerSegment

	if !bs.IsSynced {
		bs.hasFirstValue = false

		return
	}

	if !bs.hasFirstValue || bs.beatIndex < bs.firstValue {
		bs.hasFirstValue = true

		if bs.beatIndex < 0 {
			bs.firstValue = 0
		} else {
			bs.firstValue = (bs.beatIndex/segLength + 1) * segLength
		}
	}

	if bs.beatIndex >= bs.firstValue && bs.beatIndex != bs.pProgress {
		bs.playBeat(bs.beatIndex%segLength, bs.timingPoint.Signature)
	}

	bs.pProgress = bs.beatIndex
}

func (bs *NightcoreProcessor) playBeat(beatIndex int, signature int) { //nolint:gocyclo
	if !settings.Audio.PlayNightcoreSamples {
		return
	}

	if beatIndex == 0 && bs.finishSample != nil {
		bs.finishSample.Play()
	}

	switch signature {
	case 3:
		switch beatIndex % 6 {
		case 0:
			if bs.kickSample != nil {
				bs.kickSample.Play()
			}
		case 3:
			if bs.clapSample != nil {
				bs.clapSample.Play()
			}
		default:
			if bs.hatSample != nil {
				bs.hatSample.Play()
			}
		}
	case 4:
		switch beatIndex % 4 {
		case 0:
			if bs.kickSample != nil {
				bs.kickSample.Play()
			}
		case 2:
			if bs.clapSample != nil {
				bs.clapSample.Play()
			}
		default:
			if bs.hatSample != nil {
				bs.hatSample.Play()
			}
		}
	}
}
