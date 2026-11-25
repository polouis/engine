package engine

import (
	"slices"
	"testing"
	"unsafe"

	"github.com/polouis/engine/internal/backend"
	backenddummy "github.com/polouis/engine/internal/backend_dummy"
	"github.com/polouis/engine/types"
)

// uniformSize is the arena footprint of one per-draw uniform. Derived from the
// type so that adding a field to mesh2dUniform does not break this test.
const uniformSize = uint32(unsafe.Sizeof(mesh2dUniform{}))

func mustLoadEffect(t *testing.T, ctx *Context, json string) {
	t.Helper()
	if err := ctx.RM.LoadEffect(ctx, []byte(json)); err != nil {
		t.Fatalf("LoadEffect: %v", err)
	}
}

// mustLoadMaterial loads a material and returns its id, so a typo in a name or an
// effect reference fails here rather than surfacing later as a confusing
// pipeline mismatch.
func mustLoadMaterial(t *testing.T, ctx *Context, name, json string) types.MaterialID {
	t.Helper()
	if err := ctx.RM.LoadMaterial(ctx, []byte(json)); err != nil {
		t.Fatalf("LoadMaterial(%s): %v", name, err)
	}
	id, ok := ctx.RM.materialRessources.ID(name)
	if !ok {
		t.Fatalf("material %s is not registered after loading it", name)
	}
	return id
}

func TestRenderSystemSubmittedDrawCmds(t *testing.T) {
	bDummy := &backenddummy.BackendDummy{}
	ctx := &Context{W: NewWorld(), RM: NewRessourceManager(), p: bDummy, gpu: bDummy, i: bDummy}

	for _, name := range []string{"solid_color1", "solid_color2", "solid_color3"} {
		mustLoadEffect(t, ctx, `{
			"name": "`+name+`",
			"vertex_shader": "PositionColor.vert",
			"fragment_shader": "SolidColor.frag",
			"requires_vertex_format": "position_color"
		}`)
	}

	materialId1 := mustLoadMaterial(t, ctx, "mat-001", `{"name": "mat-001","effect": "solid_color1"}`)
	materialId2 := mustLoadMaterial(t, ctx, "mat-002", `{"name": "mat-002","effect": "solid_color2"}`)
	materialId3 := mustLoadMaterial(t, ctx, "mat-003", `{"name": "mat-003","effect": "solid_color3"}`)

	// Pipelines are derived and cached, so a pipeline id is decided by the
	// order pipelineFor first sees each material -- not by anything the test
	// can pass in. Warming the cache here pins the ids without hardcoding
	// them, and warming it in an order that disagrees with the MeshStore
	// insertion order below is what leaves the commands genuinely unsorted,
	// so Sort has real work to do.
	pipelineFor := func(m types.MaterialID) backend.PipelineID {
		return ctx.RM.pipelineFor(ctx, ctx.RM.materialRessources.Get(m),
			types.VertexFormatPositionColor, &backend.PassDef{})
	}
	pipeline3 := pipelineFor(materialId3)
	pipeline1 := pipelineFor(materialId1)
	pipeline2 := pipelineFor(materialId2)

	if pipeline1 == pipeline2 || pipeline2 == pipeline3 || pipeline1 == pipeline3 {
		t.Fatalf("materials with distinct effects share a pipeline: %d %d %d",
			pipeline1, pipeline2, pipeline3)
	}

	mesh1 := NewMeshComponent(ctx, 1, materialId2, types.VertexFormatPositionColor, 0, 3)
	mesh2 := NewMeshComponent(ctx, 1, materialId1, types.VertexFormatPositionColor, 0, 3)
	mesh3 := NewMeshComponent(ctx, 1, materialId3, types.VertexFormatPositionColor, 0, 3)
	mesh4 := NewMeshComponent(ctx, 1, materialId1, types.VertexFormatPositionColor, 0, 3)

	ctx.W.MeshStore.Upsert(0, mesh1)
	ctx.W.MeshStore.Upsert(1, mesh2)
	ctx.W.MeshStore.Upsert(2, mesh3)
	ctx.W.MeshStore.Upsert(3, mesh4)

	// Distinct positions so each command can be checked against its own entity.
	ctx.W.TransformStore.Upsert(0, TransformComponent{Position: Vector3{X: 10, Y: 11}})
	ctx.W.TransformStore.Upsert(1, TransformComponent{Position: Vector3{X: 20, Y: 21}})
	ctx.W.TransformStore.Upsert(2, TransformComponent{Position: Vector3{X: 30, Y: 31}})
	ctx.W.TransformStore.Upsert(3, TransformComponent{Position: Vector3{X: 40, Y: 41}})

	UpdateRenderSystem(ctx, 1)

	if len(bDummy.Submissions) != 1 {
		t.Fatalf("got %d submissions, want 1", len(bDummy.Submissions))
	}
	sub := bDummy.Submissions[0]

	// Sorted by (pipeline, vertex buffer). Uniforms are written into the arena
	// in MeshStore insertion order (mesh1..mesh4), so each entity's offset is
	// its insertion index times the uniform size: sorting reorders the commands
	// but must carry each offset along with the command it belongs to. mesh2
	// and mesh4 share materialId1, so the pipeline cache must give them the
	// same pipeline rather than creating a second one.
	want := []backend.DrawCmd{
		{Pipeline: pipeline3, Material: materialId3, VertexBuffer: mesh3.vb, ArenaOffset: 2 * uniformSize, ArenaLen: uniformSize},
		{Pipeline: pipeline1, Material: materialId1, VertexBuffer: mesh2.vb, ArenaOffset: 1 * uniformSize, ArenaLen: uniformSize},
		{Pipeline: pipeline1, Material: materialId1, VertexBuffer: mesh4.vb, ArenaOffset: 3 * uniformSize, ArenaLen: uniformSize},
		{Pipeline: pipeline2, Material: materialId2, VertexBuffer: mesh1.vb, ArenaOffset: 0 * uniformSize, ArenaLen: uniformSize},
	}
	if !slices.Equal(sub.Cmds, want) {
		t.Errorf("sort gave %+v, want %+v", sub.Cmds, want)
	}

	// Each command's uniform must hold the transform of the entity it came from,
	// which sorting must not disturb.
	wantUniform := map[backend.VertexBufferID]mesh2dUniform{
		mesh1.vb: {X: 10, Y: 11},
		mesh2.vb: {X: 20, Y: 21},
		mesh3.vb: {X: 30, Y: 31},
		mesh4.vb: {X: 40, Y: 41},
	}
	for _, c := range sub.Cmds {
		if c.ArenaOffset+c.ArenaLen > uint32(len(sub.Arena)) {
			t.Errorf("vb %d: slice [%d:%d] is outside the %d-byte arena",
				c.VertexBuffer, c.ArenaOffset, c.ArenaOffset+c.ArenaLen, len(sub.Arena))
			continue
		}
		got := *(*mesh2dUniform)(unsafe.Pointer(&sub.Arena[c.ArenaOffset]))
		if got != wantUniform[c.VertexBuffer] {
			t.Errorf("vb %d: uniform = %+v, want %+v", c.VertexBuffer, got, wantUniform[c.VertexBuffer])
		}
	}
}
