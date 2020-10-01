package settings

type fileformat struct {
	General   *general
	Graphics  *graphics
	Audio     *audio
	Input     *input
	Skin      *skin
	Cursor    *cursor
	Objects   *objects
	Playfield *playfield
	Dance     *dance
	Knockout  *knockout
}

var DEBUG = false
var PLAY = false
var KNOCKOUT = false
var PLAYERS = 1
var DIVIDES = 2
var SPEED = 1.0
var PITCH = 1.0
var TAG = 1
