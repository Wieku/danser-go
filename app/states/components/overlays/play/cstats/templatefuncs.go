package cstats

import (
	"fmt"
	"github.com/spf13/cast"
	"github.com/wieku/danser-go/app/utils"
	"github.com/wieku/danser-go/framework/math/mutils"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"math"
	"reflect"
	"strings"
	"text/template"
)

var templateFuncs = template.FuncMap{
	//conversion
	"format": func(format string, v ...any) string {
		return fmt.Sprintf(format, v)
	},
	"formatF": func(decimals int64, v any) string {
		return fmt.Sprintf("%.*f", int(decimals), cast.ToFloat64(v))
	},
	"formatF0": func(decimals int64, v any) string {
		return mutils.FormatWOZeros(cast.ToFloat64(v), int(decimals))
	},
	"formatC": func(v int64) string {
		return utils.Humanize(v)
	},
	"formatTime": tFormatTime,
	"int": func(v float64) int64 {
		return int64(math.Round(v))
	},
	"float": func(v int64) float64 {
		return float64(v)
	},

	//text
	"repeat":     func(count int64, str string) string { return strings.Repeat(str, int(count)) },
	"lower":      strings.ToLower,
	"upper":      strings.ToUpper,
	"hasPrefix":  strings.HasPrefix,
	"hasSuffix":  strings.HasSuffix,
	"split":      strings.Split,
	"contains":   func(substr string, str string) bool { return strings.Contains(str, substr) },
	"title":      func(a string) string { return cases.Title(language.English).String(a) },
	"trimSpace":  strings.TrimSpace,
	"trimSuffix": func(a, b string) string { return strings.TrimSuffix(b, a) },
	"trimPrefix": func(a, b string) string { return strings.TrimPrefix(b, a) },
	"padLeft":    padLeft,
	"padRight":   padRight,
	"join":       join,
	"addS":       func(a, b any) string { return cast.ToString(a) + cast.ToString(b) },

	//slices
	"list":   list,
	"append": push, "push": push,
	"slice": slice,

	//float math
	"add":       func(a, b any) float64 { return cast.ToFloat64(a) + cast.ToFloat64(b) },
	"sub":       func(a, b any) float64 { return cast.ToFloat64(a) - cast.ToFloat64(b) },
	"mul":       func(a, b any) float64 { return cast.ToFloat64(a) * cast.ToFloat64(b) },
	"div":       func(a, b any) float64 { return cast.ToFloat64(a) / cast.ToFloat64(b) },
	"pow":       func(a, b any) float64 { return math.Pow(cast.ToFloat64(a), cast.ToFloat64(b)) },
	"sqrt":      func(a any) float64 { return math.Sqrt(cast.ToFloat64(a)) },
	"round":     func(a any) float64 { return math.Round(cast.ToFloat64(a)) },
	"roundEven": func(a any) float64 { return math.RoundToEven(cast.ToFloat64(a)) },
	"floor":     func(a any) float64 { return math.Floor(cast.ToFloat64(a)) },
	"ceil":      func(a any) float64 { return math.Ceil(cast.ToFloat64(a)) },
	"abs":       func(a any) float64 { return mutils.Abs(cast.ToFloat64(a)) },
	"per":       func(a any) float64 { return cast.ToFloat64(a) * 100 },
	"maxf":      tMaxF,
	"minf":      tMinF,
	"clampf":    tClampF,

	//int math
	"addi":   func(a, b any) int64 { return cast.ToInt64(a) + cast.ToInt64(b) },
	"inc":    func(i interface{}) int64 { return cast.ToInt64(i) + 1 },
	"subi":   func(a, b any) int64 { return cast.ToInt64(a) - cast.ToInt64(b) },
	"muli":   func(a, b any) int64 { return cast.ToInt64(a) * cast.ToInt64(b) },
	"divi":   func(a, b any) int64 { return cast.ToInt64(a) / cast.ToInt64(b) },
	"modi":   func(a, b any) int64 { return cast.ToInt64(a) % cast.ToInt64(b) },
	"absi":   func(a any) int64 { return mutils.Abs(cast.ToInt64(a)) },
	"maxi":   tMaxI,
	"mini":   tMinI,
	"clampi": tClampI,
}

func tFormatTime(v any) (s string) {
	v1 := cast.ToInt64(v)

	if v1 < 0 {
		s += "-"
	}

	v1 = mutils.Abs(v1)

	return s + fmt.Sprintf("%d:%02d", v1/60, v1%60)
}

func tMaxF(a any, b ...any) float64 {
	ret := cast.ToFloat64(a)

	for _, v := range b {
		ret = max(ret, cast.ToFloat64(v))
	}

	return ret
}

func tMinF(a any, b ...any) float64 {
	ret := cast.ToFloat64(a)

	for _, v := range b {
		ret = min(ret, cast.ToFloat64(v))
	}

	return ret
}

func tMaxI(a any, b ...any) int64 {
	ret := cast.ToInt64(a)

	for _, v := range b {
		ret = max(ret, cast.ToInt64(v))
	}

	return ret
}

func tMinI(a any, b ...any) int64 {
	ret := cast.ToInt64(a)

	for _, v := range b {
		ret = min(ret, cast.ToInt64(v))
	}

	return ret
}

func tClampF(a, b, c any) float64 {
	return mutils.Clamp(cast.ToFloat64(a), cast.ToFloat64(b), cast.ToFloat64(c))
}

func tClampI(a, b, c any) int64 {
	return mutils.Clamp(cast.ToInt64(a), cast.ToInt64(b), cast.ToInt64(c))
}

func padRight(length any, str string) string {
	tLen := cast.ToInt(length)

	if len(str) >= tLen {
		return str
	}
	return str + strings.Repeat(" ", tLen-len(str))
}

func padLeft(length any, str string) string {
	tLen := cast.ToInt(length)

	if len(str) >= tLen {
		return str
	}
	return strings.Repeat(" ", tLen-len(str)) + str
}

func list(v ...interface{}) []interface{} {
	return v
}

func push(list interface{}, v interface{}) []interface{} {
	tp := reflect.TypeOf(list).Kind()
	switch tp {
	case reflect.Slice, reflect.Array:
		l2 := reflect.ValueOf(list)

		l := l2.Len()
		nl := make([]interface{}, l)
		for i := 0; i < l; i++ {
			nl[i] = l2.Index(i).Interface()
		}

		return append(nl, v)

	default:
		panic(fmt.Sprintf("Cannot push on type %s", tp))
	}
}

func slice(list interface{}, indices ...interface{}) interface{} {
	tp := reflect.TypeOf(list).Kind()
	switch tp {
	case reflect.Slice, reflect.Array:
		l2 := reflect.ValueOf(list)

		l := l2.Len()
		if l == 0 {
			return nil
		}

		var start, end int
		if len(indices) > 0 {
			start = cast.ToInt(indices[0])
		}
		if len(indices) < 2 {
			end = l
		} else {
			end = cast.ToInt(indices[1])
		}

		return l2.Slice(start, end).Interface()
	default:
		panic(fmt.Sprintf("list should be type of slice or array but %s", tp))
	}
}

func join(sep string, v interface{}) string {
	return strings.Join(strslice(v), sep)
}

func strslice(v interface{}) []string {
	switch v := v.(type) {
	case []string:
		return v
	case []interface{}:
		b := make([]string, 0, len(v))
		for _, s := range v {
			if s != nil {
				b = append(b, cast.ToString(s))
			}
		}
		return b
	default:
		val := reflect.ValueOf(v)
		switch val.Kind() {
		case reflect.Array, reflect.Slice:
			l := val.Len()
			b := make([]string, 0, l)
			for i := 0; i < l; i++ {
				value := val.Index(i).Interface()
				if value != nil {
					b = append(b, cast.ToString(value))
				}
			}
			return b
		default:
			if v == nil {
				return []string{}
			}

			return []string{cast.ToString(v)}
		}
	}
}
