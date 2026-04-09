package engine

import (
	"fmt"

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

var MeshCID = RegisterComponent[MeshComponent]()

func GetMesh2dComponents(w *World) *ComponentArray[MeshComponent] {
	return w.Store(MeshCID).(*ComponentArray[MeshComponent])
}

type SpriteComponent struct {
	x, y               int
	rotation           float32
	textureU, textureV float32
}

var SpriteCID = RegisterComponent[SpriteComponent]()

func GetSpriteComponents(w *World) *ComponentArray[SpriteComponent] {
	return w.Store(VelocityCID).(*ComponentArray[SpriteComponent])
}

func UpdateRenderSystem(ctx *Context, deltatime uint64) {
	for e, velocityCpnt := range GetVelocityComponents(ctx.W).All() {
		fmt.Printf("Rendering entity %d with component %T\n", e, velocityCpnt)
	}

	for e, mesh2dCpnt := range GetMesh2dComponents(ctx.W).All() {
		transform, err := GetTransformComponents(ctx.W).Get(e)
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
	for _, mesh2dCpnt := range GetMesh2dComponents(ctx.W).All() {
		ctx.b.Release(mesh2dCpnt.VB)
	}
}
