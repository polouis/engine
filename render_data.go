package engine

import (
	"cmp"
	"slices"
	"unsafe"

	"github.com/polouis/engine/internal/backend"
)

type renderData struct {
	cmds  []backend.DrawCmd
	arena []byte
}

func (r *renderData) Reset() {
	r.cmds = r.cmds[:0]   // length 0, capacity kept
	r.arena = r.arena[:0] // same
}

func (r *renderData) Sort() {
	slices.SortFunc(r.cmds, func(a, b backend.DrawCmd) int {
		if a.Pipeline != b.Pipeline {
			return cmp.Compare(a.Pipeline, b.Pipeline)
		}
		return cmp.Compare(a.VertexBuffer, b.VertexBuffer)
	})
}

func (r *renderData) pushBytes(b []byte) (off, n uint32) {
	off = uint32(len(r.arena))
	r.arena = append(r.arena, b...)
	return off, uint32(len(b))
}

func pushUniform[T any](r *renderData, v T) (off, n uint32) {
	return r.pushBytes(unsafe.Slice((*byte)(unsafe.Pointer(&v)), unsafe.Sizeof(v)))
}
