package launcher

import (
	"encoding/json"
	"github.com/wieku/danser-go/app/beatmap"
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/rplpa"
	"golang.org/x/exp/constraints"
	"strconv"
	"strings"
)

type floatParam param[float32]
type intParam param[int32]

type param[T constraints.Integer | constraints.Float] struct {
	ogValue T
	value   T
	changed bool
}

type knockoutReplay struct {
	path         string
	parsedReplay *rplpa.Replay
	included     bool
}

type builder struct {
	currentMode  Mode
	currentPMode PMode

	outputName string
	ssTime     float32

	ar floatParam
	od floatParam
	cs floatParam
	hp floatParam

	speed floatParam
	pitch floatParam

	currentMap *beatmap.BeatMap

	replayPath    string
	currentReplay *rplpa.Replay

	offset intParam
	start  intParam
	end    intParam
	skip   bool

	mirrors int32
	tags    int32

	mods difficulty.Modifier

	config string

	knockoutReplays []*knockoutReplay
}

func newBuilder() *builder {
	return &builder{
		currentMode:  CursorDance,
		currentPMode: Watch,
		speed: floatParam{
			ogValue: 1,
			value:   1,
		},
		pitch: floatParam{
			ogValue: 1,
			value:   1,
		},
		offset: intParam{
			ogValue: 0,
			value:   0,
		},
		mirrors: 1,
		tags:    1,
		config:  "default",
	}
}

func (b *builder) setMap(bMap *beatmap.BeatMap) {
	b.currentMap = bMap

	b.ar = floatParam{
		ogValue: float32(bMap.Diff.GetAR()),
		value:   float32(bMap.Diff.GetAR()),
	}

	b.od = floatParam{
		ogValue: float32(bMap.Diff.GetOD()),
		value:   float32(bMap.Diff.GetOD()),
	}

	b.cs = floatParam{
		ogValue: float32(bMap.Diff.GetCS()),
		value:   float32(bMap.Diff.GetCS()),
	}

	b.hp = floatParam{
		ogValue: float32(bMap.Diff.GetHP()),
		value:   float32(bMap.Diff.GetHP()),
	}

	b.start = intParam{}

	mEnd := int32(math32.Ceil(float32(b.currentMap.Length) / 1000))

	b.end = intParam{
		ogValue: mEnd,
		value:   mEnd,
		changed: false,
	}

	b.offset.value = int32(bMap.LocalOffset)
	b.offset.changed = b.offset.value != 0
}

func (b *builder) numKnockoutReplays() (ret int) {
	if b.knockoutReplays != nil {
		for _, r := range b.knockoutReplays {
			if r.included {
				ret++
			}
		}
	}

	return
}

func (b *builder) getArguments() (args []string) {
	args = append(args, "-nodbcheck", "-noupdatecheck")

	if b.config != "default" {
		args = append(args, "-settings", b.config)
	}

	if b.currentMode == Replay {
		args = append(args, "-replay", b.replayPath)
	} else {
		args = append(args, "-md5", b.currentMap.MD5)

		mods := ""

		if b.currentMode == Play {
			args = append(args, "-play")
		} else if b.currentMode == Knockout {
			args = append(args, "-knockout")
		} else if b.currentMode == NewKnockout {
			var list []string

			for _, r := range b.knockoutReplays {
				if r.included {
					list = append(list, r.path)
				}
			}

			data, _ := json.Marshal(list)
			args = append(args, "-knockout2", string(data))
		} else if b.currentMode == DanserReplay {
			mods = "AT"
		}

		if b.mods != difficulty.None {
			mods += b.mods.String()
		}

		if mods != "" {
			args = append(args, "-mods", mods)
		}
	}

	if b.currentMode != Play && b.currentPMode != Watch {
		oEmpty := true

		if tr := strings.TrimSpace(b.outputName); tr != "" {
			oEmpty = false
			args = append(args, "-out", tr)
		}

		if b.currentPMode == Record {
			args = append(args, "-preciseprogress")
		}

		if b.currentPMode == Screenshot {
			args = append(args, "-ss", strconv.FormatFloat(float64(b.ssTime), 'f', 3, 32))
		} else if oEmpty {
			args = append(args, "-record")
		}
	}

	if b.currentMode != Replay {
		if b.ar.changed {
			args = append(args, "-ar", strconv.FormatFloat(float64(b.ar.value), 'f', 1, 32))
		}

		if b.od.changed {
			args = append(args, "-od", strconv.FormatFloat(float64(b.od.value), 'f', 1, 32))
		}

		if b.cs.changed {
			args = append(args, "-cs", strconv.FormatFloat(float64(b.cs.value), 'f', 1, 32))
		}

		if b.hp.changed {
			args = append(args, "-hp", strconv.FormatFloat(float64(b.hp.value), 'f', 1, 32))
		}
	}

	if b.currentMode == CursorDance {
		if b.mirrors > 1 {
			args = append(args, "-cursors", strconv.Itoa(int(b.mirrors)))
		}

		if b.tags > 1 {
			args = append(args, "-tag", strconv.Itoa(int(b.tags)))
		}
	}

	if b.start.changed {
		args = append(args, "-start", strconv.FormatFloat(float64(b.start.value), 'f', 1, 32))
	}

	if b.end.changed {
		args = append(args, "-end", strconv.FormatFloat(float64(b.end.value), 'f', 1, 32))
	}

	if b.speed.changed {
		args = append(args, "-speed", strconv.FormatFloat(float64(b.speed.value), 'f', 2, 32))
	}

	if b.pitch.changed {
		args = append(args, "-pitch", strconv.FormatFloat(float64(b.pitch.value), 'f', 2, 32))
	}

	if b.skip {
		args = append(args, "-skip")
	}

	if b.offset.changed && b.currentPMode != Screenshot {
		args = append(args, "-offset", strconv.Itoa(int(b.offset.value)))
	}

	return
}
