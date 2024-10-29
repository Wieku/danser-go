package launcher

import (
	"cmp"
	"fmt"
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/sqweek/dialog"
	"github.com/wieku/danser-go/app/osuapi"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/env"
	"github.com/wieku/danser-go/framework/goroutines"
	"github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/math/mutils"
	"github.com/wieku/danser-go/framework/platform"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const padY = 30

var nameSplitter = regexp.MustCompile(`[A-Z]+[^A-Z]*`)

type settingsEditor struct {
	*popup

	searchCache  map[string]int
	scrollTo     string
	blockSearch  bool
	searchString string

	current  *settings.Config
	combined *settings.CombinedConfig

	listenerCalled bool
	sectionCache   map[string]imgui.Vec2

	active      string
	lastActive  string
	pwShowHide  map[string]bool
	comboSearch map[string]string

	keyChange       string
	keyChangeVal    reflect.Value
	keyChangeOpened bool
	danserRunning   bool

	saveListener func()

	scrollCache map[string]bool
}

func newSettingsEditor(config *settings.Config) *settingsEditor {
	editor := &settingsEditor{
		popup:        newPopup("Settings Editor", popBig),
		searchCache:  make(map[string]int),
		sectionCache: make(map[string]imgui.Vec2),
		pwShowHide:   make(map[string]bool),
		comboSearch:  make(map[string]string),
		scrollCache:  make(map[string]bool),
	}

	editor.internalDraw = editor.drawEditor

	editor.current = config
	editor.combined = config.GetCombined()

	editor.search()

	return editor
}

func (editor *settingsEditor) updateKey(_ *glfw.Window, key glfw.Key, scancode int, action glfw.Action, _ glfw.ModifierKey) {
	if editor.opened && editor.keyChange != "" && action == glfw.Press {
		keyText, ok := platform.GetKeyName(key, scancode)

		if ok && keyText != "" {
			editor.keyChangeVal.SetString(keyText)
			editor.keyChangeOpened = false
			editor.keyChange = ""

			imgui.SetWindowFocusStr("##Settings Editor")
		}
	}
}

func (editor *settingsEditor) setDanserRunning(running bool) {
	editor.danserRunning = running
}

func (editor *settingsEditor) setSaveListener(saveListener func()) {
	editor.saveListener = saveListener
}

func (editor *settingsEditor) drawEditor() {
	imgui.PushItemFlag(imgui.ItemFlags(imgui.ItemFlagsDisabled), false)

	settings.General.OsuSkinsDir = editor.combined.General.OsuSkinsDir

	imgui.PushStyleColorVec4(imgui.ColWindowBg, vec4(0, 0, 0, .9))
	imgui.PushStyleColorVec4(imgui.ColFrameBg, vec4(.2, .2, .2, 1))

	currentRunning := editor.danserRunning

	imgui.PushFont(Font20)

	height := imgui.ContentRegionAvail().Y
	if currentRunning {
		height -= imgui.FrameHeightWithSpacing() + imgui.CurrentStyle().ItemSpacing().Y
	}

	imgui.PopFont()

	navScrolling := false

	if imgui.BeginChildStrV("##EditorUp", vec2(-1, height), imgui.ChildFlagsNone, 0) {
		imgui.PushStyleVarVec2(imgui.StyleVarCellPadding, vec2(2, 0))
		if imgui.BeginTableV("Edit main table", 2, imgui.TableFlagsSizingStretchProp, vec2(-1, -1), -1) {
			imgui.PopStyleVar()

			imgui.TableSetupColumnV("Edit main table 1", imgui.TableColumnFlagsWidthFixed, 0, imgui.ID(0))
			imgui.TableSetupColumnV("Edit main table 2", imgui.TableColumnFlagsWidthStretch, 0, imgui.ID(1))

			imgui.TableNextColumn()

			imgui.PushStyleColorVec4(imgui.ColChildBg, vec4(0, 0, 0, .5))

			imgui.PushFont(FontAw)
			{

				imgui.PushStyleVarFloat(imgui.StyleVarScrollbarSize, 9)

				if imgui.BeginChildStrV("##Editor navigation", vec2(imgui.FontSize()*1.5+9, -1), imgui.ChildFlagsNone, imgui.WindowFlagsAlwaysVerticalScrollbar) {
					navScrolling = handleDragScroll()
					editor.scrollTo = ""

					imgui.PushStyleVarFloat(imgui.StyleVarFrameRounding, 0)
					imgui.PushStyleVarFloat(imgui.StyleVarFrameBorderSize, 0)
					imgui.PushStyleVarVec2(imgui.StyleVarItemSpacing, vzero())

					editor.buildNavigationFor(editor.combined)

					imgui.PopStyleVar()
					imgui.PopStyleVar()
					imgui.PopStyleVar()
				}

				imgui.PopStyleVar()

				imgui.EndChild()
			}
			imgui.PopFont()

			imgui.PopStyleColor()

			imgui.TableNextColumn()

			imgui.PushFont(Font32)
			{
				imgui.SetNextItemWidth(-1)

				if searchBox("##Editor search", &editor.searchString) {
					editor.search()
				}

				if !navScrolling && !editor.blockSearch && !imgui.IsAnyItemActive() && !imgui.IsMouseClickedBool(0) {
					imgui.SetKeyboardFocusHereV(-1)
				}
			}
			imgui.PopFont()

			imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, vec2(5, 0))

			if imgui.BeginChildStrV("##Editor main", vec2(-1, -1), imgui.ChildFlagsNone, imgui.WindowFlags(imgui.ChildFlagsAlwaysUseWindowPadding)) {
				imgui.PopStyleVar()

				editor.blockSearch = handleDragScroll()

				imgui.PushFont(Font20)

				editor.drawSettings()

				imgui.PopFont()
			} else {
				imgui.PopStyleVar()
			}

			imgui.EndChild()

			imgui.EndTable()
		} else {
			imgui.PopStyleVar()
		}
	}

	imgui.EndChild()

	imgui.PushFont(Font20)

	if currentRunning {
		centerTable("tabdanser is running", -1, func() {
			imgui.AlignTextToFramePadding()
			imgui.TextUnformatted("Danser is running! Click")
			imgui.SameLine()
			if imgui.Button("Apply##drunning") {
				if editor.saveListener != nil {
					editor.saveListener()
				}
			}
			imgui.SameLine()
			imgui.TextUnformatted("to see changes.")
		})
	}

	imgui.PopFont()

	imgui.PopStyleColor()
	imgui.PopStyleColor()

	imgui.PopItemFlag()
}

func (editor *settingsEditor) search() {
	editor.sectionCache = make(map[string]imgui.Vec2)
	editor.searchCache = make(map[string]int)
	editor.buildSearchCache("Main", reflect.ValueOf(editor.combined), editor.searchString, false)
}

func (editor *settingsEditor) buildSearchCache(path string, u reflect.Value, search string, omitSearch bool) bool {
	typ := u.Elem()
	def := u.Type().Elem()

	count := typ.NumField()

	found := false

	skipMap := make(map[string]uint8)
	consumed := make(map[string]uint8)

	for i := 0; i < count; i++ {
		field := typ.Field(i)
		dF := def.Field(i)

		if def.Field(i).Tag.Get("skip") != "" {
			continue
		}

		if editor.shouldBeHidden(consumed, skipMap, typ, dF) {
			continue
		}

		label := editor.getLabel(dF)

		sPath := path + "." + label

		match := omitSearch || strings.Contains(strings.ToLower(label), search)

		if field.Type().Kind() == reflect.Ptr && (field.CanInterface() || def.Field(i).Anonymous) && !field.IsNil() && !field.Type().AssignableTo(reflect.TypeOf(&settings.HSV{})) {
			sub := editor.buildSearchCache(sPath, field, search, match)
			match = match || sub
		} else if field.Type().Kind() == reflect.Slice && field.CanInterface() {
			for j := 0; j < field.Len(); j++ {
				sub := editor.buildSearchCache(sPath, field.Index(j), search, match)
				match = match || sub
			}
		}

		if match {
			editor.searchCache[sPath] = 1
			found = true
		}
	}

	return found
}

func (editor *settingsEditor) buildNavigationFor(u interface{}) {
	typ := reflect.ValueOf(u).Elem()
	def := reflect.TypeOf(u).Elem()

	count := typ.NumField()

	imgui.PushStyleColorVec4(imgui.ColButton, vec4(0, 0, 0, 0))

	buttonSize := imgui.FontSize() * 1.5

	cAvail := imgui.ContentRegionAvail().Y
	sc1 := imgui.ScrollY()
	sc2 := sc1 + cAvail

	for i := 0; i < count; i++ {
		label := editor.getLabel(def.Field(i))

		if editor.searchCache["Main."+label] > 0 && (typ.Field(i).CanInterface() && !typ.Field(i).IsNil()) {
			if editor.active == label {
				cColor := *imgui.StyleColorVec4(imgui.ColCheckMark)

				imgui.PushStyleColorVec4(imgui.ColButton, vec4(0.2, 0.2, 0.2, 0.6))
				imgui.PushStyleColorVec4(imgui.ColText, vec4(cColor.X*1.2, cColor.Y*1.2, cColor.Z*1.2, 1))
			}

			c1 := imgui.CursorPos().Y

			if imgui.ButtonV(def.Field(i).Tag.Get("icon"), vec2(buttonSize, buttonSize)) {
				editor.scrollTo = label
			}

			c2 := imgui.CursorPos().Y

			if editor.active == label {
				if editor.lastActive != editor.active {
					if c2 > sc2 {
						imgui.SetScrollYFloat(c2 - cAvail)
					}

					if c1 < sc1 {
						imgui.SetScrollYFloat(c1)
					}

					editor.lastActive = editor.active
				}

				imgui.PopStyleColor()
				imgui.PopStyleColor()
			}

			if imgui.IsItemHovered() {
				imgui.PushFont(Font24)
				imgui.BeginTooltip()
				setTooltip(label)
				imgui.EndTooltip()
				imgui.PopFont()
			}
		}
	}

	imgui.PopStyleColor()
}

func (editor *settingsEditor) drawSettings() {
	rVal := reflect.ValueOf(editor.combined)

	typ := rVal.Elem()
	def := rVal.Type().Elem()

	count := typ.NumField()

	sc1 := imgui.ScrollY()
	sc2 := sc1 + imgui.ContentRegionAvail().Y

	forceDrawNew := false

	for i, j := 0, 0; i < count; i++ {
		field := typ.Field(i)
		dF := def.Field(i)

		lbl := editor.getLabel(dF)

		if editor.searchCache["Main."+lbl] == 0 {
			continue
		}

		if field.CanInterface() && field.Type().Kind() == reflect.Ptr && !field.IsNil() {
			if j > 0 {
				imgui.Dummy(vec2(1, 2*padY))
			}

			drawNew := true
			if v, ok := editor.sectionCache["Main."+lbl]; ok {
				if editor.scrollTo == lbl {
					imgui.SetScrollYFloat(v.X)
				}

				if (sc1 > v.Y || sc2 < v.X) && !forceDrawNew {
					drawNew = false
					dummyExactY(v.Y - v.X)
				}
			}

			if drawNew {
				iSc1 := imgui.CursorPos().Y

				editor.buildMainSection("##"+dF.Name, "Main."+lbl, lbl, field, dF)

				iSc2 := imgui.CursorPos().Y

				cCacheVal := editor.sectionCache["Main."+lbl]

				if math32.Abs(cCacheVal.X-iSc1) > 0.001 || math32.Abs(cCacheVal.Y-iSc2) > 0.001 { // if size of the section changed (dynamically hidden items/array changes) we need to redraw stuff below to have good metrics
					forceDrawNew = true
				}

				editor.sectionCache["Main."+lbl] = vec2(iSc1, iSc2)
			}

			j++
		}
	}
}

func (editor *settingsEditor) buildMainSection(jsonPath, sPath, name string, u reflect.Value, d reflect.StructField) {
	dRunLock := editor.tryLockLive(d)

	posLocal := imgui.CursorPos()

	imgui.PushFont(Font48)
	imgui.TextUnformatted(name)

	imgui.PopFont()
	imgui.Separator()

	editor.traverseChildren(jsonPath, sPath, u, reflect.StructField{})

	if jsonPath == "##Credentials" {
		centerTable("authbutton", -1, func() {
			if imgui.Button("Reset Tokens##auth") {
				settings.Credentails.AccessToken = ""
				settings.Credentails.RefreshToken = ""
				settings.Credentails.Expiry = time.Unix(0, 0)
			}

			imgui.SameLine()

			if settings.Credentails.AuthType == "AuthorizationCode" {
				if imgui.Button("Copy callback URL##auth") {
					glfw.GetCurrentContext().SetClipboardString("http://localhost:" + strconv.Itoa(settings.Credentails.CallbackPort))
				}

				imgui.SameLine()
			}

			if imgui.Button("Authorize##auth") {
				osuapi.Authorize(func(result osuapi.AuthResult, message string) {
					goroutines.RunOS(func() {
						switch result {
						case osuapi.AuthError:
							showMessage(mError, message)
						default:
							showMessage(mInfo, message)
						}
					})
				})
			}

			imgui.SameLine()

			if imgui.Button("Help##auth") {
				platform.OpenURL("https://github.com/Wieku/danser-go/wiki/APIv2-Tutorial")
			}
		})
	}

	posLocal1 := imgui.CursorPos()

	scrY := imgui.ScrollY()
	if scrY >= posLocal.Y-padY*2 && scrY <= posLocal1.Y {
		editor.active = name
	}

	if dRunLock {
		editor.unlockLive(true)
	}
}

func (editor *settingsEditor) subSectionTempl(name string, afterTitle, content func()) {
	pos := imgui.CursorScreenPos()

	imgui.Dummy(vec2(3, 0))
	imgui.SameLine()

	imgui.BeginGroup()

	imgui.PushFont(Font24)
	imgui.TextUnformatted(strings.ToUpper(name))

	afterTitle()

	imgui.PopFont()

	imgui.WindowDrawList().AddLine(imgui.CursorScreenPos(), imgui.CursorScreenPos().Add(vec2(contentRegionMax().X, 0)), packColor(*imgui.StyleColorVec4(imgui.ColSeparator)))

	imgui.Spacing()

	content()

	imgui.EndGroup()

	pos1 := imgui.CursorScreenPos()

	pos1.X = pos.X

	imgui.WindowDrawList().AddLine(pos, pos1, packColor(vec4(1.0, 1.0, 1.0, 1.0)))
}

func (editor *settingsEditor) buildSubSection(jsonPath, sPath, name string, u reflect.Value, d reflect.StructField) {
	dRunLock := editor.tryLockLive(d)

	editor.subSectionTempl(name, func() {}, func() {
		editor.traverseChildren(jsonPath, sPath, u, d)
	})

	if dRunLock {
		editor.unlockLive(true)
	}
}

func (editor *settingsEditor) buildArray(jsonPath, sPath, name string, u reflect.Value, d reflect.StructField) {
	editor.subSectionTempl(name, func() {
		imgui.SameLine()
		imgui.Dummy(vec2(2, 0))
		imgui.SameLine()

		ImIO.SetFontGlobalScale(20.0 / 32)
		imgui.PushFont(FontAw)

		if imgui.Button("+" + jsonPath) {
			if fName, ok := d.Tag.Lookup("new"); ok {
				u.Set(reflect.Append(u, reflect.ValueOf(settings.DefaultsFactory).MethodByName(fName).Call(nil)[0]))
			}
		}

		ImIO.SetFontGlobalScale(1)
		imgui.PopFont()
	}, func() {
		for j := 0; j < u.Len(); j++ {
			if editor.buildArrayElement(fmt.Sprintf("%s[%d]", jsonPath, j), sPath, u.Index(j), d, j) && u.Len() > 1 {
				u.Set(reflect.AppendSlice(u.Slice(0, j), u.Slice(j+1, u.Len())))
				j--
			}
		}
	})
}

func (editor *settingsEditor) buildArrayElement(jsonPath, sPath string, u reflect.Value, d reflect.StructField, childNum int) (removed bool) {
	if editor.searchCache[sPath] == 0 {
		return false
	}

	if childNum > 0 {
		imgui.Dummy(vec2(0, padY/3))
	}

	contentAvail := imgui.ContentRegionAvail().X

	if imgui.BeginTableV(jsonPath+"tae", 2, imgui.TableFlagsSizingStretchProp|imgui.TableFlagsNoPadInnerX|imgui.TableFlagsNoPadOuterX|imgui.TableFlagsNoClip, vec2(contentAvail, 0), contentAvail) {
		bWidth := imgui.FontSize() + imgui.CurrentStyle().FramePadding().X*2 + imgui.CurrentStyle().ItemSpacing().X*2 + 1

		imgui.TableSetupColumnV(jsonPath+"tae1", imgui.TableColumnFlagsWidthFixed, bWidth, imgui.ID(0))
		imgui.TableSetupColumnV(jsonPath+"tae2", imgui.TableColumnFlagsWidthFixed, contentAvail-bWidth, imgui.ID(1))

		imgui.TableNextColumn()
		imgui.TableNextColumn()

		pos := imgui.CursorScreenPos().Sub(vec2(0, imgui.CurrentStyle().FramePadding().Y-1))
		posLocal := imgui.CursorPos()

		imgui.Dummy(vec2(3, 0))
		imgui.SameLine()

		imgui.BeginGroup()

		editor.traverseChildren(jsonPath, sPath, u, d)

		imgui.EndGroup()

		pos1 := imgui.CursorScreenPos().Sub(vec2(0, imgui.CurrentStyle().ItemSpacing().Y))
		posLocal1 := imgui.CursorPos().Sub(vec2(0, imgui.CurrentStyle().ItemSpacing().Y))

		pos1.X = pos.X

		imgui.WindowDrawList().AddLine(pos, pos1, packColor(vec4(1.0, 0.6, 1.0, 1.0)))

		imgui.TableSetColumnIndex(0)

		imgui.Dummy(vec2(1, 0))
		imgui.SameLine()

		ImIO.SetFontGlobalScale(0.625)
		imgui.PushFont(FontAw)

		imgui.SetCursorPos(vec2(imgui.CursorPosX(), (posLocal.Y+posLocal1.Y-imgui.FrameHeight())/2))

		removed = imgui.Button("\uF068" + jsonPath)

		ImIO.SetFontGlobalScale(1)
		imgui.PopFont()

		imgui.SameLine()
		imgui.Dummy(vec2(2, 0))

		imgui.EndTable()
	}

	return
}

func (editor *settingsEditor) traverseChildren(jsonPath, lPath string, u reflect.Value, d reflect.StructField) {
	typ := u.Elem()
	def := u.Type().Elem()

	if u.Type().AssignableTo(reflect.TypeOf(&settings.HSV{})) { // special case, if it's an array of colors we want to see color picker instead of Hue, Saturation and Value sliders
		editor.buildColor(jsonPath, u, d, false)
		return
	}

	count := typ.NumField()

	skipMap := make(map[string]uint8)
	consumed := make(map[string]uint8)

	notFirst := false
	wasRendered := false
	wasSection := false

	for i := 0; i < count; i++ {
		field := typ.Field(i)
		dF := def.Field(i)

		if (!field.CanInterface() && (!dF.Anonymous && dF.Tag.Get("vector") == "")) || dF.Tag.Get("skip") != "" {
			continue
		}

		if editor.shouldBeHidden(consumed, skipMap, typ, dF) {
			continue
		}

		label := editor.getLabel(def.Field(i))

		sPath2 := lPath + "." + label

		if editor.searchCache[sPath2] == 0 {
			continue
		}

		jsonPath1 := jsonPath + "." + dF.Name

		if tD, ok := dF.Tag.Lookup("json"); ok {
			sp := strings.Split(tD, ",")[0]

			if sp != "" {
				jsonPath1 = jsonPath + "." + sp
			}
		}

		if wasRendered {
			notFirst = true

			imgui.Dummy(vec2(0, 2))
		}

		wasRendered = true

		switch field.Type().Kind() {
		case reflect.String, reflect.Float64, reflect.Int64, reflect.Int, reflect.Int32, reflect.Bool, reflect.Slice, reflect.Ptr:
			if wasSection {
				imgui.Dummy(vec2(0, padY/2))
			}

			isSection := false

			switch field.Type().Kind() {
			case reflect.String:
				if _, ok := dF.Tag.Lookup("vector"); ok {
					lName, ok1 := dF.Tag.Lookup("left")
					rName, ok2 := dF.Tag.Lookup("right")
					if !ok1 || !ok2 {
						break
					}

					l := typ.FieldByName(lName)
					ld, _ := def.FieldByName(lName)

					r := typ.FieldByName(rName)
					rd, _ := def.FieldByName(rName)

					jsonPathL := jsonPath + "." + lName
					jsonPathR := jsonPath + "." + rName

					editor.buildVector(jsonPathL, jsonPathR, dF, l, ld, r, rd)
				} else {
					editor.buildString(jsonPath1, field, dF)
				}
			case reflect.Float64:
				editor.buildFloat(jsonPath1, field, dF)
			case reflect.Int64, reflect.Int, reflect.Int32:
				editor.buildInt(jsonPath1, field, dF)
			case reflect.Bool:
				editor.buildBool(jsonPath1, field, dF)
			case reflect.Slice:
				if notFirst {
					imgui.Dummy(vec2(0, padY/2))
				}

				editor.buildArray(jsonPath1, sPath2, label, field, dF)
				isSection = true
			case reflect.Ptr:
				if field.Type().AssignableTo(reflect.TypeOf(&settings.HSV{})) {
					editor.buildColor(jsonPath1, field, dF, true)
				} else if !field.IsNil() {
					if dF.Anonymous {
						editor.traverseChildren(jsonPath, sPath2, field, dF)
					} else if field.CanInterface() {
						if notFirst {
							imgui.Dummy(vec2(0, padY/2))
						}

						editor.buildSubSection(jsonPath1, sPath2, label, field, dF)
						isSection = true
					} else {
						isSection = wasSection
						wasRendered = false
					}
				}
			}

			wasSection = isSection
		default:
			wasRendered = false
		}
	}
}

func (editor *settingsEditor) shouldBeHidden(consumed map[string]uint8, hidden map[string]uint8, parent reflect.Value, currentSField reflect.StructField) bool {
	if _, ok := currentSField.Tag.Lookup("vector"); ok {
		lName, ok1 := currentSField.Tag.Lookup("left")
		rName, ok2 := currentSField.Tag.Lookup("right")
		if ok1 && ok2 {
			hidden[lName] = 1
			hidden[rName] = 1
		}
	}

	if hidden[currentSField.Name] > 0 {
		return true
	}

	if s, ok := currentSField.Tag.Lookup("showif"); ok {
		s1 := strings.Split(s, "=")

		// Show only if dependant field is not hidden
		if hidden[s1[0]] == 1 {
			hidden[currentSField.Name] = 1
			return true
		}

		if s1[1] != "!" {
			fld := parent.FieldByName(s1[0])

			cF := fld.String()
			if fld.CanInt() {
				cF = strconv.Itoa(int(fld.Int()))
			} else if fld.Kind() == reflect.Bool {
				cF = "false"
				if fld.Bool() {
					cF = "true"
				}
			}

			found := false

			for _, toCheck := range strings.Split(s1[1], ",") {
				if toCheck[:1] == "!" {
					found = cF != toCheck[1:]

					if !found {
						break
					}
				} else if cF == toCheck {
					found = true
					break
				}
			}

			if !found {
				hidden[currentSField.Name] = 1
				return true
			}

			consumed[s1[0]] = 1
		} else if consumed[s1[0]] == 1 {
			return true
		}
	}

	return false
}

func (editor *settingsEditor) getLabel(d reflect.StructField) string {
	if lb, ok := d.Tag.Lookup("label"); ok {
		return lb
	}

	dName := strings.Title(d.Name)

	parts := nameSplitter.FindAllString(dName, -1)
	for i := 1; i < len(parts); i++ {
		parts[i] = strings.ToLower(parts[i])
	}

	return strings.Join(parts, " ")
}

func (editor *settingsEditor) buildBool(jsonPath string, f reflect.Value, d reflect.StructField) {
	editor.drawComponent(jsonPath, editor.getLabel(d), false, true, -1, d, func() {
		base := f.Bool()

		if imgui.Checkbox(jsonPath, &base) {
			f.SetBool(base)
			editor.search()
		}
	})
}

func (editor *settingsEditor) buildVector(jsonPath1, jsonPath2 string, d reflect.StructField, l reflect.Value, ld reflect.StructField, r reflect.Value, rd reflect.StructField) {
	drawBox := func() {
		contentAvail := imgui.ContentRegionAvail().X

		if imgui.BeginTableV("tv"+jsonPath1, 3, imgui.TableFlagsSizingStretchProp, vec2(contentAvail, 0), contentAvail) {
			imgui.TableSetupColumnV("tv1"+jsonPath1, imgui.TableColumnFlagsWidthStretch, 0, imgui.ID(0))
			imgui.TableSetupColumnV("tv2"+jsonPath1, imgui.TableColumnFlagsWidthFixed, 0, imgui.ID(1))
			imgui.TableSetupColumnV("tv3"+jsonPath1, imgui.TableColumnFlagsWidthStretch, 0, imgui.ID(2))

			imgui.TableNextColumn()

			imgui.SetNextItemWidth(-1)

			if l.CanInt() {
				editor.buildIntBox(jsonPath1, l, ld)
			} else {
				editor.buildFloatBox(jsonPath1, l, ld)
			}

			imgui.TableNextColumn()

			imgui.TextUnformatted("x")

			imgui.TableNextColumn()

			imgui.SetNextItemWidth(-1)

			if r.CanInt() {
				editor.buildIntBox(jsonPath2, r, rd)
			} else {
				editor.buildFloatBox(jsonPath2, r, rd)
			}

			imgui.EndTable()
		}
	}

	cSpec, okC := d.Tag.Lookup("combo")

	editor.drawComponent(jsonPath1+"\n"+jsonPath2, editor.getLabel(d), false, false, -1, d, func() {
		imgui.SetNextItemWidth(-1)

		if okC {
			baseL := l.Int()
			baseR := r.Int()

			var values [][2]int
			var labels []string

			lb := fmt.Sprintf("%dx%d", baseL, baseR)

			hasCustom := false
			normalFound := false

			for _, s := range strings.Split(cSpec, ",") {
				if s == "custom" {
					hasCustom = true
					continue
				}

				splt := strings.Split(s, "|")
				splt2 := strings.Split(splt[0], "x")

				cL, _ := strconv.Atoi(splt2[0])
				cR, _ := strconv.Atoi(splt2[1])

				optionLabel := fmt.Sprintf("%dx%d", cL, cR)
				if len(splt) > 1 {
					optionLabel = splt[1]
				}

				values = append(values, [2]int{cL, cR})
				labels = append(labels, optionLabel)

				if int(baseL) == cL && int(baseR) == cR {
					lb = optionLabel
					normalFound = true
				}
			}

			jsonPath := jsonPath1 + ":" + jsonPath2

			if imgui.BeginCombo("##combo"+jsonPath, lb) {
				handleDragScroll()

				justOpened := imgui.IsWindowAppearing()
				editor.blockSearch = true

				for i, lbl := range labels {
					if selectableFocus(lbl+jsonPath, lbl == lb, justOpened) {
						l.SetInt(int64(values[i][0]))
						r.SetInt(int64(values[i][1]))
						editor.search()
					}
				}

				if hasCustom {
					if !normalFound {
						pad := vec2(imgui.CurrentStyle().FramePadding().X, imgui.CurrentStyle().ItemSpacing().Y*0.5)
						scPos := imgui.CursorScreenPos().Sub(pad)

						imgui.WindowDrawList().AddRectFilled(scPos, scPos.Add(vec2(imgui.ContentRegionAvail().X, imgui.FrameHeight()).Add(pad.Mul(2))), packColor(*imgui.StyleColorVec4(imgui.ColHeader)))
					}

					drawBox()
				}

				imgui.EndCombo()
			}
		} else {
			drawBox()
		}

	})
}

func (editor *settingsEditor) buildFloatBox(jsonPath string, f reflect.Value, d reflect.StructField) {
	min := float64(parseFloatOr(d.Tag.Get("min"), 0))
	max := float64(parseFloatOr(d.Tag.Get("max"), 1))
	scale := float64(parseFloatOr(d.Tag.Get("scale"), 1))

	base := f.Float()

	valSpeed := base * scale

	valText := mutils.FormatWOZeros(valSpeed, 4)
	prevText := valText

	if inputTextV(jsonPath, &valText, imgui.InputTextFlagsCharsScientific, nil) {
		parsed, err := strconv.ParseFloat(valText, 64)
		if err != nil {
			valText = prevText
		} else {
			parsed = mutils.Clamp(parsed/scale, min, max)
			f.SetFloat(parsed)
		}
	}
}

func (editor *settingsEditor) buildIntBox(jsonPath string, f reflect.Value, d reflect.StructField) {
	min := parseIntOr(d.Tag.Get("min"), 0)
	max := parseIntOr(d.Tag.Get("max"), 100)

	base := int32(f.Int())

	if imgui.InputIntV(jsonPath, &base, 1, 1, 0) {
		base = mutils.Clamp(base, int32(min), int32(max))
		f.SetInt(int64(base))
	}
}

func (editor *settingsEditor) buildString(jsonPath string, f reflect.Value, d reflect.StructField) {
	cWidth := float32(-1)
	_, okKey := d.Tag.Lookup("key")

	if okKey {
		cWidth = 120
	}

	editor.drawComponent(jsonPath, editor.getLabel(d), d.Tag.Get("long") != "", false, cWidth, d, func() {
		imgui.SetNextItemWidth(-1)

		base := f.String()

		pDesc, okP := d.Tag.Lookup("path")
		fDesc, okF := d.Tag.Lookup("file")
		cSpec, okC := d.Tag.Lookup("combo")
		cFunc, okCS := d.Tag.Lookup("comboSrc")
		_, okPW := d.Tag.Lookup("password")

		if okKey {
			if imgui.ButtonV(base+"##"+jsonPath, vec2(-1, 0)) {
				editor.keyChangeVal = f
				editor.keyChange = jsonPath
				editor.keyChangeOpened = true
			}

			if editor.keyChange == jsonPath {
				imgui.SetNextWindowFocus()

				popupSmall("KeyChange"+jsonPath, &editor.keyChangeOpened, true, 0, 0, func() {
					width := imgui.CalcTextSizeV("Click outside this box to cancel", false, 0).X + 30

					centerTable("KeyChange1"+jsonPath, width, func() {
						imgui.TextUnformatted("Press any key...")
					})

					centerTable("KeyChange2"+jsonPath, width, func() {
						imgui.TextUnformatted("Click outside this box to cancel")
					})
				})

				if !editor.keyChangeOpened {
					editor.keyChange = ""
					imgui.SetWindowFocus()
				}
			}

		} else if okP || okF {
			if imgui.BeginTableV("tbr"+jsonPath, 2, imgui.TableFlagsSizingStretchProp, vec2(-1, 0), -1) {
				imgui.TableSetupColumnV("tbr1"+jsonPath, imgui.TableColumnFlagsWidthStretch, 0, imgui.ID(0))
				imgui.TableSetupColumnV("tbr2"+jsonPath, imgui.TableColumnFlagsWidthFixed, 0, imgui.ID(1))

				imgui.TableNextColumn()

				imgui.SetNextItemWidth(-1)

				if inputText(jsonPath, &base) {
					f.SetString(base)
				}

				imgui.TableNextColumn()

				if imgui.Button("Browse" + jsonPath) {
					dir := getAbsPath(base)

					if strings.TrimSpace(base) != "" && okF {
						dir = filepath.Dir(dir)
					}

					if _, err := os.Lstat(dir); err != nil {
						dir = env.DataDir()
					}

					var p string
					var err error

					if okP {
						p, err = dialog.Directory().Title(pDesc).SetStartDir(dir).Browse()
					} else {
						spl := strings.Split(d.Tag.Get("filter"), "|")
						p, err = dialog.File().Title(fDesc).Filter(spl[0], strings.Split(spl[1], ",")...).SetStartDir(dir).Load()
					}

					if err == nil {
						oD := strings.TrimSuffix(strings.ReplaceAll(base, "\\", "/"), "/")
						nD := strings.TrimSuffix(strings.ReplaceAll(p, "\\", "/"), "/")

						if nD != oD {
							f.SetString(getRelativeOrABSPath(p))
						}
					}
				}

				imgui.EndTable()
			}
		} else if okC || okCS {
			var values []string
			var labels []string

			var options []string

			if okCS {
				options = reflect.ValueOf(settings.DefaultsFactory).MethodByName(cFunc).Call(nil)[0].Interface().([]string)
			} else {
				options = strings.Split(cSpec, ",")
			}

			lb := base

			for _, s := range options {
				splt := strings.Split(s, "|")

				optionLabel := splt[0]
				if len(splt) > 1 {
					optionLabel = splt[1]
				}

				values = append(values, splt[0])
				labels = append(labels, optionLabel)

				if base == splt[0] {
					lb = optionLabel
				}
			}

			if _, okSearch := d.Tag.Lookup("search"); okSearch {
				mWidth := imgui.CalcItemWidth() - imgui.CurrentStyle().FramePadding().X*2

				if imgui.BeginComboV(jsonPath, lb, imgui.ComboFlagsHeightLarge) {
					editor.blockSearch = true

					for _, s := range labels {
						mWidth = max(mWidth, imgui.CalcTextSizeV(s, false, 0).X+20)
					}

					imgui.SetNextItemWidth(mWidth)

					cSearch := editor.comboSearch[jsonPath]

					focusScroll := searchBox("##search"+jsonPath, &cSearch)

					editor.comboSearch[jsonPath] = cSearch

					if !imgui.IsMouseClickedBool(0) && !imgui.IsMouseClickedBool(1) && !imgui.IsAnyItemActive() && !editor.scrollCache[jsonPath] {
						imgui.SetKeyboardFocusHereV(-1)
					}

					imgui.PushStyleVarFloat(imgui.StyleVarFrameRounding, 0)
					imgui.PushStyleVarFloat(imgui.StyleVarFrameBorderSize, 0)
					imgui.PushStyleVarVec2(imgui.StyleVarFramePadding, vzero())
					imgui.PushStyleColorVec4(imgui.ColFrameBg, vec4(0, 0, 0, 0))

					searchResults := make([]string, 0, len(labels))
					searchValues := make([]string, 0, len(labels))

					search := strings.ToLower(cSearch)

					for i, s := range labels {
						if cSearch == "" || strings.Contains(strings.ToLower(s), search) {
							searchResults = append(searchResults, s)
							searchValues = append(searchValues, values[i])
						}
					}

					if len(searchResults) > 0 {
						sHeight := float32(min(8, len(searchResults)))*imgui.FrameHeightWithSpacing() - imgui.CurrentStyle().ItemSpacing().Y/2

						if imgui.BeginListBoxV("##listbox"+jsonPath, vec2(mWidth, sHeight)) {
							editor.scrollCache[jsonPath] = handleDragScroll()
							focusScroll = focusScroll || imgui.IsWindowAppearing()

							for i, l := range searchResults {
								if selectableFocus(l+jsonPath, l == lb, focusScroll) {
									f.SetString(searchValues[i])
									editor.search()
								}
							}

							imgui.EndListBox()
						}
					}

					imgui.PopStyleVar()
					imgui.PopStyleVar()
					imgui.PopStyleVar()
					imgui.PopStyleColor()

					imgui.EndCombo()
				}
			} else {
				if imgui.BeginCombo(jsonPath, lb) {
					handleDragScroll()

					justOpened := imgui.IsWindowAppearing()

					editor.blockSearch = true

					for i, l := range labels {
						if selectableFocus(l+jsonPath, l == lb, justOpened) {
							f.SetString(values[i])
							editor.search()
						}
					}

					imgui.EndCombo()
				}
			}
		} else if okPW {
			if imgui.BeginTableV("tpw"+jsonPath+"tb", 2, imgui.TableFlagsSizingStretchProp, vec2(-1, 0), -1) {
				imgui.TableSetupColumnV("tpw1"+jsonPath, imgui.TableColumnFlagsWidthStretch, 0, imgui.ID(0))
				imgui.TableSetupColumnV("tpw2"+jsonPath, imgui.TableColumnFlagsWidthFixed, 0, imgui.ID(1))

				show := editor.pwShowHide[jsonPath]

				iTFlags := imgui.InputTextFlagsNone
				if !show {
					iTFlags = imgui.InputTextFlagsPassword
				}

				imgui.TableNextColumn()

				imgui.SetNextItemWidth(-1)

				if inputTextV(jsonPath, &base, iTFlags, nil) {
					f.SetString(base)
				}

				imgui.TableNextColumn()

				tx := "Show"
				if show {
					tx = "Hide"
				}

				if imgui.ButtonV(tx+jsonPath, vec2(imgui.CalcTextSizeV("Show", false, 0).X+imgui.CurrentStyle().FramePadding().X*2, 0)) {
					editor.pwShowHide[jsonPath] = !editor.pwShowHide[jsonPath]
				}

				imgui.EndTable()
			}
		} else {
			if inputText(jsonPath, &base) {
				f.SetString(base)
			}
		}
	})
}

func (editor *settingsEditor) buildInt(jsonPath string, f reflect.Value, d reflect.StructField) {
	base := int32(f.Int())

	_, okS := d.Tag.Lookup("string")
	cSpec, okC := d.Tag.Lookup("combo")

	editor.drawComponent(jsonPath, editor.getLabel(d), !okS && !okC, false, -1, d, func() {
		imgui.SetNextItemWidth(-1)

		format := cmp.Or(d.Tag.Get("format"), "%d")

		if okC {
			var values []int
			var labels []string

			lb := fmt.Sprintf(format, base)

			hasCustom := false

			for _, s := range strings.Split(cSpec, ",") {
				if s == "custom" {
					hasCustom = true
					continue
				}

				splt := strings.Split(s, "|")
				c, _ := strconv.Atoi(splt[0])

				optionLabel := fmt.Sprintf(format, c)
				if len(splt) > 1 {
					optionLabel = splt[1]
				}

				values = append(values, c)
				labels = append(labels, optionLabel)

				if int(base) == c {
					lb = optionLabel
				}
			}

			if imgui.BeginCombo(jsonPath, lb) {
				handleDragScroll()

				justOpened := imgui.IsWindowAppearing()
				editor.blockSearch = true

				for i, l := range labels {
					if selectableFocus(l+jsonPath, l == lb, justOpened) {
						f.SetInt(int64(values[i]))
						editor.search()
					}
				}

				if hasCustom {
					min := parseIntOr(d.Tag.Get("min"), 0)
					max := parseIntOr(d.Tag.Get("max"), 100)

					if base >= int32(min) {
						pad := vec2(imgui.CurrentStyle().FramePadding().X, imgui.CurrentStyle().ItemSpacing().Y*0.5)
						scPos := imgui.CursorScreenPos().Sub(pad)

						imgui.WindowDrawList().AddRectFilled(scPos, scPos.Add(vec2(imgui.ContentRegionAvail().X, imgui.FrameHeight()).Add(pad.Mul(2))), packColor(*imgui.StyleColorVec4(imgui.ColHeader)))
					} else {
						base = 0
					}

					imgui.AlignTextToFramePadding()
					imgui.TextUnformatted("Custom:")

					imgui.SameLine()

					imgui.SetNextItemWidth(imgui.ContentRegionAvail().X)

					if imgui.InputIntV(jsonPath, &base, 1, 1, 0) {
						base = mutils.Clamp(base, int32(min), int32(max))
						f.SetInt(int64(base))
					}
				}

				imgui.EndCombo()
			}
		} else if okS {
			editor.buildIntBox(jsonPath, f, d)
		} else {
			min := parseIntOr(d.Tag.Get("min"), 0)
			max := parseIntOr(d.Tag.Get("max"), 100)

			imgui.PushStyleVarVec2(imgui.StyleVarFramePadding, vec2(0, -3))

			if sliderIntSlide(jsonPath, &base, int32(min), int32(max), "##"+format, imgui.SliderFlagsNoInput) {
				f.SetInt(int64(base))
			}

			imgui.PopStyleVar()

			if imgui.IsItemHovered() || imgui.IsItemActive() {
				imgui.SetKeyboardFocusHereV(-1)
				editor.blockSearch = true

				imgui.BeginTooltip()
				setTooltip(fmt.Sprintf(format, base))
				imgui.EndTooltip()
			}
		}
	})
}

func (editor *settingsEditor) buildFloat(jsonPath string, f reflect.Value, d reflect.StructField) {
	editor.drawComponent(jsonPath, editor.getLabel(d), d.Tag.Get("string") == "", false, -1, d, func() {
		imgui.SetNextItemWidth(-1)

		if d.Tag.Get("string") != "" {
			editor.buildFloatBox(jsonPath, f, d)
		} else {
			minV := parseFloat64Or(d.Tag.Get("min"), 0)
			maxV := parseFloat64Or(d.Tag.Get("max"), 1)
			scale := parseFloat64Or(d.Tag.Get("scale"), 1)
			format := cmp.Or(d.Tag.Get("format"), "%.2f")

			base := f.Float()
			valSpeed := base * scale

			imgui.PushStyleVarVec2(imgui.StyleVarFramePadding, vec2(0, -3))

			cSpacing := imgui.CurrentStyle().ItemSpacing()
			imgui.PushStyleVarVec2(imgui.StyleVarItemSpacing, vec2(cSpacing.X, cSpacing.Y-3))

			if sliderFloatSlide(jsonPath, &valSpeed, minV*scale, maxV*scale, "##"+format, imgui.SliderFlagsNoInput) {
				f.SetFloat(float64(valSpeed / scale))
			}

			imgui.PopStyleVar()
			imgui.PopStyleVar()

			if imgui.IsItemHovered() || imgui.IsItemActive() {
				imgui.SetKeyboardFocusHereV(-1)
				editor.blockSearch = true

				imgui.BeginTooltip()
				setTooltip(fmt.Sprintf(format, valSpeed))
				imgui.EndTooltip()
			}
		}
	})
}

func (editor *settingsEditor) buildColor(jsonPath string, f reflect.Value, d reflect.StructField, withLabel bool) {
	dComp := func() {
		imgui.SetNextItemWidth(imgui.ContentRegionAvail().X - 1)

		hsv := f.Interface().(*settings.HSV)

		r, g, b := color.HSVToRGB(float32(hsv.Hue), float32(hsv.Saturation), float32(hsv.Value))
		rgb := [3]float32{r, g, b}

		if imgui.ColorEdit3V(jsonPath, &rgb, imgui.ColorEditFlagsDisplayHSV|imgui.ColorEditFlagsNoLabel|imgui.ColorEditFlagsFloat) {
			h, s, v := color.RGBToHSV(rgb[0], rgb[1], rgb[2])
			hsv.Hue = float64(h)
			hsv.Saturation = float64(s)
			hsv.Value = float64(v)
		}

		editor.blockSearch = editor.blockSearch || imgui.IsWindowFocusedV(imgui.FocusedFlagsChildWindows) && !imgui.IsWindowFocused()
	}

	if withLabel {
		editor.drawComponent(jsonPath, editor.getLabel(d), false, false, -1, d, dComp)
	} else {
		dComp()
	}
}

func (editor *settingsEditor) drawComponent(jsonPath, label string, long, checkbox bool, customWidth float32, d reflect.StructField, draw func()) {
	dRunLock := editor.tryLockLive(d)

	width := imgui.FontSize() + imgui.CurrentStyle().FramePadding().X*2 - 1 // + imgui.CurrentStyle().ItemSpacing().X
	if !checkbox {
		if customWidth > 0 {
			width = customWidth
		} else {
			width = 240 + imgui.CalcTextSizeV("x", false, 0).X + imgui.CurrentStyle().FramePadding().X*4
		}
	}

	cCount := 1
	if !long {
		cCount = 2
	}

	contentAvail := imgui.ContentRegionAvail().X

	if imgui.BeginTableV("lbl"+jsonPath, int32(cCount), imgui.TableFlagsSizingStretchProp|imgui.TableFlagsNoPadInnerX|imgui.TableFlagsNoPadOuterX|imgui.TableFlagsNoClip, vec2(contentAvail, 0), contentAvail) {
		if !long {
			imgui.TableSetupColumnV("lbl1"+jsonPath, imgui.TableColumnFlagsWidthFixed, contentAvail-width, imgui.ID(0))
			imgui.TableSetupColumnV("lbl2"+jsonPath, imgui.TableColumnFlagsWidthFixed, width, imgui.ID(1))
		} else {
			imgui.TableSetupColumnV("lbl1"+jsonPath, imgui.TableColumnFlagsWidthFixed, contentAvail, imgui.ID(0))
		}

		imgui.TableNextColumn()

		tooltip, hasTooltip := d.Tag.Lookup("tooltip")

		if hasTooltip {
			label = "(!) " + label
		}

		imgui.BeginGroup()
		imgui.AlignTextToFramePadding()
		imgui.TextUnformatted(label)
		imgui.EndGroup()

		if imgui.IsItemHovered() {
			_, hidePath := d.Tag.Lookup("hidePath")

			showPath := !hidePath && launcherConfig.ShowJSONPaths

			if showPath || hasTooltip {
				imgui.BeginTooltip()

				tTip := ""
				if showPath {
					tTip = strings.ReplaceAll(jsonPath, "#", "")
				}

				if hasTooltip {
					if showPath {
						tTip += "\n\n"
					}

					tTip += tooltip
				}

				imgui.PushTextWrapPosV(400)

				imgui.TextUnformatted(tTip)

				imgui.PopTextWrapPos()
				imgui.EndTooltip()
			}
		}

		imgui.TableNextColumn()

		draw()

		imgui.EndTable()
	}

	if dRunLock {
		editor.unlockLive(false)
	}
}

func (editor *settingsEditor) tryLockLive(d reflect.StructField) bool {
	liveEdit := true

	if l, ok := d.Tag.Lookup("liveedit"); ok && l == "false" {
		liveEdit = false
	}

	dRunLock := !liveEdit && editor.danserRunning

	if dRunLock {
		imgui.BeginGroup()
		imgui.PushItemFlag(imgui.ItemFlags(imgui.ItemFlagsDisabled), true)
		imgui.PushStyleColorVec4(imgui.ColText, vec4(0.6, 0.6, 0.6, 1))
	}

	return dRunLock
}

func (editor *settingsEditor) unlockLive(plural bool) {
	imgui.PopStyleColor()
	imgui.PopItemFlag()
	imgui.EndGroup()

	if imgui.IsItemHovered() {
		imgui.BeginTooltip()

		if plural {
			imgui.TextUnformatted("These options can't be edited while danser is running.")
		} else {
			imgui.TextUnformatted("This option can't be edited while danser is running.")
		}

		imgui.EndTooltip()
	}
}

func parseIntOr(value string, alt int) int {
	if i, err := strconv.Atoi(value); err == nil {
		return i
	}

	return alt
}

func parseFloatOr(value string, alt float32) float32 {
	if i, err := strconv.ParseFloat(value, 32); err == nil {
		return float32(i)
	}

	return alt
}

func parseFloat64Or(value string, alt float64) float64 {
	if i, err := strconv.ParseFloat(value, 64); err == nil {
		return i
	}

	return alt
}
