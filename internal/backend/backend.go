package backend

import (
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
type PipelineID uint32

type PipelineDesc struct {
}

type GPU interface {
	CreatePipeline(PipelineDesc) PipelineID
	ReleasePipeline(p PipelineID)
	BindPipeline(p PipelineID)

	CreateVertexBuffer([]types.PositionColorVertex) VertexBufferID
	PushVertexUniformData(u Mesh2dUniform)
	DrawVertexBuffer(vb VertexBufferID)
	ReleaseVertexBuffer(vb VertexBufferID)
}

type Input interface {
	GetKeyState(k types.KeyType) bool
	GetButtonState(b types.ButtonType) bool
}
