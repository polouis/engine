package engine

import (
	"fmt"

	backendsdl "github.com/polouis/engine/internal/backend_sdl"
	"github.com/polouis/engine/types"
)

type BackendType string

const (
	SDL   BackendType = "sdl"
	Dummy BackendType = "dummy"
)

func NewBackend(bt BackendType) types.Backend {
	switch bt {
	case SDL:
		return &backendsdl.BackendSDL{}
	case Dummy:
		return &backendDummy{}
	default:
		panic(fmt.Sprintf("Cannot instanciate unknown backend '%v'", bt))
	}
}
