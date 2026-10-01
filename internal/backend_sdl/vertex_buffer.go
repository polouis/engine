package backendsdl

import (
	"errors"
	"unsafe"

	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/polouis/engine/types"
)

type BasicVertexBuffer struct {
	pipeline     *sdl.GPUGraphicsPipeline
	vertexBuffer *sdl.GPUBuffer
	len          uint32
}

func (vb *BasicVertexBuffer) Init(window *sdl.Window, device *sdl.GPUDevice, vbData []types.PositionColorVertex) error {

	vb.len = uint32(len(vbData))

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

	// create vertex buffer

	vb.vertexBuffer, err = device.CreateBuffer(&sdl.GPUBufferCreateInfo{
		Usage: sdl.GPU_BUFFERUSAGE_VERTEX,
		Size:  uint32(unsafe.Sizeof(types.PositionColorVertex{}) * uintptr(len(vbData))),
	})
	if err != nil {
		return errors.New("failed to create buffer: " + err.Error())
	}

	// to get data into the vertex buffer, we have to use a transfer buffer

	transferBuffer, err := device.CreateTransferBuffer(&sdl.GPUTransferBufferCreateInfo{
		Usage: sdl.GPU_TRANSFERBUFFERUSAGE_UPLOAD,
		Size:  uint32(unsafe.Sizeof(types.PositionColorVertex{}) * uintptr(len(vbData))),
	})
	if err != nil {
		return errors.New("failed to create transfer buffer: " + err.Error())
	}

	transferDataPtr, err := device.MapTransferBuffer(transferBuffer, false)
	if err != nil {
		return errors.New("failed to map transfer buffer: " + err.Error())
	}

	vertexData := unsafe.Slice(
		(*types.PositionColorVertex)(unsafe.Pointer(transferDataPtr)), len(vbData),
	)

	for i := 0; i < len(vbData); i++ {
		vertexData[i] = vbData[i]
	}

	device.UnmapTransferBuffer(transferBuffer)

	// upload the transfer data to the vertex buffer

	uploadCmdBuf, err := device.AcquireCommandBuffer()
	if err != nil {
		return errors.New("failed to acquire command buffer: " + err.Error())
	}

	copyPass := uploadCmdBuf.BeginCopyPass()

	copyPass.UploadToGPUBuffer(
		&sdl.GPUTransferBufferLocation{
			TransferBuffer: transferBuffer,
			Offset:         0,
		},
		&sdl.GPUBufferRegion{
			Buffer: vb.vertexBuffer,
			Offset: 0,
			Size:   uint32(unsafe.Sizeof(types.PositionColorVertex{}) * uintptr(len(vbData))),
		},
		false,
	)

	copyPass.End()
	uploadCmdBuf.Submit()
	device.ReleaseTransferBuffer(transferBuffer)

	return nil
}

func (vb *BasicVertexBuffer) draw(renderPass *sdl.GPURenderPass) error {

	renderPass.BindGraphicsPipeline(vb.pipeline)
	renderPass.BindVertexBuffers([]sdl.GPUBufferBinding{
		{Buffer: vb.vertexBuffer, Offset: 0},
	})
	renderPass.DrawPrimitives(vb.len, 1, 0, 0)

	return nil
}

func (vb *BasicVertexBuffer) release(device *sdl.GPUDevice) {
	device.ReleaseGraphicsPipeline(vb.pipeline)
	device.ReleaseBuffer(vb.vertexBuffer)
}
