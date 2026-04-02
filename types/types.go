package types

type VertexBuffer interface{}

type Backend interface {
	Run(init func(), update func(), release func()) error
	NewVertexBuffer() VertexBuffer
	Draw(vb VertexBuffer)
	Release(vb VertexBuffer)
}
