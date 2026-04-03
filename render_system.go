package engine

import (
	"fmt"

	"github.com/polouis/engine/types"
)

type Mesh2dComponent struct {
	VB  types.VertexBuffer
	Len uint
}

func NewMesh2d(ctx *Context, vertices []types.PositionColorVertex) Mesh2dComponent {
	return Mesh2dComponent{
		VB:  ctx.B.NewVertexBuffer(vertices),
		Len: uint(len(vertices)),
	}
}

var Mesh2dCID = RegisterComponent[Mesh2dComponent]()

func GetMesh2dComponents(w *World) *ComponentArray[Mesh2dComponent] {
	return w.Store(Mesh2dCID).(*ComponentArray[Mesh2dComponent])
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
		ctx.B.Draw(mesh2dCpnt.VB, uint32(mesh2dCpnt.Len))
	}
}

func ReleaseRenderSystem(ctx *Context) {
	for _, mesh2dCpnt := range GetMesh2dComponents(ctx.W).All() {
		ctx.B.Release(mesh2dCpnt.VB)
	}
}
