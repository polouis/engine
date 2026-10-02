package backendsdl

import (
	"errors"
	"unsafe"

	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/polouis/engine/types"
)

type BasicPipeline struct {
	pipeline *sdl.GPUGraphicsPipeline
}

func (vb *BasicPipeline) Init(window *sdl.Window, device *sdl.GPUDevice) error {
	// create shaders

	vertexShader, err := loadShader(
		device, "PositionColor.vert", 0, 1, 0, 0,
	)
	if err != nil {
		panic("failed to create vertex shader: " + err.Error())
	}

	fragmentShader, err := loadShader(
		device, "SolidColor.frag", 0, 0, 0, 0,
	)
	if err != nil {
		panic("failed to create fragment shader: " + err.Error())
	}

	// create pipelines

	colorTargetDescriptions := []sdl.GPUColorTargetDescription{
		{
			Format: device.SwapchainTextureFormat(window),
		},
	}

	vertexBufferDescriptions := []sdl.GPUVertexBufferDescription{
		{
			Slot:             0,
			InputRate:        sdl.GPU_VERTEXINPUTRATE_VERTEX,
			InstanceStepRate: 0,
			Pitch:            uint32(unsafe.Sizeof(types.PositionColorVertex{})),
		},
	}

	vertexAttributes := []sdl.GPUVertexAttribute{
		{
			BufferSlot: 0,
			Format:     sdl.GPU_VERTEXELEMENTFORMAT_FLOAT3,
			Location:   0,
			Offset:     0,
		},
		{
			BufferSlot: 0,
			Format:     sdl.GPU_VERTEXELEMENTFORMAT_UBYTE4_NORM,
			Location:   1,
			Offset:     uint32(unsafe.Sizeof(float32(0)) * 3),
		},
	}

	pipelineCreateInfo := sdl.GPUGraphicsPipelineCreateInfo{
		TargetInfo: sdl.GPUGraphicsPipelineTargetInfo{
			ColorTargetDescriptions: colorTargetDescriptions,
		},
		VertexInputState: sdl.GPUVertexInputState{
			VertexBufferDescriptions: vertexBufferDescriptions,
			VertexAttributes:         vertexAttributes,
		},
		PrimitiveType:  sdl.GPU_PRIMITIVETYPE_TRIANGLELIST,
		VertexShader:   vertexShader,
		FragmentShader: fragmentShader,
	}

	vb.pipeline, err = device.CreateGraphicsPipeline(&pipelineCreateInfo)
	if err != nil {
		return errors.New("failed to create pipeline: " + err.Error())
	}

	device.ReleaseShader(vertexShader)
	device.ReleaseShader(fragmentShader)

	return nil
}

func (p *BasicPipeline) bind(renderPass *sdl.GPURenderPass) {
	renderPass.BindGraphicsPipeline(p.pipeline)
}

func (p *BasicPipeline) release(device *sdl.GPUDevice) {
	device.ReleaseGraphicsPipeline(p.pipeline)
}
