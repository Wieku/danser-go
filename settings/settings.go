package settings

import (
	"danser/utils"
	"github.com/go-gl/mathgl/mgl32"
)

const SETTINGSVERSION = "v1"

type general struct {
	OsuSongsDir 		string
	Players				int
	EnableBreakandQuit 	bool
	PlayerFadeTime		float64
	HitFadeTime			int64
	SameTimeOffset		float64
	BaseSize			float64
	BaseX				float64
	BaseY				float64
	SpinnerMult			float64
	ReverseFadeMult		float64
	SpinnerMinusTime	int64
	SaveResultCache 	bool
	ReadResultCache 	bool
	ReplayDir 			string
	CacheDir 			string
	MissMult			float64
	CursorColorNum		int
	Title	 			string
	Difficulty	 		string
	CursorColorSkipNum	int
	Recorder			string
	RecordTime			string
	RecordBaseX			float64
	RecordBaseY			float64
	RecordBaseSize		float64
	EnableDT			bool
	ShowMouse1			bool
	ShowMouse2			bool
	ErrorFixFile		string
}

type graphics struct {
	Width, Height             int64
	WindowWidth, WindowHeight int64
	Fullscreen                bool  //true
	VSync                     bool  //false
	FPSCap                    int64 //1000
	MSAA                      int32 //16
}

func (gr graphics) GetSize() (int64, int64) {
	if gr.Fullscreen {
		return gr.Width, gr.Height
	}
	return gr.WindowWidth, gr.WindowHeight
}

func (gr graphics) GetSizeF() (float64, float64) {
	if gr.Fullscreen {
		return float64(gr.Width), float64(gr.Height)
	}
	return float64(gr.WindowWidth), float64(gr.WindowHeight)
}

func (gr graphics) GetWidth() int64 {
	if gr.Fullscreen {
		return gr.Width
	}
	return gr.WindowWidth
}

func (gr graphics) GetWidthF() float64 {
	if gr.Fullscreen {
		return float64(gr.Width)
	}
	return float64(gr.WindowWidth)
}

func (gr graphics) GetHeight() int64 {
	if gr.Fullscreen {
		return gr.Height
	}
	return gr.WindowHeight
}

func (gr graphics) GetHeightF() float64 {
	if gr.Fullscreen {
		return float64(gr.Height)
	}
	return float64(gr.WindowHeight)
}

func (gr graphics) GetAspectRatio() float64 {
	if gr.Fullscreen {
		return float64(gr.Width) / float64(gr.Height)
	}
	return float64(gr.WindowWidth) / float64(gr.WindowHeight)
}

type audio struct {
	GeneralVolume             float64 //0.5
	MusicVolume               float64 //=0.5
	SampleVolume              float64 //=0.5
	Offset                    int64
	IgnoreBeatmapSamples      bool //= false
	IgnoreBeatmapSampleVolume bool //= false
}

type beat struct {
	BeatScale float64 //1.4
}

type hsv struct {
	Hue, Saturation, Value float64
}

type color struct {
	EnableRainbow         bool    //true
	RainbowSpeed          float64 //8, degrees per second
	BaseColor             *hsv    //0..360, if EnableRainbow is disabled then this value will be used to calculate base color
	EnableCustomHueOffset bool    //false, false means that every iteration has an offset of i*360/n
	HueOffset             float64 //0, custom hue offset for mirror collages
	FlashToTheBeat        bool    //true, objects size is changing with music peak amplitude
	FlashAmplitude        float64 //50, hue offset for flashes
	currentHue            float64
}

func (cl *color) Update(delta float64) {
	if cl.EnableRainbow {
		cl.currentHue += cl.RainbowSpeed / 1000.0 * delta
		for cl.currentHue >= 360.0 {
			cl.currentHue -= 360.0
		}

		for cl.currentHue < 0.0 {
			cl.currentHue += 360.0
		}
	} else {
		cl.currentHue = 0
	}
}

func (cl *color) GetColors(divides int, beatScale, alpha float64) []mgl32.Vec4 {
	flashOffset := 0.0
	if cl.FlashToTheBeat {
		flashOffset = cl.FlashAmplitude * (beatScale - 1.0) / (0.4 * Beat.BeatScale)
	}
	hue := cl.BaseColor.Hue + cl.currentHue + flashOffset

	for hue >= 360.0 {
		hue -= 360.0
	}

	for hue < 0.0 {
		hue += 360.0
	}

	offset := 360.0 / float64(divides)

	if cl.EnableCustomHueOffset {
		offset = cl.HueOffset
	}

	return utils.GetColorsSV(hue, offset, divides, cl.BaseColor.Saturation, cl.BaseColor.Value, alpha)
}

type bloom struct {
	Threshold, Blur, Power float64
}

type cursor struct {
	Colors                      *color
	EnableCustomTagColorOffset  bool    //true, if enabled, value set below will be used, if not, HueOffset of previous iteration will be used
	TagColorOffset              float64 //-36, offset of the next tag cursor
	EnableTrailGlow             bool    //true
	EnableCustomTrailGlowOffset bool    //true, if enabled, value set below will be used, if not, HueOffset of previous iteration will be used (or offset of 180° for single cursor)
	TrailGlowOffset             float64 //-36, offset of the cursor trail glow
	ScaleToCS                   bool    //false, if enabled, cursor will scale to beatmap CS value
	CursorSize                  float64 //18, cursor radius in osu!pixels
	ScaleToTheBeat              bool    //true, cursor size is changing with music peak amplitude
	ShowCursorsOnBreaks         bool    //true
	BounceOnEdges               bool    //false
	TrailEndScale               float64 //0.4
	TrailDensity                float64 //0.5 - 1/TrailDensity = distance between trail points
	TrailMaxLength              int64   //2000 - maximum width (in osu!pixels) of cursortrail
	TrailRemoveSpeed            float64 //1.0 - trail removal multiplier, 0.5 means half the speed
	GlowEndScale                float64 //0.4
	InnerLengthMult             float64 //0.9 - if glow is enabled, inner trail will be shortened to 0.9 * length
}

func (cr *cursor) GetColors(divides, tag int, beatScale, alpha float64) []mgl32.Vec4 {
	if !cr.EnableCustomTagColorOffset {
		return cr.Colors.GetColors(divides*tag, beatScale, alpha)
	}
	flashOffset := 0.0
	cl := cr.Colors
	if cl.FlashToTheBeat {
		flashOffset = cl.FlashAmplitude * (beatScale - 1.0) / (0.4 * Beat.BeatScale)
	}
	hue := cl.BaseColor.Hue + cl.currentHue + flashOffset

	for hue >= 360.0 {
		hue -= 360.0
	}

	for hue < 0.0 {
		hue += 360.0
	}

	offset := 360.0 / float64(divides)

	if cl.EnableCustomHueOffset {
		offset = cl.HueOffset
	}

	return utils.GetColorsSVT(hue, offset, cr.TagColorOffset, divides, tag, cl.BaseColor.Saturation, cl.BaseColor.Value, alpha)
}

type objects struct {
	MandalaTexturesTrigger                 int     //5, minimum value of cursors needed to use more translucent texture
	MandalaTexturesAlpha                   float64 //0.3
	ForceSliderBallTexture                 bool    //true, if this is disabled, mandala texture will be used for slider ball
	DrawApproachCircles                    bool    //true
	Colors                                 *color
	ObjectsSize                            float64 //-1, objects radius in osu!pixels. If value is less than 0, beatmap's CS will be used
	CSMult                                 float64 //1.2, if ObjectsSize is -1, then CS value will be multiplied by this
	ScaleToTheBeat                         bool    //true, objects size is changing with music peak amplitude
	SliderLOD                              int64   //30, number of triangles in a circle
	SliderPathLOD                          int64   //50, int(pixelLength*(PathLOD/100)) => number of slider path points
	SliderSnakeIn                          bool
	SliderSnakeOut                         bool
	SliderMerge                            bool
	DrawFollowPoints                       bool    //true
	WhiteFollowPoints                      bool    //true
	FollowPointColorOffset                 float64 //0.0, hue offset of the followpoint
	EnableCustomSliderBorderColor          bool
	CustomSliderBorderColor                *color
	EnableCustomSliderBorderGradientOffset bool
	SliderBorderGradientOffset             float64 //18, hue offset of slider outer border
	StackEnabled                           bool    //true, stack leniency
}

type playfield struct {
	LeadInTime           float64 //5
	LeadInHold           float64 //2
	FadeOutTime          float64 //5
	BackgroundInDim      float64 //0, background dim at the start of app
	BackgroundDim        float64 // 0.95, background dim at the beatmap start
	BackgroundDimBreaks  float64 // 0.95, background dim at the breaks
	BlurEnable           bool    //true
	BackgroundInBlur     float64 //0, background blur at the start of app
	BackgroundBlur       float64 // 0.6, background blur at the beatmap start
	BackgroundBlurBreaks float64 // 0.6, background blur at the breaks
	SpectrumInDim        float64 //0, background dim at the start of app
	SpectrumDim          float64 // 0.95, background dim at the beatmap start
	SpectrumDimBreaks    float64 // 0.95, background dim at the breaks
	StoryboardEnabled    bool
	Scale                float64 //1, scale the playfield (1 means that 384 will be rescaled to 900 on FullHD monitor)
	FlashToTheBeat       bool    //true, background dim varies accoriding to music power
	UnblurToTheBeat      bool    //true, background blur varies accoriding to music power
	UnblurFill           float64 //0.8, if blur is set to 0.6, then on full beat blur will be equal to 0.12
	KiaiFactor           float64 //1.2, scale and flash factor during Kiai
	BaseRotation         float64 //0, base rotation of playfield
	RotationEnabled      bool    //false
	RotationSpeed        float64 //2, degrees per second
	BloomEnabled         bool
	BloomToTheBeat       bool
	BloomBeatAddition    float64
	Bloom                *bloom
}

type fileformat struct {
	Version   *string
	General   *general
	Graphics  *graphics
	Audio     *audio
	Beat      *beat
	Cursor    *cursor
	Objects   *objects
	Playfield *playfield
	Dance     *dance
}

var Version string
var General *general
var Graphics *graphics
var Audio *audio
var Beat *beat
var Cursor *cursor
var Objects *objects
var Playfield *playfield
var Dance *dance

var DEBUG = false
var FPS = false
var DIVIDES = 2
var SPEED = 1.0
var PITCH = 1.0
var TAG = 1
