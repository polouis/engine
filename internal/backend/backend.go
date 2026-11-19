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
type PassID uint32
type ShaderID uint32

const InvalidPipeline = internal.InvalidHandle

type PipelineDesc struct {
	VertexShader   ShaderID
	FragmentShader ShaderID
	VertexFormat   types.VertexFormat
	// PrimitiveType  types.PrimitiveType
	// the magic 0, 1, 0, 0 at pipeline.go:18 — part of the shader's interface
	// VertexUniformBuffers uint32
	// FragmentSamplers     uint32
	// BlendMode types.BlendMode
	// DepthTest bool
	// CullMode  types.CullMode
}

type PassDef struct {
	id PassID
	// targetFormat types.TextureFormat
	fillMode types.FillMode
	// depthTest    types.CompareOp
	// Non-zero: use these shaders instead of the material's.
	EffectOverride types.EffectID
}

type DrawCmd struct {
	Pipeline     PipelineID
	Material     types.MaterialID
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
	CreateShader(name string) ShaderID
	ReleaseShader(ShaderID)

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
