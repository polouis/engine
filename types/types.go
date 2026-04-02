package types

type PositionColorVertex struct {
	X, Y, Z    float32
	R, G, B, A uint8
}

func NewPosColorVert(x, y, z float32, r, g, b, a uint8) PositionColorVertex {
	return PositionColorVertex{
		X: x, Y: y, Z: z,
		R: r, G: g, B: b, A: a,
	}
}

type VertexBuffer interface{}
type Backend interface {
	Run(init func(), update func(), release func()) error
	NewVertexBuffer([]PositionColorVertex) VertexBuffer
	Draw(vb VertexBuffer)
	Release(vb VertexBuffer)
}
