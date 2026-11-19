package backendsdl

import (
	"errors"
	"fmt"

	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/polouis/engine/assets"
	"github.com/polouis/engine/internal/shadermeta"
)

type Shader struct {
	s *sdl.GPUShader
}

func (s *Shader) init(device *sdl.GPUDevice, shaderFilename string) error {
	meta, err := shadermeta.Load(fmt.Sprintf("asset/shader/%s.json", shaderFilename))
	if err != nil {
		return fmt.Errorf("shader %s: %w", shaderFilename, err)
	}

	var stage sdl.GPUShaderStage
	switch meta.Stage {
	case shadermeta.StageVertex:
		stage = sdl.GPU_SHADERSTAGE_VERTEX
	case shadermeta.StageFragment:
		stage = sdl.GPU_SHADERSTAGE_FRAGMENT
	default:
		return fmt.Errorf("shader %s: unhandled stage %q", shaderFilename, meta.Stage)
	}

	var (
		path       string
		format     sdl.GPUShaderFormat
		entrypoint string
	)
	backendFormats := device.ShaderFormats()
	switch {
	case backendFormats&sdl.GPU_SHADERFORMAT_SPIRV != 0:
		path = fmt.Sprintf("asset/shader/%s.spv", shaderFilename)
		format = sdl.GPU_SHADERFORMAT_SPIRV
		entrypoint = meta.Entrypoint
	case backendFormats&sdl.GPU_SHADERFORMAT_MSL != 0:
		path = fmt.Sprintf("asset/shader/%s.msl", shaderFilename)
		format = sdl.GPU_SHADERFORMAT_MSL
		// SPIRV-Cross renames the entry point when transpiling to MSL, so the
		// reflected name does not apply to this format.
		entrypoint = "main0"
	case backendFormats&sdl.GPU_SHADERFORMAT_DXIL != 0:
		path = fmt.Sprintf("asset/shader/%s.dxil", shaderFilename)
		format = sdl.GPU_SHADERFORMAT_DXIL
		entrypoint = meta.Entrypoint
	default:
		return errors.New("unrecognized backend shader format")
	}

	code, err := assets.ReadFile(path)
	if err != nil {
		return errors.New("failed to open shader: " + err.Error())
	}

	shaderInfo := sdl.GPUShaderCreateInfo{
		Code:               code,
		Entrypoint:         entrypoint,
		Format:             format,
		Stage:              stage,
		NumSamplers:        meta.NumSamplers,
		NumUniformBuffers:  meta.NumUniformBuffers,
		NumStorageBuffers:  meta.NumStorageBuffers,
		NumStorageTextures: meta.NumStorageTextures,
	}

	shader, err := device.CreateGPUShader(&shaderInfo)
	if err != nil {
		return errors.New("failed to create shader: " + err.Error())
	}

	s.s = shader

	return nil
}

func (s Shader) release(device *sdl.GPUDevice) error {
	device.ReleaseShader(s.s)
	return nil
}
