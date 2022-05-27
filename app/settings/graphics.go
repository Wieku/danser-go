package settings

var Graphics = initGraphics()

func initGraphics() *graphics {
	return &graphics{
		Width:        1920,
		Height:       1080,
		WindowWidth:  1280,
		WindowHeight: 720,
		Fullscreen:   true,
		VSync:        false,
		FPSCap:       0,
		MSAA:         0,
		ShowFPS:      true,
		Experimental: &experimental{
			UsePersistentBuffers: false,
		},
	}
}

type graphics struct {
	fSize        string `vector:"true" label:"Fullscreen resolution" left:"Width" right:"Height"`
	Width        int64  `min:"1" max:"30720"`
	Height       int64  `min:"1" max:"17280"`
	wSize        string `vector:"true" label:"Windowed resolution" left:"WindowWidth" right:"WindowHeight"`
	WindowWidth  int64  `min:"1" max:"30720"`
	WindowHeight int64  `min:"1" max:"17280"`
	Fullscreen   bool   //true
	VSync        bool   `label:"Vertical Sync"`                //false
	FPSCap       int64  `label:"FPS limit" min:"0" max:"1000"` //1000
	MSAA         int32  `combo:"0|OFF,2|2x,4|4x,8|8x,16|16x"`  //16
	ShowFPS      bool
	Experimental *experimental
}

type experimental struct {
	// Should persistent buffer be used in main QuadBatch. Uses more VRAM, but for high-end gpus may give a little fps boost
	UsePersistentBuffers bool
}

func (gr *graphics) SetDefaults(width, height int64) {
	gr.Width = width
	gr.Height = height
	gr.WindowWidth = width * 3 / 4
	gr.WindowHeight = height * 3 / 4
}

func (gr graphics) GetSize() (int64, int64) {
	if gr.Fullscreen {
		return gr.Width, gr.Height
	}
	return gr.WindowWidth, gr.WindowHeight
}

func (gr graphics) GetSizeF() (float64, float64) {
	if gr.Fullscreen {
		return float64(gr.Width), float64(gr.Height)
	}
	return float64(gr.WindowWidth), float64(gr.WindowHeight)
}

func (gr graphics) GetWidth() int64 {
	if gr.Fullscreen {
		return gr.Width
	}
	return gr.WindowWidth
}

func (gr graphics) GetWidthF() float64 {
	if gr.Fullscreen {
		return float64(gr.Width)
	}
	return float64(gr.WindowWidth)
}

func (gr graphics) GetHeight() int64 {
	if gr.Fullscreen {
		return gr.Height
	}
	return gr.WindowHeight
}

func (gr graphics) GetHeightF() float64 {
	if gr.Fullscreen {
		return float64(gr.Height)
	}
	return float64(gr.WindowHeight)
}

func (gr graphics) GetAspectRatio() float64 {
	if gr.Fullscreen {
		return float64(gr.Width) / float64(gr.Height)
	}
	return float64(gr.WindowWidth) / float64(gr.WindowHeight)
}
