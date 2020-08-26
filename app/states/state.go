package states

type State interface {
	Show()
	Hide()
	Draw(delta float64)
	Dispose()
}
