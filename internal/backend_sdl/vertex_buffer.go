package backendsdl

import (
	"errors"
	"unsafe"

	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/polouis/engine/types"
)

type BasicVertexBuffer struct {
	vertexBuffer *sdl.GPUBuffer
	len          uint32
}

func (vb *BasicVertexBuffer) Init(window *sdl.Window, device *sdl.GPUDevice, vbData []types.PositionColorVertex) error {

	vb.len = uint32(len(vbData))

	var err error

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

	renderPass.BindVertexBuffers([]sdl.GPUBufferBinding{
		{Buffer: vb.vertexBuffer, Offset: 0},
	})
	renderPass.DrawPrimitives(vb.len, 1, 0, 0)

	return nil
}

func (vb *BasicVertexBuffer) release(device *sdl.GPUDevice) {
	device.ReleaseBuffer(vb.vertexBuffer)
}
