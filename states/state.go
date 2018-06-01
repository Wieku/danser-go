package states

type IState interface {
	Draw()
	Dispose()
}