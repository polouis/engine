package engine

import (
	"github.com/polouis/engine/internal/backend"
	"github.com/polouis/engine/types"
)

type MeshComponent struct {
	vb           backend.VertexBufferID
	material     types.MaterialID
	vertexFormat types.VertexFormat
	len          uint32
	offset       uint32
}

func NewMeshComponent(ctx *Context, vertices []types.PositionColorVertex, material types.MaterialID, vertexFormat types.VertexFormat) MeshComponent {
	vb := ctx.gpu.CreateVertexBuffer(vertexFormat.Layout().Stride, len(vertices))
	backend.UploadVertices(ctx.gpu, vb, vertices, uint32(len(vertices)))
	return MeshComponent{
		vb:           vb,
		material:     material,
		vertexFormat: vertexFormat,
		len:          uint32(len(vertices)),
		offset:       0,
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

	passDesc := backend.PassDesc{
		ClearColor: types.Color{R: 0.3, G: 0.03, B: 0.1, A: 1.0},
	}
	if ctx.gpu.BeginPass(passDesc) != nil {
		return
	}

	for e, mesh2dCpnt := range ctx.W.MeshStore.All() {
		material := ctx.RM.materialRessources.Get(mesh2dCpnt.material)
		pipeline := ctx.RM.pipelineFor(ctx, material, mesh2dCpnt.vertexFormat, &backend.PassDef{})

		transform, err := ctx.W.TransformStore.Get(e)
		var u mesh2dUniform
		if err == nil {
			u = mesh2dUniform{X: transform.Position.X, Y: transform.Position.Y}
		} else {
			u = mesh2dUniform{X: 0, Y: 0}
		}
		areanaOff, arenaLen := pushUniform(&ctx.r, u)

		ctx.r.cmds = append(ctx.r.cmds, backend.DrawCmd{
			Pipeline:     pipeline,
			Material:     mesh2dCpnt.material,
			VertexBuffer: mesh2dCpnt.vb,
			ArenaOffset:  areanaOff,
			ArenaLen:     arenaLen,
			VBOffset:     mesh2dCpnt.offset,
			VBLen:        mesh2dCpnt.len,
		})
	}

	ctx.r.Sort()
	ctx.gpu.Submit(ctx.r.cmds, ctx.r.arena)
	ctx.gpu.EndPass()
	ctx.gpu.SubmitCommandBuffer()
	ctx.r.Reset()
}

func ReleaseRenderSystem(ctx *Context) {
	for _, mesh2dCpnt := range ctx.W.MeshStore.All() {
		ctx.gpu.ReleaseVertexBuffer(mesh2dCpnt.vb)
	}
}
