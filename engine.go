package engine

import (
	"fmt"

	"github.com/polouis/engine/internal/backend"
	backenddummy "github.com/polouis/engine/internal/backend_dummy"
	backendsdl "github.com/polouis/engine/internal/backend_sdl"
	"github.com/polouis/engine/types"
)

type Context struct {
	W *World
	b backend.Backend
}

func New(bt types.BackendType) *Context {
	switch bt {
	case types.SDL:
		return &Context{W: NewWorld(), b: &backendsdl.BackendSDL{}}
	case types.Dummy:
		return &Context{W: NewWorld(), b: &backenddummy.BackendDummy{}}
	default:
		panic(fmt.Sprintf("Cannot instanciate unknown backend '%v'", bt))
	}
}

func Run(ctx *Context, initCallback func(), updateCallback func(uint64), releaseCallback func()) error {
	return ctx.b.Run(initCallback, updateCallback, releaseCallback)
}

func GetKeyState(ctx *Context, k types.KeyType) bool {
	return ctx.b.GetKeyState(k)
}

func GetButtonState(ctx *Context, b types.ButtonType) bool {
	return ctx.b.GetButtonState(b)
}

func HandleInput(ctx *Context, deltaTime uint64) Command {
	var x, y float32 = 0, 0
	if GetKeyState(ctx, types.Up) || GetButtonState(ctx, types.ButtonUp) {
		y = .5 / 1e9 * float32(deltaTime)
	}
	if GetKeyState(ctx, types.Down) || GetButtonState(ctx, types.ButtonDown) {
		y = -.5 / 1e9 * float32(deltaTime)
	}
	if GetKeyState(ctx, types.Left) || GetButtonState(ctx, types.ButtonLeft) {
		x = -.5 / 1e9 * float32(deltaTime)
	}
	if GetKeyState(ctx, types.Right) || GetButtonState(ctx, types.ButtonRight) {
		x = .5 / 1e9 * float32(deltaTime)
	}

	if x != 0 || y != 0 {
		return NewMoveCommand(x, y, 0)
	} else {
		return nil
	}
}

type Command interface {
	Execute(ctx *Context, e EntityID)
}
