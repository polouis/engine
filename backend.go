package engine

import "fmt"

type BackendType string

const (
	SDL   BackendType = "sdl"
	Dummy BackendType = "dummy"
)

type Backend interface {
	Run(*World)
}

func NewBackend(s string) Backend {
	switch s {
	case "sdl":
		return &backendSDL{}
	case "dummy":
		return &backendDummy{}
	default:
		panic(fmt.Sprintf("Cannot instanciate unknown backend '%s'", s))
	}
}
