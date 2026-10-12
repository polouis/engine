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

func TestRenderSystemSubmittedDrawCmds(t *testing.T) {

	bDummy := &backenddummy.BackendDummy{}
	ctx := &Context{W: NewWorld(), RM: NewRessourceManager(), p: bDummy, gpu: bDummy, i: bDummy}

	// Arbitrary, deliberately unsorted pipeline ids: the render system must group
	// by pipeline regardless of the order entities are stored in.
	mesh1 := NewMeshComponent(ctx, []types.PositionColorVertex{}, 2)
	mesh2 := NewMeshComponent(ctx, []types.PositionColorVertex{}, 1)
	mesh3 := NewMeshComponent(ctx, []types.PositionColorVertex{}, 3)
	mesh4 := NewMeshComponent(ctx, []types.PositionColorVertex{}, 1)

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

	// Uniforms are written into the arena in MeshStore insertion order
	// (mesh1..mesh4), so each entity's offset is its insertion index times the
	// uniform size. Sorting reorders the commands but must carry each offset
	// along with the command it belongs to.
	want := []backend.DrawCmd{
		{Pipeline: 1, VertexBuffer: mesh2.vb, ArenaOffset: 1 * uniformSize, ArenaLen: uniformSize},
		{Pipeline: 1, VertexBuffer: mesh4.vb, ArenaOffset: 3 * uniformSize, ArenaLen: uniformSize},
		{Pipeline: 2, VertexBuffer: mesh1.vb, ArenaOffset: 0 * uniformSize, ArenaLen: uniformSize},
		{Pipeline: 3, VertexBuffer: mesh3.vb, ArenaOffset: 2 * uniformSize, ArenaLen: uniformSize},
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
