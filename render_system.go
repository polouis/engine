package engine

import (
	"github.com/polouis/engine/internal/backend"
	"github.com/polouis/engine/types"
)

type MeshComponent struct {
	vb       backend.VertexBufferID
	pipeline backend.PipelineID
	// TODO use it when implementing shared buffer between multiple entities
	len    uint32
	offset uint32
}

func NewMeshComponent(ctx *Context, vertices []types.PositionColorVertex, pipeline backend.PipelineID) MeshComponent {
	return MeshComponent{
		vb:       ctx.gpu.CreateVertexBuffer(vertices),
		pipeline: pipeline,
		len:      uint32(len(vertices)),
		offset:   0,
	}
}

type mesh2dUniform struct {
	X float32
	Y float32
}

type SpriteComponent struct {
	x, y               int
	rotation           float32
	textureU, textureV float32
}

func UpdateRenderSystem(ctx *Context, deltatime uint64) {

	for e, mesh2dCpnt := range ctx.W.MeshStore.All() {
		transform, err := ctx.W.TransformStore.Get(e)
		var u mesh2dUniform
		if err == nil {
			u = mesh2dUniform{X: transform.Position.X, Y: transform.Position.Y}
		} else {
			u = mesh2dUniform{X: 0, Y: 0}
		}
		areanaOff, arenaLen := pushUniform(&ctx.r, u)

		ctx.r.cmds = append(ctx.r.cmds, backend.DrawCmd{
			Pipeline:     mesh2dCpnt.pipeline,
			VertexBuffer: mesh2dCpnt.vb,
			ArenaOffset:  areanaOff,
			ArenaLen:     arenaLen,
		})
	}

	ctx.r.Sort()

	ctx.gpu.Submit(ctx.r.cmds, ctx.r.arena)

	ctx.r.Reset()
}

func ReleaseRenderSystem(ctx *Context) {
	for _, mesh2dCpnt := range ctx.W.MeshStore.All() {
		ctx.gpu.ReleaseVertexBuffer(mesh2dCpnt.vb)
	}
}
