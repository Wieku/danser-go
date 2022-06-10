package launcher

import (
	"fmt"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/math/mutils"
	"golang.org/x/exp/constraints"
	"strconv"
	"strings"
)

type popupType int

const (
	popDynamic = popupType(iota)
	popMedium
	popBig
)

type iPopup interface {
	draw()
	shouldClose() bool
	open()
}

type popup struct {
	internalDraw func()

	name string

	popType popupType

	opened bool
}

func newPopup(name string, popType popupType) *popup {
	return &popup{
		name:    name,
		popType: popType,
	}
}

func newPopupF(name string, popType popupType, draw func()) *popup {
	return &popup{
		name:         name,
		popType:      popType,
		internalDraw: draw,
	}
}

func (p *popup) draw() {
	p.opened = true
	switch p.popType {
	case popDynamic, popMedium:
		popupSmall(p.name, &p.opened, p.popType == popDynamic, p.internalDraw)
	case popBig:
		popupWide(p.name, &p.opened, p.internalDraw)
	}
}

func (p *popup) shouldClose() bool {
	return !p.opened
}

func (p *popup) open() {
	p.opened = true
}

func sliderC(label string, val *floatParam, min, max float32, format string) {
	imgui.Text(label + ":")

	imgui.PushFont(Font16)

	if imgui.BeginTableV("rt"+label, 2, imgui.TableFlagsSizingStretchProp, imgui.Vec2{-1, 0}, -1) {
		imgui.TableNextColumn()

		imgui.SetNextItemWidth(-1)

		if sliderFloatSlide("##"+label, &val.value, min, max, format, imgui.SliderFlagsNoInput) {
			if math32.Abs(val.value-val.ogValue) > 0.001 {
				val.changed = true
			} else {
				val.changed = false
				val.value = val.ogValue
			}
		}

		imgui.TableNextColumn()

		if imgui.Button("Reset##" + label) {
			val.value = val.ogValue
			val.changed = false
		}

		imgui.EndTable()
	}

	imgui.PopFont()
}

func sliderCStep(label string, val *floatParam, min, max, step float32, format string) {
	imgui.Text(label + ":")

	imgui.PushFont(Font16)

	if imgui.BeginTableV("rt"+label, 2, imgui.TableFlagsSizingStretchProp, imgui.Vec2{-1, 0}, -1) {
		imgui.TableNextColumn()

		imgui.SetNextItemWidth(-1)

		if sliderFloatStep("##"+label, &val.value, min, max, step, format) {
			if math32.Abs(val.value-val.ogValue) > 0.001 {
				val.changed = true
			} else {
				val.changed = false
				val.value = val.ogValue
			}
		}

		imgui.TableNextColumn()

		if imgui.Button("Reset##" + label) {
			val.value = val.ogValue
			val.changed = false
		}

		imgui.EndTable()
	}

	imgui.PopFont()
}

func sliderFloatStep(label string, val *float32, min, max, step float32, format string) bool {
	stepNum := int32((max - min) / step)

	v := int32(math32.Round((*val - min) / step))

	cPos := imgui.CursorPos()
	iW := imgui.CalcItemWidth() + imgui.CurrentStyle().FramePadding().X*2

	ret := sliderIntSlide(label, &v, 0, stepNum, "##%d", imgui.SliderFlagsNoInput)

	cPos2 := imgui.CursorPos()

	*val = (float32(v) * step) + min

	tx := fmt.Sprintf(format, *val)

	tS := imgui.CalcTextSize(tx+"f", false, 0)

	imgui.SetCursorPos(imgui.Vec2{
		X: cPos.X + (iW-tS.X)/2,
		Y: cPos.Y,
	})

	imgui.AlignTextToFramePadding()

	imgui.Text(tx)

	imgui.SetCursorPos(cPos2)

	return ret
}

func sliderIC(label string, val *intParam, min, max int32, format string) {
	imgui.Text(label + ":")

	imgui.PushFont(Font16)

	if imgui.BeginTableV("rt"+label, 2, imgui.TableFlagsSizingStretchProp, imgui.Vec2{-1, 0}, -1) {
		imgui.TableNextColumn()

		imgui.SetNextItemWidth(-1)

		if sliderIntSlide("##"+label, &val.value, min, max, format, imgui.SliderFlagsNoInput) {
			val.changed = val.value != val.ogValue
		}

		imgui.TableNextColumn()

		if imgui.Button("Reset##" + label) {
			val.value = val.ogValue
			val.changed = false
		}

		imgui.EndTable()
	}

	imgui.PopFont()
}

func popupSmall(name string, opened *bool, dynamicSize bool, content func()) {
	wSize := imgui.WindowSize()

	sX := wSize.X / 2
	if dynamicSize {
		sX = 0
	}

	popupInter(name, opened, imgui.Vec2{X: sX, Y: 0}, content)
}

func popupWide(name string, opened *bool, content func()) {
	wSize := imgui.WindowSize()
	popupInter(name, opened, imgui.Vec2{X: wSize.X * 0.9, Y: wSize.Y * 0.9}, content)
}

func popupInter(name string, opened *bool, size imgui.Vec2, content func()) {
	wSize := imgui.WindowSize()

	if *opened {
		if !imgui.IsPopupOpen("##" + name) {
			imgui.OpenPopup("##" + name)
		}

		imgui.SetNextWindowSize(size)

		imgui.SetNextWindowPosV(imgui.Vec2{
			X: wSize.X / 2,
			Y: wSize.Y / 2,
		}, imgui.ConditionAppearing, imgui.Vec2{
			X: 0.5,
			Y: 0.5,
		})

		if imgui.BeginPopupModalV("##"+name, opened, imgui.WindowFlagsNoCollapse|imgui.WindowFlagsNoResize|imgui.WindowFlagsAlwaysAutoResize|imgui.WindowFlagsNoMove|imgui.WindowFlagsNoTitleBar) {
			content()

			if (imgui.IsMouseClicked(0) || imgui.IsMouseClicked(1)) && !imgui.IsWindowHoveredV(imgui.HoveredFlagsRootAndChildWindows|imgui.HoveredFlagsAllowWhenBlockedByActiveItem|imgui.HoveredFlagsAllowWhenBlockedByPopup) {
				*opened = false
			}

			imgui.EndPopup()
		}
	}
}

func sliderIntSlide(label string, value *int32, min, max int32, format string, flags imgui.SliderFlags) (ret bool) {
	ret = imgui.SliderIntV(label, value, min, max, format, flags)

	if imgui.IsItemHovered() || imgui.IsItemActive() {
		ret = ret || keySlideInt(value, min, max)
	}

	return
}

func sliderFloatSlide(label string, value *float32, min, max float32, format string, flags imgui.SliderFlags) (ret bool) {
	ret = imgui.SliderFloatV(label, value, min, max, format, flags)

	if imgui.IsItemHovered() || imgui.IsItemActive() {
		iDot := strings.Index(format, ".")
		iF := strings.Index(format, "f")
		prec, _ := strconv.Atoi(format[iDot+1 : iF])
		step := math32.Pow(10, -float32(prec))

		ret = ret || keySlideFloat(value, min, max, step)
	}

	return
}

func keySlideInt[T constraints.Integer](value *T, min, max T) (ret bool) {
	if imgui.IsKeyPressed(imgui.KeyLeftArrow) {
		*value = mutils.Clamp(*value-1, min, max)
		ret = true
	}

	if imgui.IsKeyPressed(imgui.KeyRightArrow) {
		*value = mutils.Clamp(*value+1, min, max)
		ret = true
	}

	return
}

func keySlideFloat[T constraints.Float](value *T, min, max, step T) (ret bool) {
	if imgui.IsKeyPressed(imgui.KeyLeftArrow) {
		*value = mutils.ClampF(*value-step, min, max)
		ret = true
	}

	if imgui.IsKeyPressed(imgui.KeyRightArrow) {
		*value = mutils.ClampF(*value+step, min, max)
		ret = true
	}

	return
}
