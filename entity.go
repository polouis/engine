package engine

import (
	"fmt"
	"iter"
)

const MaxEntities = 10_000

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

func (ca *ComponentArray[T]) Add(e EntityID, c T) {
	i, exists := ca.entity2ComponentMap[e]
	if !exists && ca.count >= MaxEntities {
		// Do not grow automatically the slice when exceeding capacity to prevent
		// pointers returned by component getter to become dangling pointers.
		// TODO add component type in message
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
func (ca *ComponentArray[T]) Remove(e EntityID) {
	deletedIdx, exists := ca.entity2ComponentMap[e]
	if !exists {
		return
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

func NewComponentArray[T any]() *ComponentArray[T] {
	return &ComponentArray[T]{
		count:               0,
		arr:                 make([]T, MaxEntities),
		entity2ComponentMap: make(map[EntityID]uint),
		component2EntityMap: make(map[uint]EntityID),
	}
}

// void insertData(size_t entityId, T component)
// {
//   if (entityId >= components.size())
//   {
//     components.resize(entityId + 1);
//   }
//   components[entityId] = component;
// }
// bool hasData(size_t entityId) const
// {
//   return entityId < components.size() && components[entityId].has_value();
// }
// T& getData(size_t entityId)
// {
//   return components[entityId].value();
// }
// const T& getData(size_t entityId) const
// {
//   return components[entityId].value();
// }
// const std::vector<std::optional<T>>& getComponents() const
// {
//   return components;
// }
// std::vector<std::optional<T>>& getComponents()
// {
//   return components;
// }

type EntityID int
type ComponentID int
type storeFactory func() any

var (
	nextComponentID ComponentID
	factories       []storeFactory
)

func RegisterComponent[T any]() ComponentID {
	id := nextComponentID
	nextComponentID++
	factories = append(factories, func() any {
		return NewComponentArray[T]()
	})
	return id
}

// func NewComponentID() ComponentID {
// 	id := nextComponentID
// 	nextComponentID++
// 	return id
// }

type World struct {
	stores []any // each value is a *ComponentArray[T]
	nextID EntityID
}

func NewWorld() *World {
	w := &World{
		stores: make([]any, len(factories)),
	}
	for i, factory := range factories {
		w.stores[i] = factory()
	}
	return w
}

func (w *World) Store(id ComponentID) any {
	return w.stores[id]
}

func (w *World) NewEntity() EntityID {
	if int(w.nextID) >= MaxEntities {
		panic("World: max entity count reached")
	}
	id := w.nextID
	w.nextID++
	return id
}

// type World struct {
// 	Mesh2ds    ComponentArray[Mesh2dComponent]
// 	Sprites    ComponentArray[SpriteComponent]
// 	Velocities ComponentArray[VelocityComponent]
// }

/**
 * BASE COMPONENTS
 */

var NameCID = RegisterComponent[NameComponent]()

type NameComponent struct {
	name string
}

func GetNameComponent(w *World, e EntityID) (*NameComponent, error) {
	return w.Store(NameCID).(*ComponentArray[NameComponent]).Get(e)
}

func AddNameComponent(w *World, e EntityID, c NameComponent) {
	w.Store(NameCID).(*ComponentArray[NameComponent]).Add(e, c)
}

//   class Transform: public Base
//   {
//   public:
//       Transform(ECS::Entity entity)
//         : Base(entity),
//           position{0.0f, 0.0f, 0.0f},
//           rotation{0.0f, 0.0f, 0.0f},
//           scale{1.0f, 1.0f, 1.0f}
//       {
//       }
//       Transform(ECS::Entity entity, float x, float y, float z)
//         : Base(entity),
//           position{x, y, z},
//           rotation{0.0f, 0.0f, 0.0f},
//           scale{1.0f, 1.0f, 1.0f}
//       {
//       }
//       ~Transform() {}
//       // Position
//       void setPosition(const geometry::Vector3& vec) { position = vec; }
//       const geometry::Vector3& getPosition() const { return position; }
//       float getX() const { return position.getX(); }
//       float getY() const { return position.getY(); }
//       float getZ() const { return position.getZ(); }
//       // Rotation (in radians)
//       void setRotation(float x, float y, float z) { rotation.set(x, y, z); }
//       const geometry::Vector3& getRotation() const { return rotation; }
//       geometry::Vector3& getRotation() { return rotation; }
//       // Scale
//       void setScale(float x, float y, float z) { scale.set(x, y, z); }
//       void setUniformScale(float s) { scale.set(s, s, s); }
//       const geometry::Vector3& getScale() const { return scale; }

var TransformCID = RegisterComponent[TransformComponent]()

type TransformComponent struct {
	position Vector3
	rotation Vector3
	scale    Vector3
}

func GetTransformComponent(w *World, e EntityID) (*TransformComponent, error) {
	return w.Store(TransformCID).(*ComponentArray[TransformComponent]).Get(e)
}

func AddTransformComponent(w *World, e EntityID, c TransformComponent) {
	w.Store(TransformCID).(*ComponentArray[TransformComponent]).Add(e, c)
}
