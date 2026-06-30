package assets

import (
	"fmt"
	"os"
)

// //go:embed shaders/compiled
// var assets embed.FS

//	func ReadFile(path string) ([]byte, error) {
//		return assets.ReadFile(path)
//	}
func ReadFile(path string) ([]byte, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		panic(fmt.Errorf("Failed to open %s", path))
	}
	return b, nil
}
