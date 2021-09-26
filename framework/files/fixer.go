package files

import "strings"

var replacer = strings.NewReplacer("\\", "",
	"/", "",
	"<", "",
	">", "",
	"|", "",
	"?", "",
	"*", "",
	":", "",
	"\"", "")

func FixName(name string) string {
	return replacer.Replace(name)
}
