package backend

import "github.com/polouis/engine/types"

type Backend interface {
	Run(initCallback func(), updateCallback func(uint64), releaseCallback func()) error
	NewVertexBuffer([]types.PositionColorVertex) VertexBuffer
	GetKeyState(k types.KeyType) bool
	GetButtonState(b types.ButtonType) bool
	Draw(vb VertexBuffer)
	Release(vb VertexBuffer)
}
type VertexBuffer interface{}
