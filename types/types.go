package types

type BackendType string

const (
	SDL   BackendType = "sdl"
	Dummy BackendType = "dummy"
)

/**
 * GRAPHICS
 */

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

/**
 * INPUT Keyboard
 */
type KeyType uint8

const (
	Up KeyType = iota
	Down
	Left
	Right
	Z
	Q
	W
	A
	S
	D

	FirstKey = Up
	LastKey  = D
)

/**
 * INPUT Gamepad
 */
type ButtonType uint8

const (
	ButtonUp ButtonType = iota
	ButtonDown
	ButtonLeft
	ButtonRight
	ButtonX
	ButtonY
	ButtonA
	ButtonB

	ButtonFirst = ButtonUp
	ButtonLast  = ButtonB
)
