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

func (vb *BasicVertexBuffer) Create(window *sdl.Window, device *sdl.GPUDevice, vertexSize uint32, maxLen uint32) error {

	vb.len = maxLen

	var err error

	vb.vertexBuffer, err = device.CreateBuffer(&sdl.GPUBufferCreateInfo{
		Usage: sdl.GPU_BUFFERUSAGE_VERTEX,
		Size:  vertexSize * maxLen,
	})
	if err != nil {
		return errors.New("failed to create buffer: " + err.Error())
	}
	return nil
}

func (vb *BasicVertexBuffer) Upload(device *sdl.GPUDevice, data []byte) error {

	transferBuffer, err := device.CreateTransferBuffer(&sdl.GPUTransferBufferCreateInfo{
		Usage: sdl.GPU_TRANSFERBUFFERUSAGE_UPLOAD,
		Size:  uint32(len(data)),
	})
	if err != nil {
		return errors.New("failed to create transfer buffer: " + err.Error())
	}

	transferDataPtr, err := device.MapTransferBuffer(transferBuffer, false)
	if err != nil {
		return errors.New("failed to map transfer buffer: " + err.Error())
	}

	vertexData := unsafe.Slice((*byte)(unsafe.Pointer(transferDataPtr)), len(data))
	copy(vertexData, data)

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
			Size:   uint32(unsafe.Sizeof(types.PositionColorVertex{}) * uintptr(len(data))),
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
