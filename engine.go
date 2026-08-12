package engine

import (
	"fmt"

	"github.com/polouis/engine/internal/backend"
	backenddummy "github.com/polouis/engine/internal/backend_dummy"
	backendsdl "github.com/polouis/engine/internal/backend_sdl"
	"github.com/polouis/engine/types"
)

type Context struct {
	W   *World
	RM  *RessourceManager
	p   backend.Platform
	gpu backend.GPU
	i   backend.Input
}

func New(bt types.BackendType) *Context {
	switch bt {
	case types.SDL:
		bSDL := &backendsdl.BackendSDL{}
		return &Context{W: NewWorld(), RM: NewRessourceManager(), p: bSDL, gpu: bSDL, i: bSDL}
	case types.Dummy:
		bDummy := &backenddummy.BackendDummy{}
		return &Context{W: NewWorld(), RM: NewRessourceManager(), p: bDummy, gpu: bDummy, i: bDummy}
	default:
		panic(fmt.Sprintf("Cannot instanciate unknown backend '%v'", bt))
	}
}

func initCtxBindingCallback(ctx *Context, initCallback func(*Context)) func() {
	return func() {
		initCallback(ctx)
	}
}

func updateCtxBindingCallback(ctx *Context, updateCallback func(*Context, uint64)) func(uint64) {
	return func(delta uint64) {
		updateCallback(ctx, delta)
	}
}

func releaseCtxBindingCallback(ctx *Context, releaseCallback func(*Context)) func() {
	return func() {
		releaseCallback(ctx)
	}
}

func Run(ctx *Context, initCallback func(*Context), updateCallback func(*Context, uint64), releaseCallback func(*Context)) error {
	return ctx.p.Run(
		initCtxBindingCallback(ctx, initCallback),
		updateCtxBindingCallback(ctx, updateCallback),
		releaseCtxBindingCallback(ctx, releaseCallback),
	)
}

func GetKeyState(ctx *Context, k types.KeyType) bool {
	return ctx.i.GetKeyState(k)
}

func GetButtonState(ctx *Context, b types.ButtonType) bool {
	return ctx.i.GetButtonState(b)
}
