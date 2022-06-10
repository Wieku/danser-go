package launcher

import (
	"fmt"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/sqweek/dialog"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/env"
	"github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/mutils"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

const padY = 30

var nameSplitter = regexp.MustCompile(`[A-Z]+[^A-Z]*`)

type settingsEditor struct {
	*popup

	searchCache  map[string]int
	scrollTo     string
	comboOpened  bool
	searchString string
	id           int
	current      *settings.Config

	listenerCalled bool
	closeListener  func()
	sectionCache   map[string]imgui.Vec2

	active     string
	lastActive string
}

func newSettingsEditor(config *settings.Config, closeListener func()) *settingsEditor {
	editor := &settingsEditor{
		popup:         newPopup("Settings Editor", popBig),
		searchCache:   make(map[string]int),
		closeListener: closeListener,
		sectionCache:  make(map[string]imgui.Vec2),
	}

	editor.internalDraw = editor.drawEditor

	editor.current = config
	editor.buildSearchCache("Main", reflect.ValueOf(config), editor.searchString, false)

	return editor
}

func (editor *settingsEditor) shouldClose() bool {
	if !editor.opened && !editor.listenerCalled {
		editor.closeListener()
		editor.listenerCalled = true
	}

	return !editor.opened
}

func (editor *settingsEditor) drawEditor() {
	settings.General.OsuSkinsDir = editor.current.General.OsuSkinsDir

	imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, imgui.Vec2{0, 0})
	imgui.PushStyleColor(imgui.StyleColorWindowBg, imgui.Vec4{X: 0, Y: 0, Z: 0, W: .9})
	imgui.PushStyleColor(imgui.StyleColorFrameBg, imgui.Vec4{X: .2, Y: .2, Z: .2, W: 1})

	imgui.PushFont(Font20)
	{
		editor.id = 0

		imgui.PushStyleColor(imgui.StyleColorChildBg, imgui.Vec4{X: 0, Y: 0, Z: 0, W: .9})

		imgui.PushFont(FontAw)
		{
			if imgui.BeginChildV(editor.getId(), imgui.Vec2{X: imgui.FontSize()*1.5 + 15, Y: -1}, false, imgui.WindowFlagsAlwaysVerticalScrollbar /*imgui.ScrollbWindowFlagsNoScrollbar*/) {
				editor.scrollTo = ""
				editor.buildNavigationFor(editor.current)
			}
			imgui.EndChild()
		}
		imgui.PopFont()

		imgui.PopStyleColor()

		imgui.SameLine()

		imgui.SetCursorPos(imgui.Vec2{X: imgui.CursorPosX(), Y: imgui.WindowContentRegionMin().Y})

		if imgui.BeginChildV(editor.getId(), imgui.Vec2{X: -1, Y: -1}, false, 0) {
			imgui.PushFont(Font32)
			{
				imgui.PushItemWidth(-1)
				{
					if imgui.InputTextWithHint(editor.getId(), "Search", &editor.searchString) {
						editor.searchCache = make(map[string]int)
						editor.buildSearchCache("Main", reflect.ValueOf(editor.current), editor.searchString, false)
					}

					if !editor.comboOpened && !imgui.IsAnyItemActive() && !imgui.IsMouseClicked(0) {
						imgui.SetKeyboardFocusHereV(-1)
					}
				}
				imgui.PopItemWidth()
			}
			imgui.PopFont()

			imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, imgui.Vec2{5, 5})

			if imgui.BeginChildV(editor.getId(), imgui.Vec2{-1, -1}, false, imgui.WindowFlagsAlwaysUseWindowPadding) {
				editor.comboOpened = false

				editor.drawSettings()
			}

			imgui.EndChild()

			imgui.PopStyleVar()
		}
		imgui.EndChild()
	}
	imgui.PopFont()

	imgui.PopStyleColor()
	imgui.PopStyleColor()
	imgui.PopStyleVar()
}

func (editor *settingsEditor) buildSearchCache(path string, u reflect.Value, search string, omitSearch bool) bool {
	editor.sectionCache = make(map[string]imgui.Vec2)
	typ := u.Elem()
	def := u.Type().Elem()

	count := typ.NumField()

	found := false

	for i := 0; i < count; i++ {
		field := typ.Field(i)
		fT := field.Type()
		sT := fT.Kind()

		if /*!field.CanInterface() || */ def.Field(i).Tag.Get("skip") != "" {
			continue
		}

		label := editor.getLabel(def.Field(i))

		sPath := path + "." + label

		match := omitSearch || strings.Contains(strings.ToLower(label), search)

		if sT == reflect.Ptr && (field.CanInterface() || def.Field(i).Anonymous) && !field.IsNil() && !field.Type().AssignableTo(reflect.TypeOf(&settings.HSV{})) {
			sub := editor.buildSearchCache(sPath, field, search, match)
			match = match || sub
		} else if sT == reflect.Slice && field.CanInterface() {
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

	imgui.PushStyleColor(imgui.StyleColorButton, imgui.Vec4{0, 0, 0, 0})
	imgui.PushStyleColor(imgui.StyleColorButtonHovered, imgui.Vec4{0.3, 0.3, 0.3, 0.6})
	imgui.PushStyleColor(imgui.StyleColorButtonActive, imgui.Vec4{0.5, 0.5, 0.5, 0.6})

	buttonSize := imgui.FontSize() * 1.5

	cAvail := imgui.ContentRegionAvail().Y
	sc1 := imgui.ScrollY()
	sc2 := sc1 + cAvail

	for i := 0; i < count; i++ {
		label := editor.getLabel(def.Field(i))

		if editor.searchCache["Main."+label] > 0 && (typ.Field(i).CanInterface() && !typ.Field(i).IsNil()) {
			if editor.active == label {
				cColor := imgui.CurrentStyle().Color(imgui.StyleColorCheckMark)

				imgui.PushStyleColor(imgui.StyleColorText, imgui.Vec4{cColor.X * 1.2, cColor.Y * 1.2, cColor.Z * 1.2, 1})
			}

			c1 := imgui.CursorPos().Y

			if imgui.ButtonV(def.Field(i).Tag.Get("icon"), imgui.Vec2{buttonSize, buttonSize}) {
				editor.scrollTo = label
			}

			c2 := imgui.CursorPos().Y

			if editor.active == label {
				if editor.lastActive != editor.active {
					if c2 > sc2 {
						imgui.SetScrollY(c2 - cAvail)
					}

					if c1 < sc1 {
						imgui.SetScrollY(c1)
					}

					editor.lastActive = editor.active
				}

				imgui.PopStyleColor()
			}

			if imgui.IsItemHovered() {
				imgui.PushFont(Font24)
				imgui.BeginTooltip()
				imgui.SetTooltip(label)
				imgui.EndTooltip()
				imgui.PopFont()
			}
		}
	}

	imgui.PopStyleColor()
	imgui.PopStyleColor()
	imgui.PopStyleColor()
}

func (editor *settingsEditor) drawSettings() {
	rVal := reflect.ValueOf(editor.current)

	typ := rVal.Elem()
	def := rVal.Type().Elem()

	count := typ.NumField()

	sc1 := imgui.ScrollY()
	sc2 := sc1 + imgui.ContentRegionAvail().Y

	for i, j := 0, 0; i < count; i++ {
		field := typ.Field(i)
		dF := def.Field(i)

		lbl := editor.getLabel(dF)

		if editor.searchCache["Main."+lbl] == 0 {
			continue
		}

		if field.CanInterface() && field.Type().Kind() == reflect.Ptr && !field.IsNil() {
			if j > 0 {
				imgui.Dummy(imgui.Vec2{0, 2 * padY})
			}

			drawNew := true
			if v, ok := editor.sectionCache["Main."+lbl]; ok {
				if editor.scrollTo == lbl {
					imgui.SetScrollY(v.X)
				}

				if sc1 > v.Y || sc2 < v.X {
					drawNew = false

					imgui.SetCursorPos(imgui.Vec2{0, imgui.CursorPos().Y + v.Y - v.X})
				}
			}

			if drawNew {
				iSc1 := imgui.CursorPos().Y

				editor.buildMainSection("Main."+lbl, lbl, field)

				iSc2 := imgui.CursorPos().Y

				editor.sectionCache["Main."+lbl] = imgui.Vec2{iSc1, iSc2}
			}

			j++
		}
	}
}

func (editor *settingsEditor) buildMainSection(sPath, name string, u reflect.Value) {
	posLocal := imgui.CursorPos()

	imgui.PushFont(Font48)
	imgui.Text(name)

	imgui.PopFont()
	imgui.Separator()

	editor.traverseChildren(sPath, u, reflect.StructField{})

	posLocal1 := imgui.CursorPos()

	scrY := imgui.ScrollY()
	if scrY >= posLocal.Y-padY*2 && scrY <= posLocal1.Y /*+padY*/ {
		editor.active = name
	}
}

func (editor *settingsEditor) subSectionTempl(sPath, name string, first, last bool, afterTitle, content func()) {
	if editor.searchCache[sPath] == 0 {
		return
	}

	if !first {
		imgui.Dummy(imgui.Vec2{0, padY / 2})
	}

	pos := imgui.CursorScreenPos()

	imgui.Dummy(imgui.Vec2{5, 0})
	imgui.SameLine()

	imgui.BeginGroup()

	imgui.PushFont(Font24)
	imgui.Text(strings.ToUpper(name))

	afterTitle()

	imgui.PopFont()
	imgui.Separator()

	content()

	imgui.EndGroup()

	pos1 := imgui.CursorScreenPos()

	pos1.X = pos.X

	imgui.WindowDrawList().AddLine(pos, pos1, imgui.PackedColorFromVec4(imgui.Vec4{1.0, 1.0, 1.0, 1.0}))

	if !last {
		imgui.Dummy(imgui.Vec2{0, padY / 2})
	}
}

func (editor *settingsEditor) buildSubSection(sPath, name string, u reflect.Value, d reflect.StructField, first, last bool) {
	editor.subSectionTempl(sPath, name, first, last, func() {}, func() {
		editor.traverseChildren(sPath, u, d)
	})
}

func (editor *settingsEditor) buildArray(sPath, name string, u reflect.Value, d reflect.StructField, first, last bool) {
	editor.subSectionTempl(sPath, name, first, last, func() {
		imgui.SameLine()
		imgui.Dummy(imgui.Vec2{2, 0})
		imgui.SameLine()

		ImIO.SetFontGlobalScale(20.0 / 32)
		imgui.PushFont(FontAw)

		if imgui.Button("+" + editor.getId()) {
			log.Println(u, d)
			if fName, ok := d.Tag.Lookup("new"); ok {
				u.Set(reflect.Append(u, reflect.ValueOf(settings.DefaultsFactory).MethodByName(fName).Call(nil)[0]))
			}
		}

		ImIO.SetFontGlobalScale(1)
		imgui.PopFont()
	}, func() {
		var tRem = -1

		rCal := func(idx int) {
			tRem = idx
		}

		for j := 0; j < u.Len(); j++ {
			editor.buildArrayElement(sPath, u.Index(j), d, rCal, j)
		}

		if tRem > -1 && u.Len() > 1 {
			u.Set(reflect.AppendSlice(u.Slice(0, tRem), u.Slice(tRem+1, u.Len())))
		}
	})
}

func (editor *settingsEditor) buildArrayElement(sPath string, u reflect.Value, d reflect.StructField, removeCb func(idx int), childNum int) {
	if editor.searchCache[sPath] == 0 {
		return
	}

	if childNum > 0 {
		imgui.Dummy(imgui.Vec2{0, padY / 3})
	}

	if imgui.BeginTableV(editor.getId(), 2, imgui.TableFlagsSizingStretchProp|imgui.TableFlagsNoPadInnerX|imgui.TableFlagsNoPadOuterX|imgui.TableFlagsNoClip, imgui.Vec2{-1, 0}, -1) {
		imgui.TableSetupColumnV(editor.getId(), imgui.TableColumnFlagsWidthFixed, 0, uint(0))
		imgui.TableSetupColumnV(editor.getId(), imgui.TableColumnFlagsWidthStretch, 0, uint(1))

		imgui.TableNextColumn()
		imgui.TableNextColumn()

		pos := imgui.CursorScreenPos()
		posLocal := imgui.CursorPos()

		imgui.Dummy(imgui.Vec2{5, 0})
		imgui.SameLine()

		imgui.BeginGroup()

		editor.traverseChildren(sPath, u, d)

		imgui.EndGroup()

		pos1 := imgui.CursorScreenPos()
		posLocal1 := imgui.CursorPos()

		pos1.X = pos.X

		imgui.WindowDrawList().AddLine(pos, pos1, imgui.PackedColorFromVec4(imgui.Vec4{1.0, 0.6, 1.0, 1.0}))

		imgui.TableSetColumnIndex(0)

		imgui.Dummy(imgui.Vec2{2, 0})
		imgui.SameLine()

		ImIO.SetFontGlobalScale(0.625)
		imgui.PushFont(FontAw)

		imgui.SetCursorPos(imgui.Vec2{imgui.CursorPosX(), (posLocal.Y + posLocal1.Y - imgui.FrameHeight()) / 2})

		if imgui.Button("\uF068" + editor.getId()) {
			removeCb(childNum)
		}

		ImIO.SetFontGlobalScale(1)
		imgui.PopFont()

		imgui.SameLine()
		imgui.Dummy(imgui.Vec2{2, 0})

		imgui.EndTable()
	}
}

func (editor *settingsEditor) traverseChildren(lPath string, u reflect.Value, d reflect.StructField) {
	typ := u.Elem()
	def := u.Type().Elem()

	if u.Type().AssignableTo(reflect.TypeOf(&settings.HSV{})) {
		editor.buildColor(u, d, false)
		return
	}

	count := typ.NumField()

	skipMap := make(map[string]uint8)

	consumed := make(map[string]uint8)

	for i, index := 0, 0; i < count; i++ {
		field := typ.Field(i)
		dF := def.Field(i)

		if (!field.CanInterface() && (!dF.Anonymous && dF.Tag.Get("vector") == "")) || dF.Tag.Get("skip") != "" {
			continue
		}

		if s, ok := dF.Tag.Lookup("showif"); ok {
			s1 := strings.Split(s, "=")

			if consumed[s1[0]] == 1 {
				continue
			}

			if s1[1] != "!" {
				cF := typ.FieldByName(s1[0]).String()

				found := false

				for _, toCheck := range strings.Split(s1[1], ",") {
					if cF == toCheck {
						found = true
						consumed[s1[0]] = 1
						break
					}
				}

				if !found {
					continue
				}
			}
		}

		label := editor.getLabel(def.Field(i))

		sPath2 := lPath + "." + label

		if editor.searchCache[sPath2] == 0 || skipMap[dF.Name] == 1 {
			continue
		}

		if index > 0 {
			imgui.Dummy(imgui.Vec2{0, 2})
		}

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

				skipMap[lName] = 1
				skipMap[rName] = 1

				editor.buildVector(field, dF, l, ld, r, rd)
			} else {
				editor.buildString(field, dF)
			}
		case reflect.Float64:
			editor.buildFloat(field, dF)
		case reflect.Int64, reflect.Int, reflect.Int32:
			editor.buildInt(field, dF)
		case reflect.Bool:
			editor.buildBool(field, dF)
		case reflect.Slice:
			editor.buildArray(sPath2, label, field, dF, index == 0, index == count-1)
		case reflect.Ptr:
			if field.Type().AssignableTo(reflect.TypeOf(&settings.HSV{})) {
				editor.buildColor(field, dF, true)
			} else if !field.IsNil() {
				if dF.Anonymous {
					editor.traverseChildren(sPath2, field, dF)
				} else if field.CanInterface() {
					editor.buildSubSection(sPath2, label, field, dF, index == 0, index == count-1)
				} else {
					index--
				}
			}
		default:
			index--
		}

		index++
	}
}

func (editor *settingsEditor) getId() string {
	editor.id++
	return "##" + strconv.Itoa(editor.id)
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

func (editor *settingsEditor) buildBool(f reflect.Value, d reflect.StructField) {
	editor.drawComponent(editor.getLabel(d), false, true, d, func() {
		base := f.Interface().(bool)

		if imgui.Checkbox(editor.getId(), &base) {
			f.SetBool(base)
		}
	})
}

func (editor *settingsEditor) buildVector(f reflect.Value, d reflect.StructField, l reflect.Value, ld reflect.StructField, r reflect.Value, rd reflect.StructField) {
	editor.drawComponent(editor.getLabel(d), false, false, d, func() {
		if imgui.BeginTableV(editor.getId(), 3, imgui.TableFlagsSizingStretchProp, imgui.Vec2{-1, 0}, -1) {
			imgui.TableSetupColumnV(editor.getId(), imgui.TableColumnFlagsWidthStretch, 0, uint(0))
			imgui.TableSetupColumnV(editor.getId(), imgui.TableColumnFlagsWidthFixed, 0, uint(1))
			imgui.TableSetupColumnV(editor.getId(), imgui.TableColumnFlagsWidthStretch, 0, uint(2))

			imgui.TableNextColumn()

			imgui.SetNextItemWidth(-1)

			if l.CanInt() {
				editor.buildIntBox(l, ld)
			} else {
				editor.buildFloatBox(l, ld)
			}

			imgui.TableNextColumn()

			imgui.Text("x")

			imgui.TableNextColumn()

			imgui.SetNextItemWidth(-1)

			if r.CanInt() {
				editor.buildIntBox(r, rd)
			} else {
				editor.buildFloatBox(r, rd)
			}

			imgui.EndTable()
		}
	})
}

func (editor *settingsEditor) buildFloatBox(f reflect.Value, d reflect.StructField) {
	min := float64(parseFloatOr(d.Tag.Get("min"), 0))
	max := float64(parseFloatOr(d.Tag.Get("max"), 1))
	scale := float64(parseFloatOr(d.Tag.Get("scale"), 1))

	base := f.Float()

	valSpeed := base * scale

	valText := strconv.FormatFloat(valSpeed, 'f', 2, 64)
	prevText := valText

	if imgui.InputTextV(editor.getId(), &valText, imgui.InputTextFlagsCharsScientific, nil) {
		parsed, err := strconv.ParseFloat(valText, 64)
		if err != nil {
			valText = prevText
		} else {
			parsed = mutils.ClampF(parsed/scale, min, max)
			f.SetFloat(parsed)
		}
	}
}

func (editor *settingsEditor) buildIntBox(f reflect.Value, d reflect.StructField) {
	min := parseIntOr(d.Tag.Get("min"), 0)
	max := parseIntOr(d.Tag.Get("max"), 100)

	base := int32(f.Int())

	if imgui.InputIntV(editor.getId(), &base, 1, 1, 0) {
		base = mutils.Clamp(base, int32(min), int32(max))
		f.SetInt(int64(base))
	}
}

func (editor *settingsEditor) buildString(f reflect.Value, d reflect.StructField) {
	editor.drawComponent(editor.getLabel(d), d.Tag.Get("long") != "", false, d, func() {
		imgui.SetNextItemWidth(-1)

		base := f.String()

		pDesc, okP := d.Tag.Lookup("path")
		fDesc, okF := d.Tag.Lookup("file")
		cSpec, okC := d.Tag.Lookup("combo")
		cFunc, okCS := d.Tag.Lookup("comboSrc")

		if okP || okF {
			cId := editor.getId()
			if imgui.BeginTableV(cId, 2, imgui.TableFlagsSizingStretchProp, imgui.Vec2{-1, 0}, -1) {
				imgui.TableSetupColumnV(cId+"c1", imgui.TableColumnFlagsWidthStretch, 0, uint(0))
				imgui.TableSetupColumnV(cId+"c2", imgui.TableColumnFlagsWidthFixed, 0, uint(1))

				imgui.TableNextColumn()

				imgui.SetNextItemWidth(-1)

				if imgui.InputText(editor.getId(), &base) {
					f.SetString(base)
				}

				imgui.TableNextColumn()

				if imgui.Button("Browse" + editor.getId()) {
					var dir string

					if okP {
						dir = filepath.Join(env.DataDir(), base)

						if filepath.IsAbs(base) {
							dir = base
						}
					} else {
						dir = filepath.Dir(filepath.Join(env.DataDir(), base))

						if filepath.IsAbs(base) {
							dir = filepath.Dir(base)
						}
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
						log.Println(p)
						log.Println(env.DataDir())
						oD := strings.TrimSuffix(strings.ReplaceAll(base, "\\", "/"), "/")
						nD := strings.TrimSuffix(strings.ReplaceAll(p, "\\", "/"), "/")

						dD := strings.TrimSuffix(strings.ReplaceAll(env.DataDir(), "\\", "/"), "/") + "/"

						if nD != oD {
							f.SetString(strings.ReplaceAll(strings.TrimPrefix(nD, dD), "/", string(os.PathSeparator)))
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

			if imgui.BeginCombo(editor.getId(), lb) {
				editor.comboOpened = true

				for i, l := range labels {
					if imgui.SelectableV(l, l == lb, 0, imgui.Vec2{0, 0}) {
						f.SetString(values[i])
					}
				}

				imgui.EndCombo()
			}
		} else {
			if imgui.InputText(editor.getId(), &base) {
				f.SetString(base)
			}
		}
	})
}

func (editor *settingsEditor) buildInt(f reflect.Value, d reflect.StructField) {
	base := int32(f.Int())

	_, okS := d.Tag.Lookup("string")
	cSpec, okC := d.Tag.Lookup("combo")

	editor.drawComponent(editor.getLabel(d), !okS && !okC, false, d, func() {
		imgui.SetNextItemWidth(-1)

		format := firstOf(d.Tag.Get("format"), "%d")

		if okC {
			var values []int
			var labels []string

			lb := fmt.Sprintf(format, base)

			for _, s := range strings.Split(cSpec, ",") {
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

			if imgui.BeginCombo(editor.getId(), lb) {
				editor.comboOpened = true

				for i, l := range labels {
					if imgui.SelectableV(l, l == lb, 0, imgui.Vec2{0, 0}) {
						f.SetInt(int64(values[i]))
					}
				}

				imgui.EndCombo()
			}
		} else if okS {
			editor.buildIntBox(f, d)
		} else {
			min := parseIntOr(d.Tag.Get("min"), 0)
			max := parseIntOr(d.Tag.Get("max"), 100)

			imgui.PushStyleVarVec2(imgui.StyleVarFramePadding, imgui.Vec2{X: 0, Y: -3})

			if sliderIntSlide(editor.getId(), &base, int32(min), int32(max), "##"+format, imgui.SliderFlagsNoInput) {
				f.SetInt(int64(base))
			}

			imgui.PopStyleVar()

			if imgui.IsItemHovered() || imgui.IsItemActive() {
				imgui.BeginTooltip()
				imgui.SetTooltip(fmt.Sprintf(format, base))
				imgui.EndTooltip()
			}
		}
	})
}

func (editor *settingsEditor) buildFloat(f reflect.Value, d reflect.StructField) {
	editor.drawComponent(editor.getLabel(d), d.Tag.Get("string") == "", false, d, func() {
		imgui.SetNextItemWidth(-1)

		if d.Tag.Get("string") != "" {
			editor.buildFloatBox(f, d)
		} else {
			min := parseFloatOr(d.Tag.Get("min"), 0)
			max := parseFloatOr(d.Tag.Get("max"), 1)
			scale := parseFloatOr(d.Tag.Get("scale"), 1)
			format := firstOf(d.Tag.Get("format"), "%.2f")

			base := float32(f.Float())
			valSpeed := base * scale

			imgui.PushStyleVarVec2(imgui.StyleVarFramePadding, imgui.Vec2{X: 0, Y: -3})

			cSpacing := imgui.CurrentStyle().ItemSpacing()
			imgui.PushStyleVarVec2(imgui.StyleVarItemSpacing, imgui.Vec2{X: cSpacing.X, Y: cSpacing.Y - 3})

			if sliderFloatSlide(editor.getId(), &valSpeed, min*scale, max*scale, "##"+format, imgui.SliderFlagsNoInput) {
				f.SetFloat(float64(valSpeed / scale))
			}

			imgui.PopStyleVar()
			imgui.PopStyleVar()

			if imgui.IsItemHovered() || imgui.IsItemActive() {
				imgui.BeginTooltip()
				imgui.SetTooltip(fmt.Sprintf(format, valSpeed))
				imgui.EndTooltip()
			}
		}
	})
}

func (editor *settingsEditor) buildColor(f reflect.Value, d reflect.StructField, withLabel bool) {
	dComp := func() {
		imgui.SetNextItemWidth(-1)

		hsv := f.Interface().(*settings.HSV)

		r, g, b := color.HSVToRGB(float32(hsv.Hue), float32(hsv.Saturation), float32(hsv.Value))
		rgb := [3]float32{r, g, b}

		if imgui.ColorEdit3V(editor.getId(), &rgb, imgui.ColorEditFlagsHSV|imgui.ColorEditFlagsNoLabel|imgui.ColorEditFlagsFloat) {
			h, s, v := color.RGBToHSV(rgb[0], rgb[1], rgb[2])
			hsv.Hue = float64(h)
			hsv.Saturation = float64(s)
			hsv.Value = float64(v)
		}

		editor.comboOpened = editor.comboOpened || imgui.IsWindowFocusedV(imgui.FocusedFlagsChildWindows) && !imgui.IsWindowFocused()
	}

	if withLabel {
		editor.drawComponent(editor.getLabel(d), false, false, d, dComp)
	} else {
		dComp()
	}
}

func (editor *settingsEditor) drawComponent(label string, long, dynamic bool, d reflect.StructField, draw func()) {
	width := 240 + imgui.CalcTextSize("x", false, 0).X + imgui.CurrentStyle().FramePadding().X*4

	cCount := 1
	if !long {
		cCount = 2
	}

	if imgui.BeginTableV(editor.getId(), cCount, imgui.TableFlagsSizingStretchProp, imgui.Vec2{-1, 0}, -1) {
		imgui.TableSetupColumnV(editor.getId(), imgui.TableColumnFlagsWidthStretch, 0, uint(0))

		if dynamic {
			imgui.TableSetupColumnV(editor.getId(), imgui.TableColumnFlagsWidthFixed, 0, uint(1))
		} else if !long {
			imgui.TableSetupColumnV(editor.getId(), imgui.TableColumnFlagsWidthFixed, width, uint(1))
		}

		imgui.TableNextColumn()

		if t, ok := d.Tag.Lookup("tooltip"); ok {
			imgui.BeginGroup()
			imgui.Text(label)
			imgui.EndGroup()

			if imgui.IsItemHovered() {
				imgui.BeginTooltip()
				imgui.SetTooltip(t)
				imgui.EndTooltip()
			}
		} else {
			imgui.Text(label)
		}

		imgui.TableNextColumn()

		draw()

		imgui.EndTable()
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

func firstOf(args ...string) string {
	for _, arg := range args {
		if arg != "" {
			return arg
		}
	}

	return ""
}
