package settings

import "math"

type fileformat struct {
	General   *general   `icon:"\uF0AD"`
	Graphics  *graphics  `icon:"\uF108"`
	Audio     *audio     `icon:"\uF028"`
	Input     *input     `icon:"\uF11C"`
	Gameplay  *gameplay  `icon:"\uF140"`
	Skin      *skin      `icon:"\uF53F"`
	Cursor    *cursor    `icon:"\uF245"`
	Objects   *objects   `icon:"\uF1CD"`
	Playfield *playfield `icon:"\uF853"`
	Dance     *dance     `icon:"\uF5B7"`
	Knockout  *knockout  `icon:"\uF0CB"`
	Recording *recording `icon:"\uF03D"`
}

var DEBUG = false
var PLAY = false
var SKIP = false
var START = 0.0
var END = math.Inf(1)
var KNOCKOUT = false
var PLAYERS = 1
var DIVIDES = 2
var SPEED = 1.0
var PITCH = 1.0
var TAG = 1
var RECORD = false
var REPLAY = ""
