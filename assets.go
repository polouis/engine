package engine

import "embed"

//go:embed shaders/compiled
var assets embed.FS

func ReadFile(path string) ([]byte, error) {
	return assets.ReadFile(path)
}
