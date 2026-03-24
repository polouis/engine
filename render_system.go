package engine

type Mesh2dComponent struct {
	vertices Vertex
	count    uint
}

type SpriteComponent struct {
	x, y               int
	rotation           float32
	textureU, textureV float32
}

func RenderSystemUpdate(w *World) {

}
