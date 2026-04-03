package engine

import (
	"fmt"

	"github.com/polouis/engine/types"
)

type backendDummy struct{}

type DummyVertexBuffer struct{}

func (b *backendDummy) Run(initCallback func(), updateCallback func(uint64), releaseCallback func()) error {
	fmt.Println("I'm a dummy backend")
	return nil
}

func (b *backendDummy) NewVertexBuffer(vbData []types.PositionColorVertex) types.VertexBuffer {
	return &DummyVertexBuffer{}
}

func (b *backendDummy) Draw(vb types.VertexBuffer) {
}

func (b *backendDummy) Release(vb types.VertexBuffer) {
}
