package launcher

import (
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/app/beatmap"
	"github.com/wieku/danser-go/app/rulesets/osu/performance"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/api"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/states/components/overlays/play"
	"github.com/wieku/danser-go/framework/goroutines"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/buffer"
	"github.com/wieku/danser-go/framework/graphics/viewport"
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/util"
)

type timePopup struct {
	*popup

	bld *builder

	graphStatus string

	timeCMap *beatmap.BeatMap
	peaks    api.StrainPeaks
	sGraph   *play.StrainGraph

	fbo *buffer.Framebuffer

	oldScale float32

	fBatch *batch.QuadBatch

	lastStart int32
	lastEnd   int32
}

func newTimePopup(bld *builder) *timePopup {
	mP := &timePopup{
		popup:       newPopup("Time/Offset Menu", popMedium),
		bld:         bld,
		graphStatus: "No map selected",
		fBatch:      batch.NewQuadBatch(),
	}

	mP.internalDraw = mP.drawTimeMenu

	settings.Gameplay.StrainGraph.XPosition = 0
	settings.Gameplay.StrainGraph.YPosition = 0
	settings.Gameplay.StrainGraph.Align = "TopLeft"

	settings.Gameplay.StrainGraph.FgColor = &settings.HSV{1, 0, 1}
	settings.Gameplay.StrainGraph.BgColor = &settings.HSV{1, 0, 0.5}

	settings.Gameplay.StrainGraph.Outline.Show = true
	settings.Gameplay.StrainGraph.Outline.Width = 3
	settings.Gameplay.StrainGraph.Outline.InnerDarkness = 0.5
	settings.Gameplay.StrainGraph.Outline.InnerOpacity = 1

	return mP
}

func (m *timePopup) drawTimeMenu() {
	imgui.Checkbox("Skip map's beginning", &m.bld.skip)

	start := &m.bld.start
	end := &m.bld.end

	imgui.TextUnformatted("Start time:")
	imgui.PushFont(Font16)
	imgui.SetNextItemWidth(-1)
	if sliderIntSlide("##Start time", &start.value, 0, end.ogValue-1, util.FormatSeconds(int(start.value)), imgui.SliderFlagsNoInput) {
		start.changed = start.value != start.ogValue
	}
	imgui.PopFont()

	if start.value >= end.value {
		end.value = start.value + 1
	}

	imgui.TextUnformatted("End time:")
	imgui.PushFont(Font16)
	imgui.SetNextItemWidth(-1)
	if sliderIntSlide("##End time", &end.value, 1, end.ogValue, util.FormatSeconds(int(end.value)), imgui.SliderFlagsNoInput) {
		end.changed = end.value != end.ogValue
	}
	imgui.PopFont()

	if start.value >= end.value {
		start.value = end.value - 1
	}

	m.drawStrainGraph()

	sliderIntReset("Audio offset", &m.bld.offset, -300, 300, "%dms")
}

func (m *timePopup) drawStrainGraph() {
	if m.timeCMap != m.bld.currentMap {
		m.sGraph = nil
		m.timeCMap = m.bld.currentMap
		m.graphStatus = "Loading graph..."

		goroutines.Run(func() {
			defer func() {
				if err := recover(); err != nil { //TODO: Technically should be fixed but unexpected parsing problem won't crash whole process
					m.graphStatus = "Failed to load"
				}
			}()

			beatmap.ParseTimingPointsAndPauses(m.timeCMap)
			beatmap.ParseObjects(m.timeCMap, true, false)

			m.peaks = performance.GetDifficultyCalculator().CalculateStrainPeaks(m.timeCMap.HitObjects, m.timeCMap.Diff)

			m.graphStatus = ""
		})
	}

	sHeight := 70
	sWidth := int(math32.Round(imgui.ContentRegionAvail().X - imgui.CurrentStyle().ItemInnerSpacing().X - 1))

	if m.graphStatus != "" {
		pad := (float32(sHeight) - imgui.TextLineHeightWithSpacing()) / 2

		dummyExactY(pad)

		centerTable("sgraphstatus", -1, func() {
			imgui.TextUnformatted(m.graphStatus)
		})

		dummyExactY(pad)
	} else {
		imgui.SetCursorPos(imgui.CursorPos().Add(vec2(imgui.CurrentStyle().ItemInnerSpacing().X/2, 0)))

		settings.Gameplay.StrainGraph.Width = (float64(sWidth) / settings.Graphics.GetWidthF()) * 768 * settings.Graphics.GetAspectRatio()
		settings.Gameplay.StrainGraph.Height = (float64(sHeight) / settings.Graphics.GetHeightF()) * 768

		redraw := false

		if m.fbo == nil || m.fbo.GetWidth() != sWidth || m.fbo.GetHeight() != sHeight {
			m.fbo = buffer.NewFrame(sWidth, sHeight, false, false)
			redraw = true
		}

		if m.sGraph == nil {
			m.sGraph = play.NewStrainGraph(m.timeCMap, m.peaks, true, false)
			redraw = true
		}

		if redraw || m.bld.start.value != m.lastStart || m.bld.end.value != m.lastEnd {
			m.lastStart = m.bld.start.value
			m.lastEnd = m.bld.end.value

			m.sGraph.SetTimes(float64(m.lastStart*1000), float64(m.lastEnd*1000))

			// keep viewport on screen scaling
			viewport.Push(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()))

			m.fbo.Bind()
			m.fbo.ClearColor(0, 0, 0, 0)

			m.fBatch.SetCamera(mgl32.Ortho2D(0, float32(768*settings.Graphics.GetAspectRatio()), 0, 768))

			m.fBatch.Begin()

			m.sGraph.Draw(m.fBatch, 1)

			m.fBatch.End()

			m.fbo.Unbind()

			viewport.Pop()
		}

		imgui.Image(imgui.TextureID{Data: uintptr(m.fbo.Texture().GetID())}, vec2(float32(sWidth), float32(sHeight)))
	}
}
