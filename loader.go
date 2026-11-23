package engine

import (
	"encoding/json"
	"fmt"
)

type descEffect struct {
	Name           string `json:"name"`
	VertexShader   string `json:"vertex_shader"`
	FragmentShader string `json:"fragment_shader"`
	// RequiresVertexFormat string `json:"position_color"`
}

func LoadEffect(dat []byte) (*descEffect, error) {
	var effect descEffect
	if err := json.Unmarshal(dat, &effect); err != nil {
		return nil, fmt.Errorf("Cannot deserialize effect, got error %w", err)
	}
	return &effect, nil
}
