package engine

import (
	"slices"
	"testing"

	"github.com/polouis/engine/internal/backend"
)

func TestRenderDataSort(t *testing.T) {

	r := renderData{cmds: []backend.DrawCmd{
		{Pipeline: 2, VertexBuffer: 1},
		{Pipeline: 1, VertexBuffer: 2},
		{Pipeline: 2, VertexBuffer: 0},
		{Pipeline: 1, VertexBuffer: 1},
	}}
	r.Sort()

	want := []backend.DrawCmd{
		{Pipeline: 1, VertexBuffer: 1},
		{Pipeline: 1, VertexBuffer: 2},
		{Pipeline: 2, VertexBuffer: 0},
		{Pipeline: 2, VertexBuffer: 1},
	}

	if !slices.Equal(r.cmds, want) {
		t.Errorf("sort gave %+v, want%+v", r.cmds, want)
	}
}
