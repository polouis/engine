// Package shadermeta reflects compiled SPIR-V shaders and carries the result
// as a JSON sidecar stored beside the bytecode.
//
// SDL_GPU needs the exact number of samplers, storage textures, storage
// buffers and uniform buffers that a shader declares: wrong counts fail
// shader or pipeline creation. Those numbers are already in the compiled
// bytecode, so assetc reflects them at build time and the runtime reads them
// back, instead of having them hand-authored in asset files where they can
// drift from the shader.
package shadermeta

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// Stage is the pipeline stage a shader was compiled for, taken from its
// SPIR-V entry point rather than from its filename.
type Stage string

const (
	StageVertex   Stage = "vertex"
	StageFragment Stage = "fragment"
)

// ScalarType is the component type of a vertex input.
type ScalarType string

const (
	ScalarFloat ScalarType = "float"
	ScalarInt   ScalarType = "int"
	ScalarUint  ScalarType = "uint"
)

// Input is one vertex input the shader declares. It is the shader's half of
// the vertex-format contract: the mesh supplies the actual layout, and this
// says what the shader expects to find in it.
type Input struct {
	Location   uint32     `json:"location"`
	Type       ScalarType `json:"type"`
	Components uint32     `json:"components"`
}

// Metadata is everything about a shader that can be derived from its
// bytecode. Inputs is populated for vertex shaders only: a fragment shader's
// inputs are interpolated varyings, not vertex attributes.
type Metadata struct {
	Stage              Stage   `json:"stage"`
	Entrypoint         string  `json:"entrypoint"`
	NumSamplers        uint32  `json:"num_samplers"`
	NumStorageTextures uint32  `json:"num_storage_textures"`
	NumStorageBuffers  uint32  `json:"num_storage_buffers"`
	NumUniformBuffers  uint32  `json:"num_uniform_buffers"`
	Inputs             []Input `json:"inputs,omitempty"`
}

// SidecarPath returns the metadata path for a compiled shader, replacing the
// bytecode extension. The metadata does not depend on the shader format, so
// the .spv, .msl and .dxil builds of one shader share a single sidecar.
func SidecarPath(compiled string) string {
	return strings.TrimSuffix(compiled, filepath.Ext(compiled)) + ".json"
}

// Load reads a sidecar written by WriteFile.
func Load(path string) (*Metadata, error) {
	dat, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("shadermeta: %w", err)
	}
	var m Metadata
	if err := json.Unmarshal(dat, &m); err != nil {
		return nil, fmt.Errorf("shadermeta: cannot deserialize %s: %w", path, err)
	}
	if m.Stage != StageVertex && m.Stage != StageFragment {
		return nil, fmt.Errorf("shadermeta: %s declares unknown stage %q", path, m.Stage)
	}
	return &m, nil
}

// WriteFile writes m as indented JSON.
func (m *Metadata) WriteFile(path string) error {
	dat, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("shadermeta: cannot serialize metadata: %w", err)
	}
	return os.WriteFile(path, append(dat, '\n'), 0o644)
}

/*
 * SPIR-V reflection
 *
 * Only the handful of instructions needed to find descriptor-bound resources
 * and vertex inputs is decoded; everything else is skipped by word count.
 */

const spirvMagic = 0x07230203

// Instruction opcodes.
const (
	opEntryPoint       = 15
	opTypeInt          = 21
	opTypeFloat        = 22
	opTypeVector       = 23
	opTypeImage        = 25
	opTypeSampler      = 26
	opTypeSampledImage = 27
	opTypeArray        = 28
	opTypeRuntimeArray = 29
	opTypeStruct       = 30
	opTypePointer      = 32
	opConstant         = 43
	opVariable         = 59
	opDecorate         = 71
)

// Decorations.
const (
	decBufferBlock   = 3
	decBuiltIn       = 11
	decLocation      = 30
	decBinding       = 33
	decDescriptorSet = 34
)

// Storage classes.
const (
	storageUniformConstant = 0
	storageInput           = 1
	storageUniform         = 2
	storageStorageBuffer   = 12
)

// Execution models.
const (
	modelVertex   = 0
	modelFragment = 4
)

type typeKind int

const (
	kindOther typeKind = iota
	kindFloat
	kindInt
	kindVector
	kindImage
	kindSampler
	kindSampledImage
	kindArray
	kindRuntimeArray
	kindStruct
	kindPointer
)

type typeInfo struct {
	kind    typeKind
	elem    uint32 // vector component, array element, pointer target, sampled image's image
	count   uint32 // vector components, array length
	sampled uint32 // OpTypeImage's Sampled operand: 2 means a storage image
	storage uint32 // OpTypePointer's storage class
	signed  bool   // OpTypeInt's signedness
}

// category is what SDL_GPU counts a descriptor-bound resource as.
type category int

const (
	catSampledTexture category = iota
	catStorageTexture
	catStorageBuffer
	catUniformBuffer
	catSampler
)

type reflector struct {
	types  map[uint32]typeInfo
	consts map[uint32]uint32

	dset        map[uint32]uint32
	binding     map[uint32]uint32
	location    map[uint32]uint32
	builtin     map[uint32]bool
	bufferBlock map[uint32]bool

	varPtr   map[uint32]uint32 // variable id -> its pointer type id
	varOrder []uint32          // declaration order, so output is deterministic
}

// Reflect derives a shader's metadata from its compiled SPIR-V.
func Reflect(code []byte) (*Metadata, error) {
	words, err := decodeWords(code)
	if err != nil {
		return nil, err
	}

	r := &reflector{
		types:       map[uint32]typeInfo{},
		consts:      map[uint32]uint32{},
		dset:        map[uint32]uint32{},
		binding:     map[uint32]uint32{},
		location:    map[uint32]uint32{},
		builtin:     map[uint32]bool{},
		bufferBlock: map[uint32]bool{},
		varPtr:      map[uint32]uint32{},
	}

	var m Metadata
	seenEntry := false

	for i := 5; i < len(words); {
		n := int(words[i] >> 16)
		op := words[i] & 0xffff
		if n < 1 || i+n > len(words) {
			return nil, fmt.Errorf("shadermeta: malformed instruction at word %d", i)
		}
		o := words[i : i+n]

		switch op {
		case opEntryPoint:
			if len(o) < 4 {
				return nil, fmt.Errorf("shadermeta: truncated OpEntryPoint")
			}
			if seenEntry {
				return nil, fmt.Errorf("shadermeta: shader declares more than one entry point, so its stage is ambiguous")
			}
			seenEntry = true
			switch o[1] {
			case modelVertex:
				m.Stage = StageVertex
			case modelFragment:
				m.Stage = StageFragment
			default:
				return nil, fmt.Errorf("shadermeta: unsupported execution model %d (only vertex and fragment shaders are handled)", o[1])
			}
			m.Entrypoint = decodeString(o[3:])

		case opDecorate:
			if len(o) < 3 {
				return nil, fmt.Errorf("shadermeta: truncated OpDecorate")
			}
			target, dec := o[1], o[2]
			switch dec {
			case decBuiltIn:
				r.builtin[target] = true
			case decBufferBlock:
				r.bufferBlock[target] = true
			case decDescriptorSet, decBinding, decLocation:
				if len(o) < 4 {
					return nil, fmt.Errorf("shadermeta: OpDecorate %d has no operand", dec)
				}
				switch dec {
				case decDescriptorSet:
					r.dset[target] = o[3]
				case decBinding:
					r.binding[target] = o[3]
				case decLocation:
					r.location[target] = o[3]
				}
			}

		case opTypeInt:
			if len(o) < 4 {
				return nil, fmt.Errorf("shadermeta: truncated OpTypeInt")
			}
			r.types[o[1]] = typeInfo{kind: kindInt, signed: o[3] == 1}

		case opTypeFloat:
			r.types[o[1]] = typeInfo{kind: kindFloat}

		case opTypeVector:
			if len(o) < 4 {
				return nil, fmt.Errorf("shadermeta: truncated OpTypeVector")
			}
			r.types[o[1]] = typeInfo{kind: kindVector, elem: o[2], count: o[3]}

		case opTypeImage:
			if len(o) < 8 {
				return nil, fmt.Errorf("shadermeta: truncated OpTypeImage")
			}
			r.types[o[1]] = typeInfo{kind: kindImage, sampled: o[7]}

		case opTypeSampler:
			r.types[o[1]] = typeInfo{kind: kindSampler}

		case opTypeSampledImage:
			if len(o) < 3 {
				return nil, fmt.Errorf("shadermeta: truncated OpTypeSampledImage")
			}
			r.types[o[1]] = typeInfo{kind: kindSampledImage, elem: o[2]}

		case opTypeArray:
			if len(o) < 4 {
				return nil, fmt.Errorf("shadermeta: truncated OpTypeArray")
			}
			r.types[o[1]] = typeInfo{kind: kindArray, elem: o[2], count: o[3]} // count is a constant id for now

		case opTypeRuntimeArray:
			if len(o) < 3 {
				return nil, fmt.Errorf("shadermeta: truncated OpTypeRuntimeArray")
			}
			r.types[o[1]] = typeInfo{kind: kindRuntimeArray, elem: o[2]}

		case opTypeStruct:
			if len(o) < 2 {
				return nil, fmt.Errorf("shadermeta: truncated OpTypeStruct")
			}
			r.types[o[1]] = typeInfo{kind: kindStruct}

		case opTypePointer:
			if len(o) < 4 {
				return nil, fmt.Errorf("shadermeta: truncated OpTypePointer")
			}
			r.types[o[1]] = typeInfo{kind: kindPointer, storage: o[2], elem: o[3]}

		case opConstant:
			if len(o) < 4 {
				return nil, fmt.Errorf("shadermeta: truncated OpConstant")
			}
			r.consts[o[2]] = o[3] // low word is enough for an array length

		case opVariable:
			if len(o) < 4 {
				return nil, fmt.Errorf("shadermeta: truncated OpVariable")
			}
			r.varPtr[o[2]] = o[1]
			r.varOrder = append(r.varOrder, o[2])
		}

		i += n
	}

	if !seenEntry {
		return nil, fmt.Errorf("shadermeta: no OpEntryPoint found; is this a SPIR-V shader module?")
	}

	if err := r.countResources(&m); err != nil {
		return nil, err
	}
	if m.Stage == StageVertex {
		if err := r.collectInputs(&m); err != nil {
			return nil, err
		}
	}
	return &m, nil
}

// countResources tallies descriptor-bound resources per SDL_GPU's SPIR-V
// convention: a vertex shader puts sampled textures, storage textures and
// storage buffers in set 0 and uniform buffers in set 1; a fragment shader
// uses sets 2 and 3 for the same two groups. A resource in any other set
// means the shader was not compiled for SDL, so it is reported rather than
// silently miscounted.
func (r *reflector) countResources(m *Metadata) error {
	resourceSet, uniformSet := uint32(0), uint32(1)
	if m.Stage == StageFragment {
		resourceSet, uniformSet = 2, 3
	}

	for _, id := range r.varOrder {
		set, bound := r.dset[id]
		if !bound {
			continue // not a descriptor resource
		}
		cat, count, err := r.classify(id)
		if err != nil {
			return err
		}

		want := resourceSet
		switch cat {
		case catSampler:
			// A bare HLSL SamplerState. SDL counts texture-sampler pairs, and
			// the texture half is counted below, so counting this too would
			// double it.
			continue
		case catUniformBuffer:
			want = uniformSet
		}
		if set != want {
			return fmt.Errorf(
				"shadermeta: %s shader has a resource at binding %d in descriptor set %d, but SDL expects it in set %d",
				m.Stage, r.binding[id], set, want)
		}

		switch cat {
		case catSampledTexture:
			m.NumSamplers += count
		case catStorageTexture:
			m.NumStorageTextures += count
		case catStorageBuffer:
			m.NumStorageBuffers += count
		case catUniformBuffer:
			m.NumUniformBuffers += count
		}
	}
	return nil
}

// classify resolves what a descriptor-bound variable is, and how many
// descriptors it occupies (more than one when it is an array).
func (r *reflector) classify(id uint32) (category, uint32, error) {
	ptr, ok := r.types[r.varPtr[id]]
	if !ok || ptr.kind != kindPointer {
		return 0, 0, fmt.Errorf("shadermeta: variable %%%d has no pointer type", id)
	}

	count := uint32(1)
	target := ptr.elem
	for {
		t, ok := r.types[target]
		if !ok {
			return 0, 0, fmt.Errorf("shadermeta: variable %%%d refers to unknown type %%%d", id, target)
		}
		switch t.kind {
		case kindArray:
			length, ok := r.consts[t.count]
			if !ok {
				return 0, 0, fmt.Errorf("shadermeta: array type %%%d has a non-constant length", target)
			}
			count *= length
			target = t.elem
			continue
		case kindRuntimeArray:
			target = t.elem
			continue

		case kindSampledImage:
			return catSampledTexture, count, nil
		case kindSampler:
			return catSampler, count, nil
		case kindImage:
			if t.sampled == 2 {
				return catStorageTexture, count, nil
			}
			return catSampledTexture, count, nil
		case kindStruct:
			if ptr.storage == storageStorageBuffer || r.bufferBlock[target] {
				return catStorageBuffer, count, nil
			}
			if ptr.storage == storageUniform || ptr.storage == storageUniformConstant {
				return catUniformBuffer, count, nil
			}
			return 0, 0, fmt.Errorf("shadermeta: struct resource %%%d in unexpected storage class %d", target, ptr.storage)
		default:
			return 0, 0, fmt.Errorf("shadermeta: unsupported resource type for variable %%%d", id)
		}
	}
}

// collectInputs records the shader's declared vertex inputs, skipping
// built-ins such as gl_VertexIndex, which are not fed from a vertex buffer.
func (r *reflector) collectInputs(m *Metadata) error {
	for _, id := range r.varOrder {
		ptr, ok := r.types[r.varPtr[id]]
		if !ok || ptr.kind != kindPointer || ptr.storage != storageInput {
			continue
		}
		if r.builtin[id] {
			continue
		}
		loc, ok := r.location[id]
		if !ok {
			continue // an input without a location is not a vertex attribute
		}

		t, ok := r.types[ptr.elem]
		if !ok {
			return fmt.Errorf("shadermeta: input at location %d refers to unknown type", loc)
		}
		components := uint32(1)
		if t.kind == kindVector {
			components = t.count
			if t, ok = r.types[t.elem]; !ok {
				return fmt.Errorf("shadermeta: input at location %d has an unknown component type", loc)
			}
		}

		var scalar ScalarType
		switch t.kind {
		case kindFloat:
			scalar = ScalarFloat
		case kindInt:
			scalar = ScalarInt
			if !t.signed {
				scalar = ScalarUint
			}
		default:
			return fmt.Errorf("shadermeta: input at location %d has an unsupported component type", loc)
		}

		m.Inputs = append(m.Inputs, Input{Location: loc, Type: scalar, Components: components})
	}

	slices.SortFunc(m.Inputs, func(a, b Input) int { return int(a.Location) - int(b.Location) })
	return nil
}

// decodeWords reads a SPIR-V module, accepting either byte order.
func decodeWords(code []byte) ([]uint32, error) {
	if len(code)%4 != 0 {
		return nil, fmt.Errorf("shadermeta: bytecode length %d is not a multiple of 4", len(code))
	}
	if len(code) < 20 {
		return nil, fmt.Errorf("shadermeta: bytecode is too short to be a SPIR-V module")
	}

	order := binary.ByteOrder(binary.LittleEndian)
	switch {
	case binary.LittleEndian.Uint32(code) == spirvMagic:
	case binary.BigEndian.Uint32(code) == spirvMagic:
		order = binary.BigEndian
	default:
		return nil, fmt.Errorf("shadermeta: not a SPIR-V module (magic 0x%08x)", binary.LittleEndian.Uint32(code))
	}

	words := make([]uint32, len(code)/4)
	for i := range words {
		words[i] = order.Uint32(code[i*4:])
	}
	return words, nil
}

// decodeString reads a null-terminated UTF-8 literal packed four bytes per
// word, as SPIR-V stores names.
func decodeString(words []uint32) string {
	var b []byte
	for _, w := range words {
		for s := 0; s < 4; s++ {
			c := byte(w >> (8 * s))
			if c == 0 {
				return string(b)
			}
			b = append(b, c)
		}
	}
	return string(b)
}
