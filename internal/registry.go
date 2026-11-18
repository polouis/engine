package internal

import (
	"fmt"
	"iter"
)

type Registry[T any, I ~uint32] struct {
	slots SlotMap[T, I]
	ids   map[string]I
	names map[I]string // kept only so Remove can evict from ids
}

func (r *Registry[T, I]) Add(name string, value *T) (I, error) {
	if name == "" {
		return InvalidHandle, fmt.Errorf("registry: cannot register a value under an empty name")
	}
	if _, exists := r.ids[name]; exists {
		return InvalidHandle, fmt.Errorf("registry: name %q is already registered", name)
	}
	if r.ids == nil {
		r.ids = make(map[string]I)
		r.names = make(map[I]string)
	}

	id := r.slots.Alloc(value)
	r.ids[name] = id
	r.names[id] = name

	return id, nil
}

func (r *Registry[T, I]) ID(name string) (I, bool) {
	id, exists := r.ids[name]
	if !exists {
		return InvalidHandle, false
	}
	return id, true
}

func (r *Registry[T, I]) Name(id I) string {
	return r.names[id]
}

func (r *Registry[T, I]) Get(id I) *T {
	return r.slots.Lookup(id)
}

func (r *Registry[T, I]) Remove(id I) *T {
	value := r.slots.Free(id)
	delete(r.ids, r.names[id])
	delete(r.names, id)
	return value
}

// All yields the handle and value of every registered entry. Add and Remove
// must not be called during iteration.
func (r *Registry[T, I]) All() iter.Seq2[I, *T] {
	return r.slots.All()
}
