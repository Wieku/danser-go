package replay

import (
	"io/ioutil"
	"github.com/Mempler/rplpa"
)

func ExtractReplay(name string) *rplpa.Replay {
	buf, err := ioutil.ReadFile(name)
	if err != nil {
		panic(err)
	}
	replay, err := rplpa.ParseReplay(buf)
	if err != nil {
		panic(err)
	}
	return replay
}