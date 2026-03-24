package engine

const MaxEntities = 10_000

type ComponentArray[T any] struct {
	arr []T
}

func (ca *ComponentArray[T]) Get(e EntityID) *T {
	return &ca.arr[e]
}

func (ca *ComponentArray[T]) Set(e EntityID, c T) {
	// Do not grow automatically the slice when exceeding capacity to prevent pointers
	// // returned by component getter to become dangling pointers.
	if int(e) >= len(ca.arr) {
		panic("ComponentArray: entity ID exceeds capacity")
	}
	ca.arr[e] = c
}

// func (ca *ComponentArray[T]) Remove(e EntityID) {
// }

func NewComponentArray[T any]() *ComponentArray[T] {
	return &ComponentArray[T]{arr: make([]T, MaxEntities)}
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

func GetNameComponent(w *World, e EntityID) *NameComponent {
	return w.Store(NameCID).(*ComponentArray[NameComponent]).Get(e)
}

func SetNameComponent(w *World, e EntityID, c NameComponent) {
	w.Store(NameCID).(*ComponentArray[NameComponent]).Set(e, c)
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

func GetTransformComponent(w *World, e EntityID) *TransformComponent {
	return w.Store(TransformCID).(*ComponentArray[TransformComponent]).Get(e)
}

func SetTransformComponent(w *World, e EntityID, c TransformComponent) {
	w.Store(TransformCID).(*ComponentArray[TransformComponent]).Set(e, c)
}
