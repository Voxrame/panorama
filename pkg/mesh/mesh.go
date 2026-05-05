package mesh

import (
	"github.com/Voxrame/panorama/pkg/lm"
)

type Vertex struct {
	Position lm.Vector3
	Texcoord lm.Vector2
	Normal   lm.Vector3
}

type Mesh struct {
	Vertices []Vertex
}

func NewMesh() Mesh {
	return Mesh{
		Vertices: []Vertex{},
	}
}

type Model struct {
	Meshes []Mesh
}

func NewModel() Model {
	return Model{
		Meshes: []Mesh{},
	}
}
