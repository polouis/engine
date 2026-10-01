package backendsdl

import (
	"github.com/Zyko0/go-sdl3/sdl"
)

type BasicPipeline struct {
}

func (vb *BasicPipeline) Init(window *sdl.Window, device *sdl.GPUDevice) error {
	return nil
}
