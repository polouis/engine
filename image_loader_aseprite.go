package engine

import (
	"encoding/json"
)

type AsepriteFrame struct {
	Filename         string `json:"filename"`
	Frame            Rect   `json:"frame"`
	Rotated          bool   `json:"rotated"`
	Trimmed          bool   `json:"trimmed"`
	SpriteSourceSize Rect   `json:"spriteSourceSize"`
	SourceSize       Size   `json:"sourceSize"`
	Duration         uint   `json:"duration"`
}

type AsepriteMeta struct {
	App     string `json:"app"`
	Version string `json:"version"`
	Image   string `json:"image"`
	Format  string `json:"format"`
	Size    Size   `json:"size"`
	Scale   string `json:"scale"`
}

type AsepriteFile struct {
	Frames []AsepriteFrame `json:"frames"`
	Meta   AsepriteMeta    `json:"meta"`
}

func ImageLoadAseprite(dat []byte) []AnimationFrame {
	var aDat AsepriteFile
	if err := json.Unmarshal(dat, &aDat); err != nil {
		panic(err)
	}

	afs := make([]AnimationFrame, len(aDat.Frames))
	for i, asepriteAnimFrame := range aDat.Frames {
		afs[i] = AnimationFrame{
			frame:            asepriteAnimFrame.Frame,
			spriteSourceSize: asepriteAnimFrame.SpriteSourceSize,
			duration:         int(asepriteAnimFrame.Duration),
		}
	}

	return afs
}
