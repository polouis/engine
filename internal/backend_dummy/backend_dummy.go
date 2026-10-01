package backenddummy

import (
	"fmt"

	"github.com/polouis/engine/internal/backend"
	"github.com/polouis/engine/types"
)

type BackendDummy struct {
	nextVB       backend.VertexBufferID
	nextPipeline backend.PipelineID
}

var _ backend.Platform = (*BackendDummy)(nil)
var _ backend.Input = (*BackendDummy)(nil)
var _ backend.GPU = (*BackendDummy)(nil)

func (b *BackendDummy) Run(initCallback func(), updateCallback func(uint64), releaseCallback func()) error {
	fmt.Println("I'm a dummy backend")
	return nil
}

func (b *BackendDummy) CreateVertexBuffer(vbData []types.PositionColorVertex) backend.VertexBufferID {
	b.nextVB++
	return b.nextVB
}

func (b *BackendDummy) DrawVertexBuffer(vb backend.VertexBufferID) {}

func (b *BackendDummy) ReleaseVertexBuffer(vb backend.VertexBufferID) {}

func (b *BackendDummy) GetKeyState(k types.KeyType) bool {
	return false
}

func (b *BackendDummy) GetButtonState(btn types.ButtonType) bool {
	return false
}

func (b *BackendDummy) PushVertexUniformData(u backend.Mesh2dUniform) {

}

func (b *BackendDummy) CreatePipeline(desc backend.PipelineDesc) backend.PipelineID {
	b.nextPipeline++
	return b.nextPipeline
}

func (b *BackendDummy) ReleasePipeline(p backend.PipelineID) {
}
