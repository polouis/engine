package engine

import (
	"encoding/json"
	"fmt"

	"github.com/polouis/engine/internal"
	"github.com/polouis/engine/internal/backend"
	"github.com/polouis/engine/types"
)

type RessourceVertex struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
	Z float32 `json:"z"`
	R uint8   `json:"r"`
	G uint8   `json:"g"`
	B uint8   `json:"b"`
	A uint8   `json:"a"`
}

type ressourceMesh struct {
	Name         string            `json:"name"`
	Vertices     []RessourceVertex `json:"vertices"`
	VertexFormat string            `json:"vertex_format"`
}

type ressourceMaterial struct {
	Name   string `json:"name"`
	Effect string `json:"effect"`
}

type RessourceMeshComponent struct {
	Mesh     string `json:"mesh"`
	Material string `json:"material"`
}

type RessourceNameComponent struct {
	Name string `json:"name"`
}

type RessourceTransformComponent struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
	Z float32 `json:"z"`
}

type rawComponentEntry struct {
	Type  string          `json:"type"`
	Value json.RawMessage `json:"value"`
}

type ressourceEntity struct {
	name      *RessourceNameComponent
	mesh      *RessourceMeshComponent
	transform *RessourceTransformComponent
}

type rawRessourceEntity struct {
	Components []rawComponentEntry `json:"components"`
}

type ressourceEffect struct {
	name                 string
	vertexShader         backend.ShaderID
	fragmentShader       backend.ShaderID
	requiresVertexFormat types.VertexFormat
}

type RessourceManager struct {
	meshRessources     map[string]*ressourceMesh
	entityRessources   map[string]*ressourceEntity
	vertexBuffers      map[string]backend.VertexBufferID
	materialRessources internal.Registry[ressourceMaterial, types.MaterialID]
	effects            internal.Registry[ressourceEffect, types.EffectID]
	pipelines          map[backend.PipelineDesc]backend.PipelineID
}

func NewRessourceManager() *RessourceManager {
	return &RessourceManager{
		meshRessources:     make(map[string]*ressourceMesh),
		entityRessources:   make(map[string]*ressourceEntity),
		vertexBuffers:      make(map[string]backend.VertexBufferID),
		effects:            internal.Registry[ressourceEffect, types.EffectID]{},
		materialRessources: internal.Registry[ressourceMaterial, types.MaterialID]{},
		pipelines:          make(map[backend.PipelineDesc]backend.PipelineID),
	}
}

func (rm *RessourceManager) LoadMesh(dat []byte) error {
	var mesh ressourceMesh
	if err := json.Unmarshal(dat, &mesh); err != nil {
		return fmt.Errorf("Cannot deserialize mesh, got error %w", err)
	}
	_, exists := rm.meshRessources[mesh.Name]
	if exists {
		return fmt.Errorf("mesh Id %s already exists", mesh.Name)
	}
	rm.meshRessources[mesh.Name] = &mesh
	return nil
}

func (rm *RessourceManager) LoadEffect(ctx *Context, dat []byte) error {
	desc, err := LoadEffect(dat)
	if err != nil {
		return fmt.Errorf("Cannot deserialize effect, got error %w", err)
	}
	if desc.RequiresVertexFormat == "" {
		return fmt.Errorf("effect %s does not declare requires_vertex_format", desc.Name)
	}
	rvf, err := types.ParseVertexFormat(desc.RequiresVertexFormat)
	if err != nil {
		return fmt.Errorf("effect %s: %w", desc.Name, err)
	}

	vertexShader := ctx.gpu.CreateShader(desc.VertexShader)
	fragmentShader := ctx.gpu.CreateShader(desc.FragmentShader)
	effect := ressourceEffect{
		name:                 desc.Name,
		vertexShader:         vertexShader,
		fragmentShader:       fragmentShader,
		requiresVertexFormat: rvf,
	}
	if _, err := rm.effects.Add(desc.Name, &effect); err != nil {
		ctx.gpu.ReleaseShader(vertexShader)
		ctx.gpu.ReleaseShader(fragmentShader)
		return err
	}
	return nil
}

func (rm *RessourceManager) LoadMaterial(ctx *Context, dat []byte) error {
	var material ressourceMaterial
	if err := json.Unmarshal(dat, &material); err != nil {
		return fmt.Errorf("Cannot deserialize material, got error %w", err)
	}

	_, err := rm.materialRessources.Add(material.Name, &material)
	return err
}

func resolvePipelineDesc(e *ressourceEffect, m *ressourceMaterial, vf types.VertexFormat, p *backend.PassDef) backend.PipelineDesc {
	return backend.PipelineDesc{
		VertexShader:   e.vertexShader,
		FragmentShader: e.fragmentShader,
		// The mesh's actual layout, not the effect's requirement: this is what
		// the pipeline's vertex input state is built from, so it has to
		// describe the buffer that will be bound. Spawn is what guarantees the
		// two agree.
		VertexFormat: vf,
	}
}

func (rm *RessourceManager) pipelineFor(
	ctx *Context, m *ressourceMaterial, vf types.VertexFormat, p *backend.PassDef,
) backend.PipelineID {
	effectId, _ := rm.effects.ID(m.Effect)
	e := rm.effects.Get(effectId)
	if p.EffectOverride != types.InvalidEffectId {
		e = rm.effects.Get(p.EffectOverride)
	}

	desc := resolvePipelineDesc(e, m, vf, p) // pure function
	if id, ok := rm.pipelines[desc]; ok {
		return id
	}
	id := ctx.gpu.CreatePipeline(desc)
	rm.pipelines[desc] = id
	return id
}

func (rm *RessourceManager) LoadEntity(id string, dat []byte) error {
	_, exists := rm.entityRessources[id]
	if exists {
		return fmt.Errorf("cannot load entity : id %s already exists", id)
	}

	var rawEnt rawRessourceEntity
	if err := json.Unmarshal(dat, &rawEnt); err != nil {
		return fmt.Errorf("cannot deserialize entity : %w", err)
	}
	var ent ressourceEntity

	for _, c := range rawEnt.Components {
		switch c.Type {
		case "mesh":
			var rsMesh RessourceMeshComponent
			if err := json.Unmarshal(c.Value, &rsMesh); err != nil {
				return fmt.Errorf("cannot deserialize entity's mesh component : %w", err)
			}
			ent.mesh = &rsMesh
		case "name":
			var rsName RessourceNameComponent
			if err := json.Unmarshal(c.Value, &rsName); err != nil {
				return fmt.Errorf("cannot deserialize entity's name component : %w", err)
			}
			ent.name = &rsName
		case "transform":
			var rsTransform RessourceTransformComponent
			if err := json.Unmarshal(c.Value, &rsTransform); err != nil {
				return fmt.Errorf("cannot deserialize entity's transform component : %w", err)
			}
			ent.transform = &rsTransform
		}
	}

	rm.entityRessources[id] = &ent

	return nil
}

func (rm *RessourceManager) Release(ctx *Context) {
	for _, pipelineId := range rm.pipelines {
		ctx.gpu.ReleasePipeline(pipelineId)
	}
	for _, effect := range rm.effects.All() {
		ctx.gpu.ReleaseShader(effect.fragmentShader)
		ctx.gpu.ReleaseShader(effect.vertexShader)
	}
}

func (rm *RessourceManager) Spawn(ctx *Context, id string) (EntityID, error) {
	ent, exists := rm.entityRessources[id]
	if !exists {
		return InvalidEntityID, fmt.Errorf("cannot spawn entity : id %s not found", id)
	}

	e := ctx.W.NewEntity()
	if ent.name != nil {
		ctx.W.NameStore.Upsert(e, NameComponent{Name: ent.name.Name})
	}
	if ent.transform != nil {
		ctx.W.TransformStore.Upsert(e, TransformComponent{
			Position: Vector3{X: ent.transform.X, Y: ent.transform.Y, Z: ent.transform.Z},
		})
	}
	if ent.mesh != nil {
		mesh := rm.meshRessources[ent.mesh.Mesh]
		if mesh == nil {
			return InvalidEntityID, fmt.Errorf("Cannot instanciate entity %s referencing unknown mesh id %s", id, ent.mesh.Mesh)
		}
		var verts []types.PositionColorVertex = make([]types.PositionColorVertex, len(mesh.Vertices))
		for i, v := range mesh.Vertices {
			verts[i].X = v.X
			verts[i].Y = v.Y
			verts[i].Z = v.Z
			verts[i].R = v.R
			verts[i].G = v.G
			verts[i].B = v.B
			verts[i].A = v.A
		}
		matId, exists := rm.materialRessources.ID(ent.mesh.Material)
		if !exists {
			return InvalidEntityID, fmt.Errorf("Cannot instanciate entity %s referencing unknown material id %s", id, ent.mesh.Material)
		}
		mat := rm.materialRessources.Get(matId)
		effectId, exists := rm.effects.ID(mat.Effect)
		if !exists {
			return InvalidEntityID, fmt.Errorf("Cannot instanciate entity %s referencing unknown effect id %s", id, mat.Effect)
		}
		effect := rm.effects.Get(effectId)

		vf, err := types.ParseVertexFormat(mesh.VertexFormat)
		if err != nil {
			return InvalidEntityID, err
		}
		if vf != effect.requiresVertexFormat {
			return InvalidEntityID, fmt.Errorf(
				"cannot instanciate entity %s: mesh %s has vertex format %s, but material %s uses effect %s, which requires %s",
				id, ent.mesh.Mesh, vf, ent.mesh.Material, mat.Effect, effect.requiresVertexFormat)
		}
		ctx.W.MeshStore.Upsert(e, NewMeshComponent(ctx, verts, matId, vf))
	}

	return e, nil
}
