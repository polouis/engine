package backenddummy

import (
	"fmt"

	"github.com/polouis/engine/internal/backend"
	"github.com/polouis/engine/types"
)

type BackendDummy struct{}

type DummyVertexBuffer struct{}

func (b *BackendDummy) Run(initCallback func(), updateCallback func(uint64), releaseCallback func()) error {
	fmt.Println("I'm a dummy backend")
	return nil
}

func (b *BackendDummy) NewVertexBuffer(vbData []types.PositionColorVertex) backend.VertexBuffer {
	return &DummyVertexBuffer{}
}

func (b *BackendDummy) Draw(vb backend.VertexBuffer) {
}

func (b *BackendDummy) Release(vb backend.VertexBuffer) {
}

func (b *BackendDummy) GetKeyState(k types.KeyType) bool {
	return false
}

func (b *BackendDummy) GetButtonState(btn types.ButtonType) bool {
	return false
}

func (b *BackendDummy) PushVertexUniformData(u backend.Mesh2dUniform) {

}
