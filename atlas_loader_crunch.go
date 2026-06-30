package engine

import (
	"encoding/json"
	"fmt"
)

type crunchImage struct {
	N string `json:"n"`
	X int    `json:"X"`
	Y int    `json:"Y"`
	W int    `json:"W"`
	H int    `json:"H"`
}

type crunchTexture struct {
	Name   string        `json:"name"`
	Images []crunchImage `json:"images"`
}

type crunchFile struct {
	Textures []crunchTexture `json:"textures"`
}

func AtlasLoadCrunch(dat []byte) []AtlasEntry {
	var aDat crunchFile
	if err := json.Unmarshal(dat, &aDat); err != nil {
		panic(err)
	}

	if len(aDat.Textures) != 1 {
		panic(fmt.Errorf("Atas contains :d textures, should be one exactly", len(aDat.Textures)))
	}

	aes := make([]AtlasEntry, len(aDat.Textures[0].Images))

	for i, image := range aDat.Textures[0].Images {
		aes[i] = AtlasEntry{
			name: image.N,
			x:    image.X,
			y:    image.Y,
			w:    image.W,
			h:    image.H,
		}
	}

	return aes
}
