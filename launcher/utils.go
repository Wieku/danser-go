package launcher

import (
	"fmt"
	"github.com/sqweek/dialog"
	"github.com/wieku/danser-go/framework/platform"
	"strings"
)

type messageType int

const (
	mInfo = messageType(iota)
	mError
	mQuestion
)

func showMessage(typ messageType, format string, args ...any) bool {
	message := fmt.Sprintf(format, args...)

	switch typ {
	case mInfo:
		dialog.Message(message).Info()
	case mError:
		if urlIndex := strings.Index(message, "http"); urlIndex > -1 {
			if dialog.Message(message + "\n\nDo you want to go there?").ErrorYesNo() {
				url := message[urlIndex:]
				platform.OpenURL(url)
				return true
			}
		} else {
			dialog.Message(message).Error()
		}
	case mQuestion:
		return dialog.Message(message).YesNo()
	}

	return false
}
