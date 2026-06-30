package engine

import (
	"fmt"
	"os"
)

type AtlasEntry struct {
	name       string
	x, y, w, h int
}

type AtlasLoad func(dat []byte) []AtlasEntry

type AnimationFrame struct {
	frame            Rect
	spriteSourceSize Rect
	duration         int
}

type ImageLoad func(dat []byte) []AnimationFrame

type Atlas struct {
	entries map[string]AtlasEntry
}

func NewAtlas() *Atlas {
	return &Atlas{entries: make(map[string]AtlasEntry)}
}

func (a *Atlas) Load(al AtlasLoad, il ImageLoad) error {

	if len(a.entries) != 0 {
		return fmt.Errorf("Atlas already loaded")
	}
	atlasPath := "asset/sprite/atlas.json"
	b, err := os.ReadFile(atlasPath)
	if err != nil {
		return fmt.Errorf("Failed reading atlas %s %w", atlasPath, err)
	}
	atlasEntries := al(b)
	if len(atlasEntries) == 0 {
		return fmt.Errorf("Atlas %s has no entries", atlasPath)
	}

	for _, atlasEntry := range atlasEntries {
		a.entries[atlasEntry.name] = atlasEntry
		imagePath := fmt.Sprintf("asset/sprite/%s.json", atlasEntry.name)
		b, err := os.ReadFile(imagePath)
		if err != nil {
			return fmt.Errorf("Failed opening image %s %w", imagePath, err)
		}
		il(b)
	}
	return nil
}
