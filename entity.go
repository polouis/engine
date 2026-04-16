package engine

import (
	"fmt"
	"iter"
)

const MaxEntities = 10_000

type EntityID uint

/******************************************************************************
 * COMPONENT STORE
 *****************************************************************************/

type ComponentArray[T any] struct {
	arr                 []T
	entity2ComponentMap map[EntityID]uint
	component2EntityMap map[uint]EntityID
	count               uint
}

func (ca *ComponentArray[T]) Get(e EntityID) (*T, error) {
	i, ok := ca.entity2ComponentMap[e]
	if !ok {
		return nil, fmt.Errorf("ComponentArray: entity %d has no component in component array %T", e, ca)
	}
	return &ca.arr[i], nil
}

func (ca *ComponentArray[T]) Upsert(e EntityID, c T) {
	i, exists := ca.entity2ComponentMap[e]
	if !exists && ca.count >= MaxEntities {
		// Do not grow automatically the slice when exceeding capacity to prevent
		// pointers returned by component getter to become dangling pointers.
		panic(fmt.Sprintf("ComponentArray: component count exceeds capacity for component array %T", ca))
	}

	if exists {
		ca.arr[i] = c
	} else {
		ca.arr[ca.count] = c
		ca.entity2ComponentMap[e] = ca.count
		ca.component2EntityMap[ca.count] = e
		ca.count++
	}
}

// Keep array dense. Move last active component in place of removed one.
func (ca *ComponentArray[T]) Remove(e EntityID) error {
	deletedIdx, exists := ca.entity2ComponentMap[e]
	if !exists {
		return fmt.Errorf("Failed to delete %d as it is not present in component array %T", e, ca)
	}
	lastIdx := ca.count - 1
	lastE := ca.component2EntityMap[lastIdx]

	delete(ca.entity2ComponentMap, e)
	delete(ca.component2EntityMap, deletedIdx)
	ca.count--

	if deletedIdx != lastIdx {
		ca.arr[deletedIdx] = ca.arr[lastIdx]
		ca.entity2ComponentMap[lastE] = deletedIdx
		delete(ca.component2EntityMap, lastIdx)
		ca.component2EntityMap[deletedIdx] = lastE
	}

	return nil
}

func (ca *ComponentArray[T]) All() iter.Seq2[EntityID, *T] {
	return func(yield func(EntityID, *T) bool) {
		for i := range ca.count {
			if !yield(ca.component2EntityMap[i], &ca.arr[i]) {
				return
			}
		}
	}
}

func (ca *ComponentArray[T]) Has(e EntityID) bool {
	_, exists := ca.entity2ComponentMap[e]
	return exists
}

func NewComponentArray[T any]() *ComponentArray[T] {
	return &ComponentArray[T]{
		count:               0,
		arr:                 make([]T, MaxEntities),
		entity2ComponentMap: make(map[EntityID]uint),
		component2EntityMap: make(map[uint]EntityID),
	}
}

var (
	nextID EntityID
)

/******************************************************************************
 * WORLD
 *****************************************************************************/

type World struct {
	NameStore      *ComponentArray[NameComponent]
	TransformStore *ComponentArray[TransformComponent]
	VelocityStore  *ComponentArray[VelocityComponent]
	MeshStore      *ComponentArray[MeshComponent]
}

func NewWorld() *World {
	return &World{
		NameStore:      NewComponentArray[NameComponent](),
		TransformStore: NewComponentArray[TransformComponent](),
		VelocityStore:  NewComponentArray[VelocityComponent](),
		MeshStore:      NewComponentArray[MeshComponent](),
	}
}

func (w *World) NewEntity() EntityID {
	if int(nextID) >= MaxEntities {
		panic("World: max entity count reached")
	}
	id := nextID
	nextID++
	return id
}

/******************************************************************************
 * BASE COMPONENTS
 *****************************************************************************/

type NameComponent struct {
	Name string
}

type TransformComponent struct {
	Position Vector3
	Rotation Vector3
	Scale    Vector3
}
