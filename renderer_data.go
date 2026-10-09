package engine

import (
	"unsafe"

	"github.com/polouis/engine/internal/backend"
)

type rendererData struct {
	cmds  []backend.DrawCmd
	arena []byte
}

func (r *rendererData) Reset() {
	r.cmds = r.cmds[:0]   // length 0, capacity kept
	r.arena = r.arena[:0] // same
}

func (r *rendererData) pushBytes(b []byte) (off, n uint32) {
	off = uint32(len(r.arena))
	r.arena = append(r.arena, b...)
	return off, uint32(len(b))
}

func pushUniform[T any](r *rendererData, v T) (off, n uint32) {
	return r.pushBytes(unsafe.Slice((*byte)(unsafe.Pointer(&v)), unsafe.Sizeof(v)))
}
