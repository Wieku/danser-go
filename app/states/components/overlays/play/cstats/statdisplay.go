package cstats

import (
	"bytes"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/font"
	color2 "github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/vector"
	"log"
	"maps"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"text/template"
)

type Stat struct {
	display *StatDisplay
	engine  *template.Template

	compiledTemplate *template.Template

	usedStats []string

	pastValues map[string]any

	config *settings.Statistic

	generatedText []string

	metrics vector.Vector2d

	anchor, origin, textOrigin vector.Vector2d

	target map[string]any
	dirty  bool
	mtx    *sync.Mutex

	errorCounter int
}

func NewStat(display *StatDisplay, statSetting *settings.Statistic, templateEngine *template.Template) *Stat {
	statDisp := &Stat{
		display:    display,
		engine:     templateEngine,
		config:     statSetting,
		anchor:     vector.ParseOrigin(statSetting.Anchor).AddS(1, 1).Scl(0.5),
		origin:     vector.ParseOrigin(statSetting.Align).AddS(1, 1).Mult(vector.NewVec2d(0, 0.5)),
		pastValues: make(map[string]any),
		target:     make(map[string]any),
		dirty:      false,
		mtx:        &sync.Mutex{},
	}

	statDisp.textOrigin = vector.NewVec2d(vector.ParseOrigin(statSetting.Align).X, -1)

	if statDisp.compile() {
		return statDisp
	}

	return nil
}

func (stat *Stat) compile() bool {
	var err error

	stat.compiledTemplate, err = stat.engine.Parse(stat.config.Template)

	if err != nil {
		log.Println("Failed to compile template:", err.Error())
		return false
	}

	compiled, err2 := regexp.Compile("\\W[.]\\w+")
	if err2 != nil {
		log.Println("Failed to compile regular expression:", err2.Error())
		return false
	}

	flds := compiled.FindAllString(stat.config.Template, -1)

	if flds != nil && len(flds) > 0 {
		for _, fld := range flds {
			extracted := fld[2:]

			stat.usedStats = append(stat.usedStats, extracted)

			if strings.Contains(extracted, "Roll") {
				spl := strings.Split(extracted, "Roll")

				if len(spl) == 1 {
					log.Println("Missing decimals number for rolling value:", extracted)
					return false
				}

				decimals, errP := strconv.Atoi(strings.TrimSuffix(spl[1], "S"))
				if errP != nil {
					log.Println("Failed to parse deicmals:", errP.Error())
					return false
				}

				stat.display.registerRollingValue(spl[0], decimals, strings.HasSuffix(spl[1], "S"))
			}
		}
	}

	stat.refreshText()

	return true
}

func (stat *Stat) Update(valMap map[string]any) {
	shouldUpdate := false

	for _, field := range stat.usedStats {
		v, ok := valMap[field]

		if !ok {
			continue
		}

		comp := reflect.ValueOf(v).Comparable()

		if (comp && v != stat.pastValues[field]) || (!comp && !reflect.DeepEqual(v, stat.pastValues[field])) {
			stat.pastValues[field] = v
			shouldUpdate = true
		}
	}

	if shouldUpdate {
		stat.mtx.Lock()
		maps.Copy(stat.target, stat.pastValues)
		stat.dirty = true

		stat.mtx.Unlock()
	}
}

func (stat *Stat) refreshText() {
	var doc bytes.Buffer

	err := stat.compiledTemplate.Execute(&doc, stat.target)
	if err != nil && stat.errorCounter < 10 {
		log.Println("Failed to render template:", err.Error())
		stat.errorCounter++
	}

	fnt := font.GetFont("HUDFont")

	fntSize := stat.config.Size

	stat.generatedText = strings.Split(doc.String(), "\n")

	stat.metrics = vector.NewVec2d(0, fntSize*float64(len(stat.generatedText)))

	for i := range stat.generatedText {
		stat.metrics.X = max(stat.metrics.X, fnt.GetWidthMonospaced(fntSize, stat.generatedText[i]))
	}
}

func (stat *Stat) Draw(batch *batch.QuadBatch, alpha float64, sclWidth, sclHeight float64) {
	stat.mtx.Lock()

	if stat.dirty {
		stat.refreshText()

		stat.dirty = false
	}

	stat.mtx.Unlock()

	batch.ResetTransform()

	ppAlpha := stat.config.Opacity * alpha

	if ppAlpha < 0.001 || !stat.config.Show {
		return
	}

	fnt := font.GetFont("HUDFont")
	fntSize := stat.config.Size

	position := vector.NewVec2d(sclWidth, sclHeight).Mult(stat.anchor).AddS(stat.config.XPosition, stat.config.YPosition).Sub(stat.metrics.Mult(stat.origin))

	cS := stat.config.Color
	color := color2.NewHSVA(float32(cS.Hue), float32(cS.Saturation), float32(cS.Value), float32(ppAlpha))

	for i := 0; i < len(stat.generatedText); i++ {
		subPos := position.AddS(0, float64(i)*fntSize)

		batch.SetColor(0, 0, 0, ppAlpha*0.8)
		fnt.DrawOriginV(batch, subPos.AddS(0, 1), stat.textOrigin, fntSize, true, stat.generatedText[i])

		batch.SetColorM(color)
		fnt.DrawOriginV(batch, subPos, stat.textOrigin, fntSize, true, stat.generatedText[i])
	}

	batch.ResetTransform()
}
