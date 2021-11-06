package main

import (
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"

	tui "github.com/charmbracelet/bubbletea"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/wieku/danser-go/app/beatmap"
	"github.com/wieku/danser-go/app/database"
	"github.com/wieku/danser-go/app/settings"
)

var (
	debugLog       *bool
	interactiveCmd = &cobra.Command{
		Use:   "interactive",
		Short: "danser interactive tui",
		Long:  "danser interactive tui.",
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			var isExecute bool
			defer func() {
				if err != nil {
					log.Fatalln(err)
				}
			}()

			if debugLog != nil && !*debugLog {
				log.Default().SetOutput(io.Discard)
			}

			if newSettings := settings.LoadSettings(*settingsVersion); newSettings {
				monitor := glfw.GetPrimaryMonitor()
				mWidth, mHeight := monitor.GetVideoMode().Width, monitor.GetVideoMode().Height
				settings.Graphics.SetDefaults(int64(mWidth), int64(mHeight))
				settings.Save()
			}

			if err = database.Init(); err != nil {
				return
			}
			if err = tui.NewProgram(newTuiModel(&isExecute, database.LoadBeatmaps(*noDbCheck))).Start(); err != nil {
				return
			}

			if isExecute {
				mainReal()
			}
		},
	}
)

func init() {
	debugLog = interactiveCmd.PersistentFlags().BoolP("debug", "L", false, "Enable debug mode")
	rootCmd.AddCommand(interactiveCmd)
}

//

type tuiData struct {
	vmin    int
	vselect int // +vmin representation of hovered item
	vmax    int

	vshow int // how many bm showed at the list
}

//
type modeType int

const (
	ModeExit      modeType = 1
	ModeSearch    modeType = 2
	ModeHasBmData modeType = 3
)

type bmTempIndex struct {
	index int
	*beatmap.BeatMap
}

type tuiModel struct {
	isExecute *bool
	bm        []*beatmap.BeatMap

	head    string
	headTmp string
	p       string

	searchBm []bmTempIndex

	serachValue string
	data        tuiData

	mode map[modeType]struct{}
}

func newTuiModel(isExecute *bool, bm []*beatmap.BeatMap) *tuiModel {
	return &tuiModel{
		bm:          bm,
		isExecute:   isExecute,
		head:        "",
		p:           "",
		serachValue: "",
		mode:        map[modeType]struct{}{},
	}
}
func (m *tuiModel) Init() tui.Cmd {
	m.headTmp = ""
	return func() tui.Msg {
		return tuiData{
			vmin:    0,
			vselect: 0,
			vmax:    0,

			vshow: 20,
		}
	}
}

func (m *tuiModel) Update(msg tui.Msg) (tui.Model, tui.Cmd) {
	switch msg := msg.(type) {
	case tui.KeyMsg:
		return m.handleKeyboardInput(msg)
	case tuiData:
		return m.handleTuiData(msg)
	}
	return m, nil
}

func (m *tuiModel) View() string {
	m.updateHeader()
	defer m.updateHeader()
	if m.hasMode(ModeExit) {
		return m.head
	}

	dthead := pterm.TableData{
		{"No", "Artist", "Creator", "Title", "Difficulty"},
	}
	dt := pterm.TableData{}
	m.searchBm = []bmTempIndex{}
	for i, bm := range m.bm {
		var s = []string{
			strconv.Itoa(i),
			bm.Artist,
			bm.Creator,
			bm.Name,
			bm.Difficulty,
		}
		if m.hasMode(ModeSearch) {
			fts := strings.ToLower(strings.Join(s, " "))
			if !strings.Contains(fts, strings.ToLower(m.serachValue)) && m.serachValue != "" {
				continue
			}
		}
		m.searchBm = append(m.searchBm, bmTempIndex{
			index:   i,
			BeatMap: bm,
		})
		dt = append(dt, s)
	}

	if m.data.vselect < m.data.vmin {
		m.data.vmin--
	}
	if m.data.vselect > m.data.vmin+m.data.vmax {
		m.data.vmin++
	}
	for i, row := range dt {
		if i == m.data.vselect {
			for i, v := range row {
				row[i] = pterm.Yellow(v)
			}
		}
	}

	cap := m.data.vshow
	if l := len(dt); l < m.data.vshow {
		cap = l
	}
	m.data.vmax = cap
	if k := cap + m.data.vmin; k < len(dt) {
		cap = k
	} else {
		cap = len(dt)
	}
	lcap := m.data.vmin
	if m.data.vmin < 0 {
		lcap = 0
	}
	if m.data.vmin > len(dt) {
		lcap = len(dt)
	}

	dtview := dt[lcap:cap]
	if len(dtview) < 1 {
		m.headTmp = m.headTmp + pterm.Info.Sprintf("Nothing found :/\n")
	}

	// render bm list
	var s string
	var err error
	if s, err = pterm.DefaultTable.
		WithHasHeader().
		WithData(append(dthead, dtview...)).
		Srender(); err != nil {
		s = pterm.Error.Sprintf("table render error\n")
	}

	return m.head + s + "\n" + m.p
}

//

func (m *tuiModel) updateHeader() {
	m.head = m.gradText("danser interactive") + "\n"
	if m.hasMode(ModeExit) {
		m.head = m.head + pterm.Info.Sprintf("Exiting\n")
		return
	}

	if m.hasMode(ModeHasBmData) {
		m.head = m.head + pterm.Info.Sprintf(m.gradText(pterm.Sprintf("Beatmap loaded, %d entry. ctrl+f to find, enter to load. [%d/%d %d]", len(m.bm),
			m.data.vselect, len(m.searchBm),
			m.data.vshow))) + "\n"
	}

	if m.hasMode(ModeSearch) {
		m.head = m.head + pterm.Info.Sprintf(m.gradText(pterm.Sprintf("Search: %s", m.serachValue))) + "\n"
	}

	if m.headTmp != "" {
		m.head = m.head + m.headTmp
		m.headTmp = ""
	}
}

func (m *tuiModel) handleTuiData(msg tuiData) (tui.Model, tui.Cmd) {
	m.addMode(ModeHasBmData)
	m.data = msg
	return m, nil
}

func (m *tuiModel) handleKeyboardInput(msg tui.KeyMsg) (tui.Model, tui.Cmd) {
	// Unicode 《CJK》 https://github.com/charmbracelet/bubbletea/issues/153
	switch ms := msg.String(); ms {
	case "ctrl+c":
		m.addMode(ModeExit)
		return m, tui.Quit
	case "ctrl+f":
		m.addMode(ModeSearch)
	case "esc":
		if m.hasMode(ModeSearch) {
			m.headTmp = m.headTmp + pterm.Info.Sprintf("Search canceled.\n")
			m.serachValue = ""
			m.delMode(ModeSearch)
		}
	case "backspace":
		if m.hasMode(ModeSearch) && len(m.serachValue) > 0 {
			m.serachValue = m.serachValue[:len(m.serachValue)-1]
		}
	case "enter":
		if m.hasMode(ModeSearch) {
			m.headTmp = m.headTmp + pterm.Info.Sprintf("You searched %q\n", m.serachValue)
			m.serachValue = ""
			m.delMode(ModeSearch)
		}
		if len(m.searchBm) > 0 {
			var choosed = m.searchBm[m.data.vselect]
			m.headTmp = m.headTmp + pterm.Info.Sprintf("Starting [%s (%s)]\n", choosed.Name, choosed.Difficulty)
			*artist = choosed.Artist
			*title = choosed.Name
			*difficulty = choosed.Difficulty

			*m.isExecute = true
			return m, tui.Quit
		} else {
			m.headTmp = m.headTmp + pterm.Error.Sprintf("Empty selection. try again.\n")
		}
	case "up":
		// m.headTmp = m.headTmp + pterm.Info.Sprintf("Up vmin:%d vmax:%d vselect:%d vshow:%d\n", m.data.vmin, m.data.vmax, m.data.vselect, m.data.vshow)
		if m.data.vselect > 0 {
			m.data.vselect = m.data.vselect - 1
		}
	case "down":
		// m.headTmp = m.headTmp + pterm.Info.Sprintf("Down vmin:%d vmax:%d vselect:%d vshow:%d\n", m.data.vmin, m.data.vmax, m.data.vselect, m.data.vshow)
		if m.data.vselect < len(m.bm) {
			m.data.vselect = m.data.vselect + 1
		}
	default:
		if m.hasMode(ModeSearch) {
			if len(msg.Runes) > 0 { // if not mone
				m.serachValue = m.serachValue + string(msg.Runes)
				m.data.vselect = 0
			}
		}
	}
	m.p = fmt.Sprintf("\n:%s", msg)
	return m, nil
}

//
func (m *tuiModel) addMode(mr modeType) {
	if m.hasMode(mr) {
		return
	}
	m.mode[mr] = struct{}{}
}
func (m *tuiModel) delMode(mr modeType) {
	delete(m.mode, mr)
}
func (m *tuiModel) hasMode(mr modeType) bool {
	_, ok := m.mode[mr]
	return ok
}

//

func (m *tuiModel) gradText(text string) (ret string) {
	from := pterm.NewRGB(0, 255, 255)
	to := pterm.NewRGB(255, 0, 255)
	ss := strings.Split(text, "")
	for i := 0; i < len(text); i++ {
		ret += from.Fade(0, float32(len(text)), float32(i), to).Sprint(ss[i])
	}
	return
}
