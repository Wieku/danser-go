package hitjudge

import (
	"danser/score"
	"github.com/flesnuk/oppai5"
)

type TotalResult struct {
	N300   	uint16
	N100   	uint16
	N50    	uint16
	Misses 	uint16
	Combo  	uint16
	Mods   	uint32
	Acc		float64
	Rank    score.Rank
	PP 		oppai.PPv2
}
