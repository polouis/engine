package backend

import (
	"github.com/polouis/engine/internal"
	"github.com/polouis/engine/types"
)

type Platform interface {
	Run(initCallback func(), updateCallback func(uint64), releaseCallback func()) error
}

type VertexBufferID uint32
type PipelineID uint32

const InvalidPipeline = internal.InvalidHandle

type PipelineDesc struct {
}

type DrawCmd struct {
	Pipeline     PipelineID
	VertexBuffer VertexBufferID
	ArenaOffset  uint32
	ArenaLen     uint32
	// TODO implement VB sharing
	// VBIndex      uint32
	// VBLen        uint32
	// TODO implement render state sorting
	// SortKey      uint64 // pass | pipeline | material | depth
}

type PassDesc struct {
	ClearColor types.Color
}

type GPU interface {
	CreatePipeline(PipelineDesc) PipelineID
	ReleasePipeline(PipelineID)

	CreateVertexBuffer([]types.PositionColorVertex) VertexBufferID
	ReleaseVertexBuffer(VertexBufferID)

	BeginPass(PassDesc) error
	Submit(cmds []DrawCmd, frameArena []byte)
	EndPass()
	SubmitCommandBuffer()
}

type Input interface {
	GetKeyState(k types.KeyType) bool
	GetButtonState(b types.ButtonType) bool
}
