package engine

import "fmt"

type Mesh2dComponent struct {
	vertices []Vertex
	count    uint
}

var Mesh2dCID = RegisterComponent[Mesh2dComponent]()

func GetMesh2dComponents(w *World) *ComponentArray[Mesh2dComponent] {
	return w.Store(VelocityCID).(*ComponentArray[Mesh2dComponent])
}

type SpriteComponent struct {
	x, y               int
	rotation           float32
	textureU, textureV float32
}

var SpriteCID = RegisterComponent[SpriteComponent]()

func GetSpriteComponents(w *World) *ComponentArray[SpriteComponent] {
	return w.Store(VelocityCID).(*ComponentArray[SpriteComponent])
}

func UpdateRenderSystem(w *World, deltatime uint64) {
	for e, velocityCpnt := range GetVelocityComponents(w).All() {
		fmt.Printf("Rendering entity %d with component %T\n", e, velocityCpnt)
	}
}
