package settings

var Debug = initDebug()

func initDebug() *debug {
	return &debug{
		PerfGraph: &perfGraph{
			MaxRange: 1.2,
		},
	}
}

type debug struct {
	PerfGraph *perfGraph `label:"Performance Graph"`
}

type perfGraph struct {
	MaxRange float64 `label:"Max range (ms)" string:"true" min:"0.1" max:"10727"`
}
