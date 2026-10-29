package types

import "fmt"

type VertexFormat uint8

const (
	VertexFormatInvalid VertexFormat = iota
	VertexFormatPositionColor
)

type VertexElementFormat uint8

const (
	VertexElementFormatInvalid VertexElementFormat = iota
	VertexElementFormatInt
	VertexElementFormatInt2
	VertexElementFormatInt3
	VertexElementFormatInt4
	VertexElementFormatUInt
	VertexElementFormatUInt2
	VertexElementFormatUInt3
	VertexElementFormatUInt4
	VertexElementFormatFloat
	VertexElementFormatFloat2
	VertexElementFormatFloat3
	VertexElementFormatFloat4
	VertexElementFormatByte2
	VertexElementFormatByte4
	VertexElementFormatUByte2
	VertexElementFormatUByte4
	VertexElementFormatByte2_Norm
	VertexElementFormatByte4_Norm
	VertexElementFormatUByte2_Norm
	VertexElementFormatUByte4_Norm
	VertexElementFormatShort2
	VertexElementFormatShort4
	VertexElementFormatUShort2
	VertexElementFormatUShort4
	VertexElementFormatShort2_Norm
	VertexElementFormatShort4_Norm
	VertexElementFormatUShort2_Norm
	VertexElementFormatUShort4_Norm
	VertexElementFormatHalf2
	VertexElementFormatHalf4
)

type VertexAttribute struct {
	Location uint32
	Format   VertexElementFormat
	Offset   uint32
}

type VertexLayout struct {
	Stride     uint32
	Attributes []VertexAttribute
}

func (f VertexFormat) Layout() VertexLayout {
	switch f {
	case VertexFormatInvalid:
		panic("Cannot get layout from invalid vertex format")
	case VertexFormatPositionColor:
		return VertexLayout{
			Stride: 28,
			Attributes: []VertexAttribute{
				{Location: 0, Format: VertexElementFormatFloat3, Offset: 0},
				{Location: 1, Format: VertexElementFormatFloat4, Offset: 12},
			},
		}
	default:
		panic(fmt.Errorf("Unhandled vertex format %d", f))
	}
}

func ParseVertexFormat(name string) (VertexFormat, error) {
	switch name {
	case "position_color":
		return VertexFormatPositionColor, nil
	default:
		return VertexFormatInvalid, fmt.Errorf("Unhandled vertex format name %s", name)
	}
}
