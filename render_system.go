package engine

import (
	"github.com/polouis/engine/internal/backend"
	"github.com/polouis/engine/types"
)

type MeshComponent struct {
	VB backend.VertexBuffer
	// TODO use it when implementing shared buffer between multiple entities
	Len    uint32
	Offset uint32
}

func NewMeshComponent(ctx *Context, vertices []types.PositionColorVertex) MeshComponent {
	return MeshComponent{
		VB:     ctx.b.NewVertexBuffer(vertices),
		Len:    uint32(len(vertices)),
		Offset: 0,
	}
}

type SpriteComponent struct {
	x, y               int
	rotation           float32
	textureU, textureV float32
}

func UpdateRenderSystem(ctx *Context, deltatime uint64) {
	for e, mesh2dCpnt := range ctx.W.MeshStore.All() {
		transform, err := ctx.W.TransformStore.Get(e)
		var u backend.Mesh2dUniform
		if err == nil {
			u = backend.Mesh2dUniform{X: transform.Position.X, Y: transform.Position.Y}
		} else {
			u = backend.Mesh2dUniform{X: 0, Y: 0}
		}
		ctx.b.PushVertexUniformData(u)

		ctx.b.Draw(mesh2dCpnt.VB)
	}
}

func ReleaseRenderSystem(ctx *Context) {
	for _, mesh2dCpnt := range ctx.W.MeshStore.All() {
		ctx.b.Release(mesh2dCpnt.VB)
	}
}
