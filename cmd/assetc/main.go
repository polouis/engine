// Command assetc builds game assets: shaders, sprites, meshes, entities.
// Run from a consumer module: go run github.com/polouis/engine/cmd/assetc all
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type config struct {
	assetDir, distDir, tmpDir string
}

func main() {
	var c config
	flag.StringVar(&c.assetDir, "asset", "asset", "source asset directory")
	flag.StringVar(&c.distDir, "dist", "dist", "output dist directory")
	flag.StringVar(&c.tmpDir, "tmp", "tmp", "temporary working directory")
	flag.Parse()

	cmd := flag.Arg(0)
	if cmd == "" {
		cmd = "all"
	}
	if err := c.run(cmd); err != nil {
		fmt.Fprintln(os.Stderr, "assetc:", err)
		os.Exit(1)
	}
}

func (c config) run(cmd string) error {
	if err := c.mkdirs(); err != nil {
		return err
	}
	switch cmd {
	case "shaders":
		return c.shaders()
	case "sprites":
		return c.sprites()
	case "meshes":
		return c.copyJSON("mesh")
	case "entities":
		return c.copyJSON("entity")
	case "all":
		for _, step := range []func() error{
			c.shaders, c.sprites,
			func() error { return c.copyJSON("mesh") },
			func() error { return c.copyJSON("entity") },
		} {
			if err := step(); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("unknown command %q (want: shaders, sprites, meshes, entities, all)", cmd)
	}
}

func (c config) mkdirs() error {
	for _, d := range []string{
		filepath.Join(c.distDir, "asset", "shader"),
		filepath.Join(c.distDir, "asset", "mesh"),
		filepath.Join(c.distDir, "asset", "entity"),
		filepath.Join(c.distDir, "asset", "sprite"),
		filepath.Join(c.tmpDir, "sprite"),
	} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}
	return nil
}

// shaders: glslc on *.vert.hlsl and *.frag.hlsl  (was asset_vert/frag_shader)
func (c config) shaders() error {
	vulkan := os.Getenv("VULKAN_SDK")
	if vulkan == "" {
		return fmt.Errorf("VULKAN_SDK is not set")
	}
	glslc := filepath.Join(vulkan, "bin", "glslc")

	for pattern, stage := range map[string]string{
		"*.vert.hlsl": "vertex",
		"*.frag.hlsl": "fragment",
	} {
		files, err := filepath.Glob(filepath.Join(c.assetDir, "shaders", "source", pattern))
		if err != nil {
			return err
		}
		for _, in := range files {
			out := filepath.Join(c.distDir, "asset", "shader",
				strings.TrimSuffix(filepath.Base(in), ".hlsl")+".spv")
			if err := sh(glslc, "-fshader-stage="+stage, "-o", out, in); err != nil {
				return err
			}
		}
	}
	return nil
}

// sprites: aseprite export + crunch atlas  (was asset_aseprite + asset_crunch)
func (c config) sprites() error {
	files, err := filepath.Glob(filepath.Join(c.assetDir, "aseprite", "*.aseprite"))
	if err != nil {
		return err
	}
	for _, in := range files {
		name := strings.TrimSuffix(filepath.Base(in), ".aseprite")
		sheet := filepath.Join(c.tmpDir, "sprite", name+".png")
		data := filepath.Join(c.distDir, "asset", "sprite", name+".json")
		if err := sh("aseprite", "-b", in, "--sheet", sheet,
			"--data", data, "--format", "json-array", "--trim"); err != nil {
			return err
		}
	}
	atlas := filepath.Join(c.distDir, "asset", "sprite", "atlas")
	return sh("crunch", atlas, filepath.Join(c.tmpDir, "sprite"), "-v", "-j")
}

// copyJSON: cp asset/<kind>/*.json -> dist/asset/<kind>  (was asset_mesh/entity)
func (c config) copyJSON(kind string) error {
	files, err := filepath.Glob(filepath.Join(c.assetDir, kind, "*.json"))
	if err != nil {
		return err
	}
	for _, in := range files {
		if err := copyFile(in, filepath.Join(c.distDir, "asset", kind, filepath.Base(in))); err != nil {
			return err
		}
	}
	return nil
}

func sh(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}
