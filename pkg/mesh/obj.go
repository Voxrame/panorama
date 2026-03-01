package mesh

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/lord-server/panorama/pkg/lm"
)

func parseVector3(fields []string) (lm.Vector3, error) {
	if len(fields) < 3 {
		return lm.Vector3{}, fmt.Errorf("expected at least 3 vector elements, found %d", len(fields))
	}

	x, err := strconv.ParseFloat(fields[0], 32)
	if err != nil {
		return lm.Vector3{}, err
	}

	y, err := strconv.ParseFloat(fields[1], 32)
	if err != nil {
		return lm.Vector3{}, err
	}

	z, err := strconv.ParseFloat(fields[2], 32)
	if err != nil {
		return lm.Vector3{}, err
	}

	return lm.Vec3(x, y, z), nil
}

func parseVector2(fields []string) (lm.Vector2, error) {
	if len(fields) < 2 {
		return lm.Vector2{}, fmt.Errorf("expected at least 3 vector elements, found %d", len(fields))
	}

	x, err := strconv.ParseFloat(fields[0], 32)
	if err != nil {
		return lm.Vector2{}, err
	}

	y, err := strconv.ParseFloat(fields[1], 32)
	if err != nil {
		return lm.Vector2{}, err
	}

	return lm.Vec2(x, y), nil
}

type Triplet struct {
	positionIndex int
	texcoordIndex *int
	normalIndex   *int
}

func parseFace(fields []string) ([]Triplet, error) {
	if len(fields) < 3 {
		return []Triplet{}, fmt.Errorf("face needs at least 3 fields, got %v", len(fields))
	}

	triplets := []Triplet{}

	for _, field := range fields {
		parts := strings.SplitN(field, "/", 3)

		var err error

		triplet := Triplet{}

		triplet.positionIndex, err = strconv.Atoi(parts[0])
		if err != nil {
			return []Triplet{}, err
		}

		if len(parts) > 1 && len(parts[1]) != 0 {
			texcoordIndex, err := strconv.Atoi(parts[1])
			if err != nil {
				return []Triplet{}, err
			}

			triplet.texcoordIndex = &texcoordIndex
		}

		if len(parts) > 2 && len(parts[2]) != 0 {
			normalIndex, err := strconv.Atoi(parts[2])
			if err != nil {
				return []Triplet{}, err
			}

			triplet.normalIndex = &normalIndex
		}

		triplets = append(triplets, triplet)
	}

	return triplets, nil
}

type objParser struct {
	positions []lm.Vector3
	texcoords []lm.Vector2
	normals   []lm.Vector3

	mesh Mesh
}

func (o *objParser) vertexAt(triplet Triplet) Vertex {
	texcoord := lm.Vector2{}
	normal := lm.Vector3{}

	if triplet.texcoordIndex != nil {
		texcoord = o.texcoords[*triplet.texcoordIndex-1]
	}

	if triplet.normalIndex != nil {
		normal = o.normals[*triplet.normalIndex-1]
	}

	return Vertex{
		Position: o.positions[triplet.positionIndex-1],
		Texcoord: texcoord,
		Normal:   normal,
	}
}

func (o *objParser) triangulatePolygon(triplets []Triplet) []Vertex {
	vertices := []Vertex{}

	origin := o.vertexAt(triplets[0])

	for i := 2; i < len(triplets); i++ {
		vertices = append(vertices, origin, o.vertexAt(triplets[i-1]), o.vertexAt(triplets[i]))
	}

	return vertices
}

func (o *objParser) processLine(line string) error {
	// Skip comments
	if strings.HasPrefix(line, "#") {
		return nil
	}

	fields := strings.Fields(line)

	// Skip empty lines
	if len(fields) == 0 {
		return nil
	}

	switch fields[0] {
	case "v":
		position, err := parseVector3(fields[1:])
		if err != nil {
			return err
		}

		o.positions = append(o.positions, position)

	case "vt":
		texcoord, err := parseVector2(fields[1:])
		if err != nil {
			return err
		}

		o.texcoords = append(o.texcoords, texcoord)

	case "vn":
		normal, err := parseVector3(fields[1:])
		if err != nil {
			return err
		}

		o.normals = append(o.normals, normal)

	case "f":
		triplets, err := parseFace(fields[1:])
		if err != nil {
			return err
		}

		vertices := o.triangulatePolygon(triplets)

		o.mesh.Vertices = append(o.mesh.Vertices, vertices...)
	}

	return nil
}

func LoadOBJ(path string) (Model, error) {
	file, err := os.Open(path)
	if err != nil {
		return Model{}, err
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("error closing file: %v", err)
		}
	}()

	scanner := bufio.NewScanner(file)
	parser := objParser{
		positions: []lm.Vector3{},
		texcoords: []lm.Vector2{},
		normals:   []lm.Vector3{},
		mesh:      NewMesh(),
	}

	lineNumber := 1

	for scanner.Scan() {
		lineNumber += 1

		err := parser.processLine(scanner.Text())
		if err != nil {
			return Model{}, err
		}
	}

	if err := scanner.Err(); err != nil {
		return Model{}, err
	}

	model := NewModel()
	model.Meshes = append(model.Meshes, parser.mesh)

	return model, nil
}
