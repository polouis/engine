package backendsdl

import (
	"errors"
	"fmt"
	"strings"
	"unsafe"

	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/polouis/engine/assets"
	"github.com/polouis/engine/types"
)

type BasicVertexBuffer struct {
	pipeline     *sdl.GPUGraphicsPipeline
	vertexBuffer *sdl.GPUBuffer
	len          uint32
}

func loadShader(
	device *sdl.GPUDevice,
	shaderFilename string,
	samplerCount uint32,
	uniformBufferCount uint32,
	storageBufferCount uint32,
	storageTextureCount uint32,
) (*sdl.GPUShader, error) {
	var stage sdl.GPUShaderStage
	if strings.Contains(shaderFilename, ".vert") {
		stage = sdl.GPU_SHADERSTAGE_VERTEX
	} else if strings.Contains(shaderFilename, ".frag") {
		stage = sdl.GPU_SHADERSTAGE_FRAGMENT
	} else {
		return nil, errors.New("invalid shader stage")
	}

	path := ""
	backendFormats := device.ShaderFormats()
	format := sdl.GPU_SHADERFORMAT_INVALID
	entrypoint := ""

	// fmt.Printf("BACKEND FORMATS: %08b\n", backendFormats)

	if backendFormats&sdl.GPU_SHADERFORMAT_SPIRV == sdl.GPU_SHADERFORMAT_SPIRV {
		path = fmt.Sprintf("shaders/compiled/%s.spv", shaderFilename)
		format = sdl.GPU_SHADERFORMAT_SPIRV
		entrypoint = "main"
	} else if backendFormats&sdl.GPU_SHADERFORMAT_MSL == sdl.GPU_SHADERFORMAT_MSL {
		path = fmt.Sprintf("shaders/compiled/%s.msl", shaderFilename)
		format = sdl.GPU_SHADERFORMAT_MSL
		entrypoint = "main0"
	} else if backendFormats&sdl.GPU_SHADERFORMAT_DXIL == sdl.GPU_SHADERFORMAT_DXIL {
		path = fmt.Sprintf("shaders/compiled/%s.dxil", shaderFilename)
		format = sdl.GPU_SHADERFORMAT_DXIL
		entrypoint = "main"
	} else {
		return nil, errors.New("unrecognized backend shader format")
	}

	code, err := assets.ReadFile(path)
	if err != nil {
		return nil, errors.New("failed to open shader: " + err.Error())
	}

	shaderInfo := sdl.GPUShaderCreateInfo{
		Code:               code,
		Entrypoint:         entrypoint,
		Format:             format,
		Stage:              stage,
		NumSamplers:        samplerCount,
		NumUniformBuffers:  uniformBufferCount,
		NumStorageBuffers:  storageBufferCount,
		NumStorageTextures: storageTextureCount,
	}

	shader, err := device.CreateGPUShader(&shaderInfo)
	if err != nil {
		return nil, errors.New("failed to create shader: " + err.Error())
	}

	return shader, nil
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

func (vb *BasicVertexBuffer) release(window *sdl.Window, device *sdl.GPUDevice) {
	device.ReleaseGraphicsPipeline(vb.pipeline)
	device.ReleaseBuffer(vb.vertexBuffer)
}
