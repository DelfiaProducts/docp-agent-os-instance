package interfaces

type IManager interface {
	Start() error
}

type IAgent interface {
	Start() error
}

type IUpdater interface {
	Start() error
}
