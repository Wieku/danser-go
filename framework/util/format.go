package util

import (
	"fmt"
	"strconv"
)

func FormatSeconds(seconds int) (ret string) {
	hFormat := "%dh"
	mFormat := "%dm"
	sFormat := "%ds"

	if days := seconds / (3600 * 24); days > 0 {
		ret += strconv.Itoa(days) + "d"
		hFormat = "%02dh"
	}

	if hours := (seconds / 3600) % 24; hours > 0 {
		ret += fmt.Sprintf(hFormat, hours)
		mFormat = "%02dm"
	}

	if minutes := (seconds / 60) % 60; minutes > 0 {
		ret += fmt.Sprintf(mFormat, minutes)
		sFormat = "%02ds"
	}

	ret += fmt.Sprintf(sFormat, seconds%60)

	return
}
