package settings

var Audio = initAudio()

func initAudio() *audio {
	return &audio{
		GeneralVolume:              0.5,
		MusicVolume:                0.5,
		SampleVolume:               0.5,
		Offset:                     0,
		OnlineOffset:               false,
		HitsoundPositionMultiplier: 1.0,
		IgnoreBeatmapSamples:       false,
		IgnoreBeatmapSampleVolume:  false,
		PlayNightcoreSamples:       true,
		BeatScale:                  1.2,
		BeatUseTimingPoints:        false,
		NonWindows: &nonWindows{
			BassPlaybackBufferLength: 100,
			BassDeviceBufferLength:   10,
			BassUpdatePeriod:         5,
			BassDeviceUpdatePeriod:   10,
		},
	}
}

type audio struct {
	GeneralVolume              float64 `scale:"100.0" format:"%.0f%%"` //0.5
	MusicVolume                float64 `scale:"100.0" format:"%.0f%%"` //=0.5
	SampleVolume               float64 `scale:"100.0" format:"%.0f%%"` //=0.5
	Offset                     int64   `min:"-300" max:"300" format:"%dms" label:"Universal Offset"`
	OnlineOffset               bool    `label:"Apply online offset (needs API access)"`
	HitsoundPositionMultiplier float64
	IgnoreBeatmapSamples       bool        `label:"Ignore beatmap hitsounds"`       //= false
	IgnoreBeatmapSampleVolume  bool        `label:"Ignore hitsound volume changes"` //= false
	PlayNightcoreSamples       bool        `label:"Play nightcore beats" liveedit:"false"`
	BeatScale                  float64     `min:"1.0" max:"2.0"`
	BeatUseTimingPoints        bool        `label:"Add metronome to Beat scale"`
	NonWindows                 *nonWindows `json:"Linux/Unix" label:"Linux/Unix only" liveedit:"false"`
}

type nonWindows struct {
	BassPlaybackBufferLength int64 `max:"500"`
	BassDeviceBufferLength   int64
	BassUpdatePeriod         int64
	BassDeviceUpdatePeriod   int64
}
