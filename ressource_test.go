package engine

import (
	"strings"
	"testing"

	backenddummy "github.com/polouis/engine/internal/backend_dummy"
	"github.com/polouis/engine/types"
)

func newTestContext() *Context {
	b := &backenddummy.BackendDummy{}
	return &Context{W: NewWorld(), RM: NewRessourceManager(), p: b, gpu: b, i: b}
}

const (
	testEffectJSON = `{
		"name": "solid_color",
		"vertex_shader": "PositionColor.vert",
		"fragment_shader": "SolidColor.frag",
		"requires_vertex_format": "position_color"
	}`
	testMaterialJSON = `{"name": "mat-001", "effect": "solid_color"}`
	testMeshJSON     = `{
		"name": "sq-01",
		"vertices": [{"x": 0, "y": 0, "z": 0, "r": 255, "g": 0, "b": 0, "a": 255}],
		"vertex_format": "position_color"
	}`
	testEntityJSON = `{"components": [
		{"type": "mesh", "value": {"mesh": "sq-01", "material": "mat-001"}}
	]}`
)

// loadSpawnableAssets loads the mesh/effect/material/entity chain that Spawn
// walks, so each test only has to perturb the one asset it is about.
func loadSpawnableAssets(t *testing.T, ctx *Context, effectJSON string) {
	t.Helper()
	if err := ctx.RM.LoadMesh([]byte(testMeshJSON)); err != nil {
		t.Fatalf("LoadMesh: %v", err)
	}
	if err := ctx.RM.LoadEffect(ctx, []byte(effectJSON)); err != nil {
		t.Fatalf("LoadEffect: %v", err)
	}
	if err := ctx.RM.LoadMaterial(ctx, []byte(testMaterialJSON)); err != nil {
		t.Fatalf("LoadMaterial: %v", err)
	}
	if err := ctx.RM.LoadEntity("square", []byte(testEntityJSON)); err != nil {
		t.Fatalf("LoadEntity: %v", err)
	}
}

// An effect's requires_vertex_format is parsed when the effect is loaded, so
// an unknown or missing name is reported against the asset that holds it
// rather than against whichever entity happens to reference it first.
func TestLoadEffectValidatesRequiredVertexFormat(t *testing.T) {
	for name, effectJSON := range map[string]string{
		"unknown format": `{"name": "e", "vertex_shader": "v", "fragment_shader": "f",
			"requires_vertex_format": "position_normal_color"}`,
		"missing format": `{"name": "e", "vertex_shader": "v", "fragment_shader": "f"}`,
	} {
		ctx := newTestContext()
		err := ctx.RM.LoadEffect(ctx, []byte(effectJSON))
		if err == nil {
			t.Errorf("%s: LoadEffect succeeded, want an error", name)
			continue
		}
		if !strings.Contains(err.Error(), "e") {
			t.Errorf("%s: error %q does not name the effect", name, err)
		}
		// A rejected effect must not be registered, or a later material could
		// resolve it and reach a pipeline with no usable vertex layout.
		if _, ok := ctx.RM.effects.ID("e"); ok {
			t.Errorf("%s: the rejected effect was registered anyway", name)
		}
	}
}

func TestSpawnAcceptsMatchingVertexFormat(t *testing.T) {
	ctx := newTestContext()
	loadSpawnableAssets(t, ctx, testEffectJSON)

	e, err := ctx.RM.Spawn(ctx, "square")
	if err != nil {
		t.Fatalf("Spawn: %v", err)
	}
	mesh, err := ctx.W.MeshStore.Get(e)
	if err != nil {
		t.Fatalf("spawned entity has no mesh component: %v", err)
	}
	if mesh.vertexFormat != types.VertexFormatPositionColor {
		t.Errorf("vertexFormat = %s, want %s", mesh.vertexFormat, types.VertexFormatPositionColor)
	}
}

// A mesh whose layout the effect's shaders were not compiled against must be
// refused at spawn time: the pipeline is built from the mesh's format, so the
// GPU would otherwise be handed a buffer the shader cannot read, which is
// undefined rather than a reported failure.
//
// Only one vertex format exists so far, so the requirement is forced to a
// different value here instead of being authored in JSON. Replace this with a
// second real format once there is one.
func TestSpawnRejectsVertexFormatMismatch(t *testing.T) {
	ctx := newTestContext()
	loadSpawnableAssets(t, ctx, testEffectJSON)

	effectId, ok := ctx.RM.effects.ID("solid_color")
	if !ok {
		t.Fatal("effect solid_color is not registered")
	}
	ctx.RM.effects.Get(effectId).requiresVertexFormat = types.VertexFormatInvalid

	_, err := ctx.RM.Spawn(ctx, "square")
	if err == nil {
		t.Fatal("Spawn accepted a mesh in a format the effect does not take")
	}
	// The message has to name the assets to fix, not just the formats.
	for _, want := range []string{"sq-01", "mat-001", "solid_color", "position_color"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not mention %q", err, want)
		}
	}
}
