package backendsdl

import (
	"errors"
	"fmt"
	"os"

	"github.com/Zyko0/go-sdl3/bin/binsdl"
	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/polouis/engine/internal/backend"
	"github.com/polouis/engine/types"
)

type BackendSDL struct {
	window       *sdl.Window
	device       *sdl.GPUDevice
	rp           *sdl.GPURenderPass
	keyStates    [types.LastKey + 1]bool
	buttonStates [types.ButtonLast + 1]bool
}

type drawable interface {
	draw(rp *sdl.GPURenderPass) error
}

type releasable interface {
	release(window *sdl.Window, device *sdl.GPUDevice)
}

func (b *BackendSDL) Run(initCallback func(), updateCallback func(uint64), releaseCallback func()) error {
	defer binsdl.Load().Unload() // sdl.LoadLibrary(sdl.Path())
	defer sdl.Quit()

	if err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_GAMEPAD); err != nil {
		return errors.New("failed to initialize SDL: " + err.Error())
	}

	var err error
	b.window, err = sdl.CreateWindow("Hello world", 500, 500, sdl.WINDOW_RESIZABLE)
	if err != nil {
		return errors.New("failed to create window: " + err.Error())
	}
	defer b.window.Destroy()

	b.device, err = sdl.CreateGPUDevice(sdl.GPU_SHADERFORMAT_SPIRV, true, "")
	if err != nil {
		return errors.New("failed to create gpu device: " + err.Error())
	}
	defer b.device.Destroy()

	fmt.Println("Driver: " + b.device.Driver())

	b.device.ClaimWindow(b.window)

	var gp *sdl.Gamepad

	initCallback()

	sdl.RunLoop(func() error {
		var event sdl.Event

		for sdl.PollEvent(&event) {
			switch event.Type {
			case sdl.EVENT_QUIT:
				return sdl.EndLoop
			case sdl.EVENT_GAMEPAD_ADDED:
				evt := event.GamepadDeviceEvent()
				if gp == nil {
					gp, err = evt.Which.OpenGamepad()
					if err != nil {
						fmt.Fprintf(os.Stderr, "failed to open gamepad ID %d: %s\n", evt.Which, err.Error())
					}
				}
			}
		}

		for i := types.ButtonFirst; i <= types.ButtonLast; i++ {
			b.buttonStates[i] = false
		}
		if gp != nil {
			if gp.Button(sdl.GAMEPAD_BUTTON_DPAD_UP) {
				b.buttonStates[types.ButtonUp] = true
			}
			if gp.Button(sdl.GAMEPAD_BUTTON_DPAD_DOWN) {
				b.buttonStates[types.ButtonDown] = true
			}
			if gp.Button(sdl.GAMEPAD_BUTTON_DPAD_LEFT) {
				b.buttonStates[types.ButtonLeft] = true
			}
			if gp.Button(sdl.GAMEPAD_BUTTON_DPAD_RIGHT) {
				b.buttonStates[types.ButtonRight] = true
			}
		}

		for i := types.FirstKey; i <= types.LastKey; i++ {
			b.keyStates[i] = false
		}
		keyStates := sdl.GetKeyboardState()
		if keyStates[sdl.SCANCODE_UP] || keyStates[sdl.SCANCODE_W] {
			b.keyStates[types.Up] = true
		}
		if keyStates[sdl.SCANCODE_DOWN] || keyStates[sdl.SCANCODE_S] {
			b.keyStates[types.Down] = true
		}
		if keyStates[sdl.SCANCODE_LEFT] || keyStates[sdl.SCANCODE_A] {
			b.keyStates[types.Left] = true
		}
		if keyStates[sdl.SCANCODE_RIGHT] || keyStates[sdl.SCANCODE_D] {
			b.keyStates[types.Right] = true
		}

		b.update(updateCallback)

		return nil
	})

	releaseCallback()

	return nil
}

func (b *BackendSDL) update(updateCallback func(uint64)) error {
	ticksNS := sdl.TicksNS()
	cmdbuf, err := b.device.AcquireCommandBuffer()
	if err != nil {
		return errors.New("failed to acquire command buffer: " + err.Error())
	}

	swapchainTexture, err := cmdbuf.WaitAndAcquireGPUSwapchainTexture(b.window)
	if err != nil {
		return errors.New("failed to acquire swapchain texture: " + err.Error())
	}

	if swapchainTexture != nil {
		colorTargetInfo := sdl.GPUColorTargetInfo{
			Texture:    swapchainTexture.Texture,
			ClearColor: sdl.FColor{R: 0.3, G: 0.4, B: 0.5, A: 1.0},
			LoadOp:     sdl.GPU_LOADOP_CLEAR,
			StoreOp:    sdl.GPU_STOREOP_STORE,
		}

		b.rp = cmdbuf.BeginRenderPass(
			[]sdl.GPUColorTargetInfo{colorTargetInfo}, nil,
		)

		updateCallback(ticksNS)

		b.rp.End()
	}

	cmdbuf.Submit()

	return nil
}

func (b *BackendSDL) NewVertexBuffer(vbData []types.PositionColorVertex) backend.VertexBuffer {
	var vb BasicVertexBuffer
	vb.Init(b.window, b.device, vbData)
	return &vb
}

func (b *BackendSDL) Draw(vb backend.VertexBuffer) {
	if vbSdl, ok := vb.(drawable); ok {
		vbSdl.draw(b.rp)
	}
}

func (b *BackendSDL) Release(vb backend.VertexBuffer) {
	if d, ok := vb.(releasable); ok {
		d.release(b.window, b.device)
	}
}

func (b *BackendSDL) GetKeyState(k types.KeyType) bool {
	return b.keyStates[k]
}

func (b *BackendSDL) GetButtonState(btn types.ButtonType) bool {
	return b.buttonStates[btn]
}
