package backend

import "github.com/polouis/engine/types"

type Mesh2dUniform struct {
	X float32
	Y float32
}

type Backend interface {
	Run(initCallback func(), updateCallback func(uint64), releaseCallback func()) error
	NewVertexBuffer([]types.PositionColorVertex) VertexBuffer
	GetKeyState(k types.KeyType) bool
	GetButtonState(b types.ButtonType) bool
	Draw(vb VertexBuffer)
	PushVertexUniformData(u Mesh2dUniform)
	Release(vb VertexBuffer)
}
type VertexBuffer interface{}
