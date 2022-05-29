package discord

import (
	"fmt"
	"github.com/nattawitc/rich-go/client"
	"github.com/wieku/danser-go/app/graphics"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/build"
	"github.com/wieku/danser-go/framework/goroutines"
	"log"
	"sync"
	"time"
)

const appId = "658093518396588032"

var queue chan func()
var connected bool
var endSync *sync.WaitGroup

var startTime = time.Now()
var endTime = time.Now()
var mapString string

var lastSentActivity = "Idle"

func Connect() {
	if !settings.General.DiscordPresenceOn {
		return
	}

	log.Println("Trying to connect to Discord RPC...")

	err := client.Login(appId)
	if err != nil {
		log.Println("Can't login to Discord RPC! Error:", err.Error())
		return
	}

	log.Println("Connected!")

	connected = true

	queue = make(chan func(), 100)

	endSync = &sync.WaitGroup{}

	endSync.Add(1)

	goroutines.Run(func() {
		for {
			f, keepOpen := <-queue

			if f != nil {
				f()
			}

			if !keepOpen {
				endSync.Done()
				break
			}
		}
	})
}

func SetDuration(duration int64) {
	startTime = time.Now()
	endTime = time.Now().Add(time.Duration(duration) * time.Millisecond)

	sendActivity(lastSentActivity)
}

func SetMap(artist, title, version string) {
	mapString = fmt.Sprintf("%s - %s [%s]", artist, title, version)

	sendActivity(lastSentActivity)
}

func sendActivity(state string) {
	if !connected {
		return
	}

	queue <- func() {
		lastSentActivity = state
		err := client.SetActivity(client.Activity{
			State:      state,
			Details:    mapString,
			LargeImage: "danser-logo",
			LargeText:  "danser-go " + build.VERSION,
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

func UpdateKnockout(alive, players int) {
	sendActivity(fmt.Sprintf("Watching knockout (%d of %d alive)", alive, players))
}

func UpdatePlay(cursor *graphics.Cursor) {
	state := "Clicking circles"
	if !cursor.IsPlayer || cursor.IsAutoplay {
		state = fmt.Sprintf("Watching %s", cursor.Name)
	}

	sendActivity(state)
}

func UpdateDance(tag, divides int) {
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

	sendActivity(statusText)
}

func ClearActivity() {
	if !connected {
		return
	}

	queue <- func() {
		lastSentActivity = "Idle"

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

	connected = false

	close(queue)

	endSync.Wait()

	err := client.ClearActivity()
	if err != nil {
		log.Println("Can't clear activity")
	}

	client.Logout()
}
