package shadermeta

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func reflectFile(t *testing.T, name string) *Metadata {
	t.Helper()
	code, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	m, err := Reflect(code)
	if err != nil {
		t.Fatalf("Reflect(%s): %v", name, err)
	}
	return m
}

// The counts asserted here are the ones that used to be hardcoded as the
// literal arguments 0, 1, 0, 0 and 0, 0, 0, 0 at the shader creation calls.
func TestReflectVertexShader(t *testing.T) {
	m := reflectFile(t, "PositionColor.vert.spv")

	if m.Stage != StageVertex {
		t.Errorf("stage = %q, want %q", m.Stage, StageVertex)
	}
	if m.Entrypoint != "main" {
		t.Errorf("entrypoint = %q, want %q", m.Entrypoint, "main")
	}
	if got := [4]uint32{m.NumSamplers, m.NumStorageTextures, m.NumStorageBuffers, m.NumUniformBuffers}; got != [4]uint32{0, 0, 0, 1} {
		t.Errorf("counts = %v, want [0 0 0 1]", got)
	}
}

func TestReflectFragmentShader(t *testing.T) {
	m := reflectFile(t, "SolidColor.frag.spv")

	if m.Stage != StageFragment {
		t.Errorf("stage = %q, want %q", m.Stage, StageFragment)
	}
	if got := [4]uint32{m.NumSamplers, m.NumStorageTextures, m.NumStorageBuffers, m.NumUniformBuffers}; got != [4]uint32{0, 0, 0, 0} {
		t.Errorf("counts = %v, want [0 0 0 0]", got)
	}
	// Fragment inputs are interpolated varyings, not vertex attributes.
	if m.Inputs != nil {
		t.Errorf("Inputs = %v, want none for a fragment shader", m.Inputs)
	}
}

// The vertex inputs are the shader's half of the vertex-format contract, so
// they must come back with their real component counts rather than being
// assumed to match the mesh.
func TestReflectVertexInputs(t *testing.T) {
	m := reflectFile(t, "PositionColor.vert.spv")

	want := []Input{
		{Location: 0, Type: ScalarFloat, Components: 3}, // a_position : POSITION
		{Location: 1, Type: ScalarFloat, Components: 4}, // a_color    : COLOR
	}
	if !slices.Equal(m.Inputs, want) {
		t.Errorf("Inputs = %+v, want %+v", m.Inputs, want)
	}
}

func TestSidecarPathIsFormatIndependent(t *testing.T) {
	for _, in := range []string{
		"dist/asset/shader/PositionColor.vert.spv",
		"dist/asset/shader/PositionColor.vert.msl",
		"dist/asset/shader/PositionColor.vert.dxil",
	} {
		if got, want := SidecarPath(in), "dist/asset/shader/PositionColor.vert.json"; got != want {
			t.Errorf("SidecarPath(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRoundTripThroughSidecar(t *testing.T) {
	want := reflectFile(t, "PositionColor.vert.spv")

	path := filepath.Join(t.TempDir(), "PositionColor.vert.json")
	if err := want.WriteFile(path); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Stage != want.Stage || got.NumUniformBuffers != want.NumUniformBuffers ||
		!slices.Equal(got.Inputs, want.Inputs) {
		t.Errorf("round trip changed the metadata:\n got %+v\nwant %+v", got, want)
	}
}

func TestReflectRejectsNonSPIRV(t *testing.T) {
	for name, code := range map[string][]byte{
		"empty":      {},
		"short":      []byte("abcd"),
		"unaligned":  make([]byte, 21),
		"bad magic":  make([]byte, 40),
		"truncated?": append([]byte{0x03, 0x02, 0x23, 0x07}, make([]byte, 16)...),
	} {
		if _, err := Reflect(code); err == nil {
			t.Errorf("Reflect(%s) succeeded, want an error", name)
		}
	}
}

func TestLoadRejectsUnknownStage(t *testing.T) {
	path := filepath.Join(t.TempDir(), "weird.json")
	if err := os.WriteFile(path, []byte(`{"stage":"geometry"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Load(path)
	if err == nil {
		t.Fatal("Load accepted an unknown stage")
	}
	if !strings.Contains(err.Error(), "geometry") {
		t.Errorf("error %q does not name the offending stage", err)
	}
}
