package launcher

import (
	"fmt"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/wieku/danser-go/app/beatmap"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/bass"
	"github.com/wieku/danser-go/framework/graphics/texture"
	"github.com/wieku/danser-go/framework/math/animation"
	"github.com/wieku/danser-go/framework/math/mutils"
	"github.com/wieku/danser-go/framework/platform"
	"github.com/wieku/danser-go/framework/qpc"
	"github.com/wieku/danser-go/framework/util"
	"golang.org/x/exp/slices"
	"math"
	"math/rand"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
)

type SortBy int

const (
	Title = SortBy(iota)
	Artist
	Creator
	DateAdded
	Difficulty
)

var sortMethods = []SortBy{Title, Artist, Creator, DateAdded, Difficulty}

func (s SortBy) String() string {
	switch s {
	case Title:
		return "Title"
	case Artist:
		return "Artist"
	case Creator:
		return "Creator"
	case DateAdded:
		return "Date Added"
	case Difficulty:
		return "Difficulty"
	}

	return ""
}

type mapWithName struct {
	name string
	bMap *beatmap.BeatMap
}

func newMapWithName(bMap *beatmap.BeatMap) *mapWithName {
	return &mapWithName{
		name: strings.ToLower(fmt.Sprintf("%s - %s [%s] by %s %d %d", bMap.Artist, bMap.Name, bMap.Difficulty, bMap.Creator, bMap.SetID, bMap.ID)),
		bMap: bMap,
	}
}

type beatmapSet struct {
	bounds  imgui.Vec2
	bMaps   []*beatmap.BeatMap
	hovered bool
}

type maps []*mapWithName

func (e maps) String(i int) string {
	return e[i].name
}

func (e maps) Len() int {
	return len(e)
}

type songSelectPopup struct {
	*popup

	bld      *builder
	beatmaps maps

	searchResults  []*beatmapSet
	sizeCalculated int

	searchStr string

	prevMap       *beatmap.BeatMap
	PreviewedSong *bass.TrackBass
	volume        *animation.Glider
	stopTime      float64
	thumbTex      *texture.TextureSingle
	lastThumbPath string
	drawTex       bool
	lastScrollY   float32

	preIndex, postIndex int
	focusTheMap         bool

	comboOpened bool
}

func newSongSelectPopup(bld *builder, beatmaps []*beatmap.BeatMap) *songSelectPopup {
	mP := &songSelectPopup{
		popup:    newPopup("Song select", popBig),
		bld:      bld,
		beatmaps: make([]*mapWithName, 0),
		volume:   animation.NewGlider(0),
	}

	mP.internalDraw = mP.drawSongSelect

	for _, bMap := range beatmaps {
		mP.beatmaps = append(mP.beatmaps, newMapWithName(bMap))
	}

	mP.search()

	return mP
}

func (m *songSelectPopup) update() {
	cT := qpc.GetMilliTimeF()

	m.volume.Update(cT)
	if m.PreviewedSong != nil {
		m.PreviewedSong.SetVolumeRelative(m.volume.GetValue() * launcherConfig.PreviewVolume)

		if cT >= m.stopTime {
			m.stopPreview()
		}
	}
}

func (m *songSelectPopup) drawSongSelect() {
	imgui.PushFont(Font32)

	imgui.SetNextItemWidth(-1)
	if searchBox("##searchpath", &m.searchStr) {
		m.search()
		m.focusTheMap = true
	}

	if !m.comboOpened && !imgui.IsAnyItemActive() && !imgui.IsMouseClicked(0) {
		imgui.SetKeyboardFocusHereV(-1)
	}

	imgui.PopFont()

	imgui.PushFont(Font20)

	if imgui.BeginTableV("sortrandom", 2, 0, vec2(-1, 0), -1) {
		imgui.TableSetupColumnV("##sortrandom1", imgui.TableColumnFlagsWidthStretch, 0, uint(0))
		imgui.TableSetupColumnV("##sortrandom2", imgui.TableColumnFlagsWidthFixed, 0, uint(1))

		imgui.TableNextColumn()

		imgui.AlignTextToFramePadding()
		imgui.Text("Sort by:")

		imgui.SameLine()

		m.comboOpened = false

		imgui.SetNextItemWidth(150)

		if imgui.BeginCombo("##sortcombo", launcherConfig.SortMapsBy.String()) {
			m.comboOpened = true

			for _, s := range sortMethods {
				if imgui.SelectableV(s.String(), s == launcherConfig.SortMapsBy, 0, vzero()) && s != launcherConfig.SortMapsBy {
					launcherConfig.SortMapsBy = s
					m.search()
					m.focusTheMap = true
					saveLauncherConfig()
				}
			}

			imgui.EndCombo()
		}

		imgui.SameLine()

		ImIO.SetFontGlobalScale(20.0 / 32)
		imgui.PushFont(FontAw)

		sDir := "\uF882"
		if launcherConfig.SortAscending {
			sDir = "\uF15D"
		}

		if imgui.Button(sDir) {
			launcherConfig.SortAscending = !launcherConfig.SortAscending
			m.search()
			m.focusTheMap = true
			saveLauncherConfig()
		}

		ImIO.SetFontGlobalScale(1)
		imgui.PopFont()

		imgui.TableNextColumn()

		if imgui.Button("Random") {
			m.selectRandom()
		}

		imgui.EndTable()
	}

	imgui.PopFont()

	csPos := imgui.CursorScreenPos()

	imgui.BeginChild("##bsets")

	if m.sizeCalculated > 1 { // we need at least 2 passes to have correct metrics
		m.sizeCalculated = 2

		sc1 := imgui.ScrollY()

		if sc1 != m.lastScrollY {
			sc2 := sc1 + imgui.ContentRegionAvail().Y

			preIndex, postIndex := 0, len(m.searchResults)-1

			for i, b := range m.searchResults {
				if b.bounds.Y < sc1 {
					preIndex = i + 1
				}

				if sc2 < b.bounds.X {
					if postIndex > i-1 {
						postIndex = i - 1
					}
				}
			}

			m.preIndex = preIndex
			m.postIndex = postIndex

			m.lastScrollY = sc1
		}
	}

	imgui.PushStyleVarVec2(imgui.StyleVarFramePadding, vec2(5, 0))

	if imgui.BeginTableV("bsetstab", 1, imgui.TableFlagsRowBg|imgui.TableFlagsPadOuterX|imgui.TableFlagsBordersH, vec2(-1, 0), -1) {
		imgui.TableSetBgColor(imgui.TableBgTargetRowBg1, vec4(0.5, 0.5, 0.5, 1))

		for i := 0; i < len(m.searchResults); i++ {
			b := m.searchResults[i]

			imgui.TableNextColumn()

			if m.focusTheMap && m.sizeCalculated > 1 {
				if m.bld.currentMap != nil && m.bld.currentMap.Dir == b.bMaps[0].Dir {
					imgui.SetScrollY(b.bounds.X)
				}
			}

			if i < m.preIndex || i > m.postIndex {
				imgui.SetCursorPos(vec2(imgui.CursorPos().X, b.bounds.Y))
				continue
			}

			isPreviewed := false

			for _, bMap := range b.bMaps {
				if bMap == m.prevMap {
					isPreviewed = true
					break
				}
			}

			c1 := imgui.CursorPos().Y

			rId := strconv.Itoa(i)

			imgui.BeginGroup()

			if imgui.BeginTableV("bsetstab"+rId, 2, imgui.TableFlagsSizingStretchProp, vec2(-1, 0), -1) {
				imgui.PushFont(Font32)

				imgui.TableSetupColumnV("##hhh"+rId, imgui.TableColumnFlagsWidthStretch, 0, uint(0))
				imgui.TableSetupColumnV("##hhhg"+rId, imgui.TableColumnFlagsWidthFixed, imgui.FrameHeight()*2+imgui.CurrentStyle().ItemSpacing().X, uint(1))

				imgui.TableNextColumn()

				imgui.PushTextWrapPos()

				imgui.Text(b.bMaps[0].Name)

				imgui.PopTextWrapPos()

				imgui.PopFont()

				imgui.TableNextColumn()

				if b.hovered {
					imgui.PushFont(Font20)

					imgui.PushStyleVarFloat(imgui.StyleVarFrameBorderSize, 0)
					imgui.PushStyleColor(imgui.StyleColorButton, vec4(0, 0, 0, 1))
					imgui.PushStyleColor(imgui.StyleColorButtonActive, vec4(0.2, 0.2, 0.2, 1))
					imgui.PushStyleColor(imgui.StyleColorButtonHovered, vec4(0.4, 0.4, 0.4, 1))

					s := b.bMaps[0].SetID == 0

					if s {
						imgui.PushItemFlag(imgui.ItemFlagsDisabled, true)
					}

					ImIO.SetFontGlobalScale(16.0 / 32)
					imgui.PushFont(FontAw)

					imgui.AlignTextToFramePadding()
					if imgui.ButtonV("\uF7A2##"+rId, vec2(imgui.FrameHeight()*2, imgui.FrameHeight()*2)) {
						platform.OpenURL(fmt.Sprintf("https://osu.ppy.sh/s/%d", b.bMaps[0].SetID))
					}

					ImIO.SetFontGlobalScale(1)
					imgui.PopFont()

					if imgui.IsItemHoveredV(imgui.HoveredFlagsAllowWhenDisabled) {
						imgui.BeginTooltip()

						if s {
							imgui.Text("Not available")
						} else {
							imgui.Text(fmt.Sprintf("https://osu.ppy.sh/s/%d", b.bMaps[0].SetID))
						}

						imgui.EndTooltip()
					}

					if s {
						imgui.PopItemFlag()
					}

					imgui.SameLine()

					name := "\uF04B"
					if isPreviewed {
						name = "\uF04D"
					}

					ImIO.SetFontGlobalScale(16.0 / 32)
					imgui.PushFont(FontAw)

					imgui.AlignTextToFramePadding()
					if imgui.ButtonV(name+"##"+rId, vec2(imgui.FrameHeight()*2, imgui.FrameHeight()*2)) {
						m.stopPreview()

						if name == "\uF04B" {
							m.startPreview(b.bMaps[0])
						}
					}

					ImIO.SetFontGlobalScale(1)
					imgui.PopFont()

					if imgui.IsItemHoveredV(imgui.HoveredFlagsAllowWhenDisabled) {
						imgui.BeginTooltip()

						if isPreviewed {
							imgui.Text("Stop preview")
						} else {
							imgui.Text("Play preview")
						}

						imgui.EndTooltip()
					}

					imgui.PopStyleVar()
					imgui.PopStyleColor()
					imgui.PopStyleColor()
					imgui.PopStyleColor()

					imgui.PopFont()
				}

				imgui.EndTable()
			}

			imgui.Text(fmt.Sprintf("%s // %s", b.bMaps[0].Artist, b.bMaps[0].Creator))

			imgui.PushFont(Font20)

			for j, bMap := range b.bMaps {
				fDiffName := ">   " + bMap.Difficulty

				tSiz := imgui.CalcTextSize(fDiffName, false, 0)

				sPos := imgui.CursorScreenPos()

				if imgui.SelectableV(fDiffName+"##"+rId+"s"+strconv.Itoa(j), bMap == m.bld.currentMap, 0, vzero()) {
					m.bld.setMap(bMap)

					if !isPreviewed && launcherConfig.PreviewSelected {
						m.stopPreview()
						m.startPreview(bMap)
					}

					m.opened = false
				}

				if imgui.IsItemHovered() && ImIO.MousePosition().X <= sPos.X+tSiz.X {
					imgui.PushFont(Font24)

					const tgAsp = float32(4.0 / 3)

					imgui.BeginTooltip()

					cPos := imgui.CursorPos()

					thumbPath := filepath.Join(settings.General.GetSongsDir(), bMap.Dir, bMap.Bg)

					if m.lastThumbPath != thumbPath {
						if m.thumbTex != nil {
							m.thumbTex.Dispose()
							m.thumbTex = nil
						}

						pX, err := texture.NewPixmapFileString(thumbPath)
						if err == nil {
							m.thumbTex = texture.LoadTextureSingle(pX.RGBA(), 4)

							pX.Dispose()
						}

						m.lastThumbPath = thumbPath
						m.drawTex = false
					}

					if m.thumbTex != nil {
						uvTL := vec2(0, 0)
						uvBR := vec2(1, 1)

						asp := float32(m.thumbTex.GetWidth()) / float32(m.thumbTex.GetHeight())

						if asp > tgAsp {
							uvTL.X = (1 - tgAsp/asp) / 2
							uvBR.X = 1 - uvTL.X
						} else {
							uvTL.Y = (1 - asp/tgAsp) / 2
							uvBR.Y = 1 - uvTL.X
						}

						imgui.ImageV(imgui.TextureID(m.thumbTex.GetID()), vec2(200*tgAsp, 200), uvTL, uvBR, imgui.Vec4{X: 1, Y: 1, Z: 1, W: 0.3}, imgui.Vec4{})
					}

					imgui.SetCursorPos(cPos)

					sR := "N/A"
					if bMap.Stars >= 0 {
						sR = mutils.FormatWOZeros(bMap.Stars, 2)
					}

					bpm := fmt.Sprintf("%.0f", bMap.MinBPM)
					if math.Abs(bMap.MinBPM-bMap.MaxBPM) > 0.01 {
						bpm = fmt.Sprintf("%.0f - %.0f", bMap.MinBPM, bMap.MaxBPM)
					}

					if imgui.BeginTableV("btooltip", 4, imgui.TableFlagsSizingStretchProp|imgui.TableFlagsNoClip, vec2(200.0*tgAsp, 0), -1) {
						imgui.TableSetupColumnV("btooltip1", imgui.TableColumnFlagsWidthFixed, 0, uint(0))
						imgui.TableSetupColumnV("btooltip2", imgui.TableColumnFlagsWidthStretch, 0, uint(1))
						imgui.TableSetupColumnV("btooltip3", imgui.TableColumnFlagsWidthFixed, 0, uint(0))
						imgui.TableSetupColumnV("btooltip4", imgui.TableColumnFlagsWidthFixed, imgui.CalcTextSize("9.9", false, 0).X, uint(1))

						tRow := func(text string, text2 string, args ...any) {
							imgui.TableNextColumn()
							imgui.Text(text)

							imgui.TableNextColumn()
							imgui.Textf(text2, args...)
						}

						tRow("Stars: ", sR)
						tRow("", "")

						tRow("Objects: ", "%d", bMap.Circles+bMap.Sliders+bMap.Spinners)
						tRow("AR: ", mutils.FormatWOZeros(bMap.Diff.GetAR(), 2))

						tRow("Circles: ", "%d", bMap.Circles)
						tRow("OD: ", mutils.FormatWOZeros(bMap.Diff.GetOD(), 2))

						tRow("Sliders: ", "%d", bMap.Sliders)
						tRow("CS: ", mutils.FormatWOZeros(bMap.Diff.GetCS(), 2))

						tRow("Spinners: ", "%d", bMap.Spinners)
						tRow("HP: ", mutils.FormatWOZeros(bMap.Diff.GetHP(), 2))

						tRow("BPM: ", bpm)
						tRow("", "")

						tRow("Length: ", util.FormatSeconds(bMap.Length/1000))
						tRow("", "")

						imgui.EndTable()
					}

					imgui.PopFont()
					imgui.EndTooltip()
				}
			}

			imgui.PopFont()

			imgui.SetCursorPos(imgui.CursorPos().Plus(vec2(imgui.ContentRegionAvail().X, 0))) //hack to get cell hovering to work

			imgui.EndGroup()

			b.hovered = imgui.IsItemHoveredV(imgui.HoveredFlagsAllowWhenBlockedByActiveItem | imgui.HoveredFlagsAllowWhenOverlapped)

			c2 := imgui.CursorPos().Y

			b.bounds = vec2(c1, c2)
		}

		if m.sizeCalculated > 1 {
			m.focusTheMap = false
		}

		imgui.EndTable()
	}

	imgui.PopStyleVar()

	m.sizeCalculated++

	imgui.EndChild()

	imgui.WindowDrawList().AddLine(csPos, csPos.Plus(vec2(imgui.ContentRegionAvail().X, 0)), imgui.PackedColorFromVec4(imgui.CurrentStyle().Color(imgui.StyleColorSeparator)))
}

func (m *songSelectPopup) selectRandom() {
	if len(m.searchResults) == 0 {
		return
	}

	i := rand.Intn(len(m.searchResults))

	bMap := m.searchResults[i].bMaps[len(m.searchResults[i].bMaps)-1]

	m.bld.setMap(bMap)
	m.focusTheMap = true

	if launcherConfig.PreviewSelected {
		m.stopPreview()
		m.startPreview(bMap)
	}
}

func (m *songSelectPopup) stopPreview() {
	if m.PreviewedSong != nil {
		m.PreviewedSong.Stop()
		m.PreviewedSong = nil
		m.prevMap = nil
	}
}

func (m *songSelectPopup) startPreview(bMap *beatmap.BeatMap) {
	cT := qpc.GetMilliTimeF()

	track := bass.NewTrack(filepath.Join(settings.General.OsuSongsDir, bMap.Dir, bMap.Audio))

	if track != nil {
		beatmap.ParseTimingPointsAndPauses(bMap)

		prevTime := float64(bMap.PreviewTime)
		if prevTime < 0 {
			prevTime = float64(bMap.Length) * 0.4
		}

		track.SetPosition(prevTime / 1000)
		track.PlayV(0)
		m.PreviewedSong = track

		m.volume.Reset()
		m.volume.AddEventS(cT, cT+1000, 0, 1)
		m.volume.AddEventS(cT+9000, cT+10000, 1, 0)
		m.stopTime = cT + 10000
		m.prevMap = bMap
	}
}

func (m *songSelectPopup) search() {
	m.lastScrollY = float32(-1.0)
	m.sizeCalculated = 0
	m.searchResults = m.searchResults[:0]

	sString := strings.ToLower(m.searchStr)

	foundMaps := make([]*beatmap.BeatMap, 0, len(m.beatmaps))

	for _, b := range m.beatmaps {
		if sString != "" && !strings.Contains(b.name, sString) {
			continue
		}

		foundMaps = append(foundMaps, b.bMap)
	}

	sortMaps(foundMaps, launcherConfig.SortMapsBy)

	for _, b := range foundMaps {
		if len(m.searchResults) == 0 || m.searchResults[len(m.searchResults)-1].bMaps[0].Dir != b.Dir {
			m.searchResults = append(m.searchResults, &beatmapSet{bMaps: make([]*beatmap.BeatMap, 0, 1)})
		}

		m.searchResults[len(m.searchResults)-1].bMaps = append(m.searchResults[len(m.searchResults)-1].bMaps, b)
	}

	m.preIndex = 0
	m.postIndex = len(m.searchResults) - 1
}

func (m *songSelectPopup) open() {
	m.focusTheMap = true

	m.popup.open()
}

func compareStrings(l, r string) int {
	lRa := []rune(l)
	rRa := []rune(r)
	lenM := mutils.Min(len(lRa), len(rRa))

	for i := 0; i < lenM; i++ {
		cL := unicode.ToLower(lRa[i])
		cR := unicode.ToLower(rRa[i])
		if cL < cR {
			return -1
		} else if cL > cR {
			return 1
		}
	}

	if len(lRa) == len(rRa) {
		return 0
	} else if len(lRa) > len(rRa) {
		return 1
	}

	return -1
}

func sortMaps(bMaps []*beatmap.BeatMap, sortBy SortBy) {
	slices.SortStableFunc(bMaps, func(b1, b2 *beatmap.BeatMap) bool {
		var res int

		switch sortBy {
		case Title:
			res = compareStrings(b1.Name, b2.Name)
		case Artist:
			res = compareStrings(b1.Artist, b2.Artist)
		case Creator:
			res = compareStrings(b1.Creator, b2.Creator)
		case DateAdded:
			if compareStrings(b1.Dir, b2.Dir) != 0 || mutils.Abs(b1.LastModified/1000-b2.LastModified/1000) > 10 {
				res = mutils.Compare(b1.LastModified/1000, b2.LastModified/1000)
			} else {
				res = 0
			}
		case Difficulty:
			res = mutils.Compare(b1.Stars, b2.Stars)
		}

		if !launcherConfig.SortAscending {
			res = -res
		}

		if res != 0 {
			return res < 0
		}

		res = compareStrings(b1.Dir, b2.Dir)

		if !launcherConfig.SortAscending {
			res = -res
		}

		if res != 0 {
			return res < 0
		}

		return mutils.Compare(b1.Stars, b2.Stars) < 1 // Don't flip grouped difficulties
	})
}
