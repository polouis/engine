package engine

import (
	"encoding/json"
	"fmt"

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

type RessourceMesh struct {
	Id       string            `json:"id"`
	Vertices []RessourceVertex `json:"vertices"`
}

type RessourceMaterial struct {
	Id             string `json:"id"`
	VertexShader   string `json:"vertex_shader"`
	FragmentShader string `json:"fragment_shader"`
	VertexFormat   string `json:"vertex_format"`
}

type RessourceMeshComponent struct {
	MeshId     string `json:"mesh_id"`
	MaterialId string `json:"material_id"`
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
	Components []rawComponentEntry `json:"components"`
}

type RessourceManager struct {
	meshRessources     map[string]*RessourceMesh
	materialRessources map[string]*RessourceMaterial
	entityRessources   map[string]*ressourceEntity
	pipelines          map[string]backend.PipelineID
	vertexBuffers      map[string]backend.VertexBufferID
}

func NewRessourceManager() *RessourceManager {
	return &RessourceManager{
		meshRessources:     make(map[string]*RessourceMesh),
		materialRessources: make(map[string]*RessourceMaterial),
		entityRessources:   make(map[string]*ressourceEntity),
		pipelines:          make(map[string]backend.PipelineID),
		vertexBuffers:      make(map[string]backend.VertexBufferID),
	}
}

func (rm *RessourceManager) LoadMesh(dat []byte) error {
	var mesh RessourceMesh
	if err := json.Unmarshal(dat, &mesh); err != nil {
		return fmt.Errorf("Cannot deserialize mesh, got error %w", err)
	}
	_, exists := rm.meshRessources[mesh.Id]
	if exists {
		return fmt.Errorf("mesh Id %s already exists", mesh.Id)
	}
	rm.meshRessources[mesh.Id] = &mesh
	return nil
}

func (rm *RessourceManager) LoadMaterial(dat []byte) error {
	var material RessourceMaterial
	if err := json.Unmarshal(dat, &material); err != nil {
		return fmt.Errorf("Cannot deserialize material, got error %w", err)
	}
	_, exists := rm.materialRessources[material.Id]
	if exists {
		return fmt.Errorf("material Id %s already exists", material.Id)
	}
	rm.materialRessources[material.Id] = &material
	return nil
}

func (rm *RessourceManager) LoadEntity(id string, dat []byte) error {
	_, exists := rm.entityRessources[id]
	if exists {
		return fmt.Errorf("cannot load entity : id %s already exists", id)
	}

	var ent ressourceEntity
	if err := json.Unmarshal(dat, &ent); err != nil {
		return fmt.Errorf("cannot deserialize entity : %w", err)
	}

	for _, c := range ent.Components {
		switch c.Type {
		case "mesh":
			var rsMesh RessourceMeshComponent
			if err := json.Unmarshal(c.Value, &rsMesh); err != nil {
				return fmt.Errorf("cannot deserialize entity's mesh component : %w", err)
			}
		case "name":
			var rsName RessourceNameComponent
			if err := json.Unmarshal(c.Value, &rsName); err != nil {
				return fmt.Errorf("cannot deserialize entity's name component : %w", err)
			}
		case "transform":
			var rsTransform RessourceTransformComponent
			if err := json.Unmarshal(c.Value, &rsTransform); err != nil {
				return fmt.Errorf("cannot deserialize entity's transform component : %w", err)
			}
		}
	}

	rm.entityRessources[id] = &ent

	return nil
}

func (rm *RessourceManager) LoadPipeline(ctx *Context, dat []byte) error {

	// TODO load pipelines configuration from file. Hard coded for now.
	pipelineId := ctx.gpu.CreatePipeline(backend.PipelineDesc{})
	rm.pipelines["the pipeline"] = pipelineId

	return nil
}

func (rm *RessourceManager) ReleasePipelines(ctx *Context) {
	for _, pipelineId := range rm.pipelines {
		ctx.gpu.ReleasePipeline(pipelineId)
	}
}

func (rm *RessourceManager) Spawn(ctx *Context, id string) (EntityID, error) {
	ent, exists := rm.entityRessources[id]
	if !exists {
		return InvalidEntityID, fmt.Errorf("cannot spawn entity : id %s not found", id)
	}

	e := ctx.W.NewEntity()
	for _, c := range ent.Components {
		switch c.Type {
		case "name":
			var r RessourceNameComponent
			json.Unmarshal(c.Value, &r)
			ctx.W.NameStore.Upsert(e, NameComponent{Name: r.Name})
		case "transform":
			var r RessourceTransformComponent
			json.Unmarshal(c.Value, &r)
			ctx.W.TransformStore.Upsert(e, TransformComponent{
				Position: Vector3{X: r.X, Y: r.Y, Z: r.Z},
			})
		case "mesh":
			var r RessourceMeshComponent
			json.Unmarshal(c.Value, &r)
			mesh := rm.meshRessources[r.MeshId]
			if mesh == nil {
				return InvalidEntityID, fmt.Errorf("Cannot instanciate entity %s referencing unknown mesh id %s", id, r.MeshId)
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
			material := rm.materialRessources[r.MaterialId]
			if material == nil {
				return InvalidEntityID, fmt.Errorf("Cannot instanciate entity %s referencing unknown material id %s", id, r.MaterialId)
			}

			pipelineName := "the pipeline"
			pipelineId := rm.pipelines[pipelineName]

			ctx.W.MeshStore.Upsert(e, NewMeshComponent(ctx, verts, pipelineId))
		}
	}
	return e, nil
}
