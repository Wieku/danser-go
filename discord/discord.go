package discord

import (
	"fmt"
	"github.com/wieku/danser-go/beatmap"
	"github.com/wieku/danser-go/settings"
	"github.com/wieku/rich-go/client"
	"log"
	"time"
)

const appId = ""

var startTime time.Time
var endTime time.Time
var mapString string

var queue chan func()

var connected bool

func Connect() {
	if !settings.General.DiscordPresenceOn {
		return
	}

	err := client.Login(appId)
	if err != nil {
		log.Println("Can't login to Discord RPC")
		return
	}

	connected = true

	queue = make(chan func())
	go func() {
		for {
			f, keepOpen := <-queue
			if keepOpen {
				f()
				continue
			}
			break
		}
	}()
}

func SetDuration(duration int64) {
	startTime = time.Now()
	endTime = time.Now().Add(time.Duration(duration) * time.Millisecond)
}

func SetMap(beatMap *beatmap.BeatMap) {
	mapString = fmt.Sprintf("%s - %s [%s]", beatMap.Artist, beatMap.Name, beatMap.Difficulty)
}

func UpdateKnockout(alive, players int) {
	if !connected {
		return
	}

	queue <- func() {
		err := client.SetActivity(client.Activity{
			State:      fmt.Sprintf("Watching knockout (%d of %d alive)", alive, players),
			Details:    mapString,
			LargeImage: "danser-logo",
			Timestamps: &client.Timestamps{
				Start: &startTime,
				End:   &endTime,
			},
		})
		if err != nil {
			log.Println("Can't send activity")
		}
	}
}

func UpdatePlay(name string) {
	if !connected {
		return
	}

	state := "Clicking circles"
	if name != "" {
		state = fmt.Sprintf("Watching %s", name)
	}

	queue <- func() {
		err := client.SetActivity(client.Activity{
			State:      state,
			Details:    mapString,
			LargeImage: "danser-logo",
			Timestamps: &client.Timestamps{
				Start: &startTime,
				End:   &endTime,
			},
		})
		if err != nil {
			log.Println("Can't send activity")
		}
	}
}

func UpdateDance(tag, divides int) {
	if !connected {
		return
	}

	statusText := "Watching "
	if tag > 1 {
		statusText += fmt.Sprintf("TAG%d ", tag)
	}

	if divides > 2 {
		statusText += "mandala"
	} else if divides == 2 {
		statusText += "mirror collage"
	} else {
		statusText += "cursor dance"
	}

	queue <- func() {
		err := client.SetActivity(client.Activity{
			State:      statusText,
			Details:    mapString,
			LargeImage: "danser-logo",
			Timestamps: &client.Timestamps{
				Start: &startTime,
				End:   &endTime,
			},
		})
		if err != nil {
			log.Println("Can't send activity")
		}
	}
}

func ClearActivity() {
	if !connected {
		return
	}

	queue <- func() {
		err := client.ClearActivity()
		if err != nil {
			log.Println("Can't clear activity")
		}
	}
}

func Disconnect() {
	if !connected {
		return
	}
	ClearActivity()
	connected = false
	close(queue)
	client.Logout()
}
