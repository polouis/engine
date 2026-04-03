package engine

import (
	"fmt"

	"github.com/polouis/engine/types"
)

type MeshComponent struct {
	VB types.VertexBuffer
	// TODO use it when implementing shared buffer between multiple entities
	Len    uint32
	Offset uint32
}

func NewMeshComponent(ctx *Context, vertices []types.PositionColorVertex) MeshComponent {
	return MeshComponent{
		VB:     ctx.B.NewVertexBuffer(vertices),
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

func LoadRenderSystem(ctx *Context) {
}

func UpdateRenderSystem(ctx *Context, deltatime uint64) {
	for e, velocityCpnt := range GetVelocityComponents(ctx.W).All() {
		fmt.Printf("Rendering entity %d with component %T\n", e, velocityCpnt)
	}

	for _, mesh2dCpnt := range GetMesh2dComponents(ctx.W).All() {
		ctx.B.Draw(mesh2dCpnt.VB)
	}
}

func ReleaseRenderSystem(ctx *Context) {
	for _, mesh2dCpnt := range GetMesh2dComponents(ctx.W).All() {
		ctx.B.Release(mesh2dCpnt.VB)
	}
}
