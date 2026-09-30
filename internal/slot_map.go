package internal

import "fmt"

const (
	idxBits       = 20
	idxMask       = (1 << idxBits) - 1
	genBits       = 32 - idxBits
	genMask       = (1 << genBits) - 1
	InvalidHandle = 0
)

type slot[T any] struct {
	value *T
	gen   uint32
}

type SlotMap[T any, I ~uint32] struct {
	slots    []slot[T]
	freeList []uint32
}

func makeID[I ~uint32](idx, gen uint32) I {
	return I(gen<<idxBits | idx)
}

func splitID[I ~uint32](id I) (idx, gen uint32) {
	return uint32(id) & idxMask, uint32(id) >> idxBits
}

func (s *SlotMap[T, I]) resolve(id I) (*slot[T], uint32) {
	if id == InvalidHandle {
		panic(fmt.Errorf("slot map: use of the reserved zero handle (unset or zero-valued)"))
	}
	idx, gen := splitID(id)
	if int(idx) >= len(s.slots) {
		panic(fmt.Errorf("slot map: handle index %d out of bounds (len=%d)", idx, len(s.slots)))
	}
	sl := &s.slots[idx]
	if sl.gen != gen {
		panic(fmt.Errorf("slot map: stale handle (index=%d, gen=%d, live gen=%d)", idx, gen, sl.gen))
	}
	return sl, idx
}

func (s *SlotMap[T, I]) Alloc(value *T) I {
	var idx uint32
	if n := len(s.freeList); n > 0 {
		idx = s.freeList[n-1]
		s.freeList = s.freeList[:n-1]
		s.slots[idx].value = value // gen was already bumped by Release
	} else {
		idx = uint32(len(s.slots))
		if idx > idxMask {
			panic(fmt.Errorf("slot map full: %d slots allocated, max %d", idx, idxMask))
		}
		s.slots = append(s.slots, slot[T]{value: value, gen: 1})
	}
	return makeID[I](idx, s.slots[idx].gen)
}

func (s *SlotMap[T, I]) Lookup(id I) *T {
	sl, _ := s.resolve(id)
	return sl.value
}

func (s *SlotMap[T, I]) Free(id I) *T {
	sl, idx := s.resolve(id)
	value := sl.value
	sl.value = nil
	sl.gen = (sl.gen + 1) & genMask
	if sl.gen == 0 {
		sl.gen = 1
	}
	s.freeList = append(s.freeList, idx)
	return value
}
