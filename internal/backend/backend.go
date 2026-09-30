package backend

import (
	"github.com/polouis/engine/internal"
	"github.com/polouis/engine/types"
)

type Mesh2dUniform struct {
	X float32
	Y float32
}

type Platform interface {
	Run(initCallback func(), updateCallback func(uint64), releaseCallback func()) error
}

type VertexBufferID uint32

const InvalidBuffer VertexBufferID = internal.InvalidHandle

type GPU interface {
	NewVertexBuffer([]types.PositionColorVertex) VertexBufferID
	PushVertexUniformData(u Mesh2dUniform)
	Draw(vb VertexBufferID)
	Release(vb VertexBufferID)
}

type Input interface {
	GetKeyState(k types.KeyType) bool
	GetButtonState(b types.ButtonType) bool
}
