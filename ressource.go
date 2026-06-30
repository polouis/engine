package engine

import (
	"encoding/json"
	"fmt"

	"github.com/polouis/engine/types"
)

// /*
//   - Need a table describing the VB when shared between multiple entities
//   - - mesh id/name
//   - -
//  */

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

type RessourceMeshId struct {
	Id string `json:"id"`
}

type RessourceName struct {
	Name string `json:"name"`
}

type RessourceTransform struct {
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
	meshRessources   map[string]*RessourceMesh
	entityRessources map[string]*ressourceEntity
}

func NewRessourceManager() *RessourceManager {
	return &RessourceManager{
		meshRessources:   make(map[string]*RessourceMesh),
		entityRessources: make(map[string]*ressourceEntity),
	}
}

func (rm *RessourceManager) LoadMesh(dat []byte) error {
	var mesh RessourceMesh
	if err := json.Unmarshal(dat, &mesh); err != nil {
		return fmt.Errorf("Cannot deserialize mesh", err)
	}
	_, exists := rm.meshRessources[mesh.Id]
	if exists {
		return fmt.Errorf("mesh Id %s already exists", mesh.Id)
	}
	rm.meshRessources[mesh.Id] = &mesh
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
		fmt.Println(c.Type)
		switch c.Type {
		case "mesh":
			var rsMesh RessourceMeshId
			if err := json.Unmarshal(c.Value, &rsMesh); err != nil {
				return fmt.Errorf("cannot deserialize entity's mesh component : %w", err)
			}
			fmt.Println(rsMesh.Id)
		case "name":
			var rsName RessourceName
			if err := json.Unmarshal(c.Value, &rsName); err != nil {
				return fmt.Errorf("cannot deserialize entity's name component : %w", err)
			}
			fmt.Println(rsName.Name)
		case "transform":
			var rsTransform RessourceTransform
			if err := json.Unmarshal(c.Value, &rsTransform); err != nil {
				return fmt.Errorf("cannot deserialize entity's transform component : %w", err)
			}
			fmt.Println(rsTransform)
		}
	}

	rm.entityRessources[id] = &ent

	return nil
}

func (rm *RessourceManager) Spawn(ctx *Context, id string) (EntityID, error) {
	ent, exists := rm.entityRessources[id]
	if !exists {
		return 0, fmt.Errorf("cannot spawn entity : id %s not found", id)
	}

	e := ctx.W.NewEntity()
	for _, c := range ent.Components {
		switch c.Type {
		case "name":
			var r RessourceName
			json.Unmarshal(c.Value, &r)
			ctx.W.NameStore.Upsert(e, NameComponent{Name: r.Name})
		case "transform":
			var r RessourceTransform
			json.Unmarshal(c.Value, &r)
			ctx.W.TransformStore.Upsert(e, TransformComponent{
				Position: Vector3{X: r.X, Y: r.Y, Z: r.Z},
			})
		case "mesh":
			var r RessourceMeshId
			json.Unmarshal(c.Value, &r)
			mesh := rm.meshRessources[r.Id] // resolve from cache
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
			ctx.W.MeshStore.Upsert(e, NewMeshComponent(ctx, verts))
		}
	}
	return e, nil
}
