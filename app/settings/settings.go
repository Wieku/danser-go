package settings

type beat struct {
	BeatScale float64 //1.2
}

type fileformat struct {
	General   *general
	Graphics  *graphics
	Audio     *audio
	Input     *input
	Cursor    *cursor
	Objects   *objects
	Playfield *playfield
	Dance     *dance
	Knockout  *knockout
}

var Beat = &beat{1.2}

var DEBUG = false
var PLAY = false
var KNOCKOUT = false
var PLAYERS = 1
var DIVIDES = 2
var SPEED = 1.0
var PITCH = 1.0
var TAG = 1
