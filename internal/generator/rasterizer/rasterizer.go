package rasterizer

import (
	"image"
	"image/color"
	"math"

	"github.com/Voxrame/panorama/internal/game"
	"github.com/Voxrame/panorama/pkg/lm"
	"github.com/Voxrame/panorama/pkg/mesh"
)

const Gamma = 2.2
const BaseResolution = 16

type RenderableNode struct {
	Name        string
	Light       float64
	Param2      uint8
	HiddenFaces mesh.CubeFaces
}

type NodeRasterizer struct {
	cache map[RenderableNode]*RenderBuffer

	projection lm.Matrix3
}

func New(projection lm.Matrix3) NodeRasterizer {
	return NodeRasterizer{
		cache: make(map[RenderableNode]*RenderBuffer),

		projection: projection,
	}
}

func cartesianToBarycentric(p lm.Vector2, a, b, c lm.Vector2) lm.Vector3 {
	u := lm.Vec3(c.X-a.X, b.X-a.X, a.X-p.X)
	v := lm.Vec3(c.Y-a.Y, b.Y-a.Y, a.Y-p.Y)
	w := u.Cross(v)

	return lm.Vec3(1-(w.X+w.Y)/w.Z, w.Y/w.Z, w.X/w.Z)
}

func sampleTriangle(x, y int, a, b, c lm.Vector2) (bool, lm.Vector3) {
	p := lm.Vec2(float64(x), float64(y))

	samplePointOffset := lm.Vec2(0.5, 0.5)

	barycentric := cartesianToBarycentric(p.Add(samplePointOffset), a, b, c)

	if barycentric.X > 0 && barycentric.Y > 0 && barycentric.Z > 0 {
		return true, barycentric
	}

	return false, lm.Vector3{}
}

func sampleTexture(tex *image.NRGBA, texcoord lm.Vector2) lm.Vector4 {
	x := int(texcoord.X * float64(tex.Rect.Dx()))
	y := int(texcoord.Y * float64(tex.Rect.Dy()))
	c := tex.NRGBAAt(x, y)

	return lm.Vector4{
		X: float64(c.R) / 255,
		Y: float64(c.G) / 255,
		Z: float64(c.B) / 255,
		W: float64(c.A) / 255,
	}
}

var SunLightDir = lm.Vec3(-0.5, 1, -0.8).Normalize()
var SunLightIntensity = 0.95 / SunLightDir.MaxComponent()

func shadePixel(lighting float64, texture *image.NRGBA, normal lm.Vector3, texcoord lm.Vector2) color.NRGBA {
	light := SunLightIntensity * lighting * lm.Clamp(math.Abs(normal.Dot(SunLightDir))*0.8+0.2, 0.0, 1.0)

	if texture != nil {
		rgba := sampleTexture(texture, texcoord)
		col := rgba.XYZ().PowScalar(Gamma).MulScalar(lighting).PowScalar(1.0/Gamma).ClampScalar(0.0, 1.0)

		return color.NRGBA{
			R: uint8(255 * col.X),
			G: uint8(255 * col.Y),
			B: uint8(255 * col.Z),
			A: uint8(255 * rgba.W),
		}
	} else {
		return color.NRGBA{
			R: uint8(255 * light),
			G: uint8(255 * light),
			B: uint8(255 * light),
			A: 255,
		}
	}
}

func (r *NodeRasterizer) drawTriangle(target *RenderBuffer, tex *image.NRGBA, lighting float64, a, b, c mesh.Vertex) {
	origin := lm.Vector2{
		X: float64(target.Color.Bounds().Dx()) / 2,
		Y: float64(target.Color.Bounds().Dy()) / 2,
	}

	a.Position = r.projection.MulVec(a.Position)
	b.Position = r.projection.MulVec(b.Position)
	c.Position = r.projection.MulVec(c.Position)

	screenSpaceA := a.Position.XY().Mul(lm.Vec2(1, -1)).MulScalar(BaseResolution * math.Sqrt2 / 2).Add(origin)
	screenSpaceB := b.Position.XY().Mul(lm.Vec2(1, -1)).MulScalar(BaseResolution * math.Sqrt2 / 2).Add(origin)
	screenSpaceC := c.Position.XY().Mul(lm.Vec2(1, -1)).MulScalar(BaseResolution * math.Sqrt2 / 2).Add(origin)

	bboxMin := screenSpaceA.Min(screenSpaceB).Min(screenSpaceC)
	bboxMax := screenSpaceA.Max(screenSpaceB).Max(screenSpaceC)

	for y := int(bboxMin.Y); y < int(bboxMax.Y)+1; y++ {
		for x := int(bboxMin.X); x < int(bboxMax.X)+1; x++ {
			pointIsInsideTriangle, barycentric := sampleTriangle(x, y, screenSpaceA, screenSpaceB, screenSpaceC)

			if !pointIsInsideTriangle {
				continue
			}

			pixelDepth := lm.Vec3(a.Position.Z, b.Position.Z, c.Position.Z).Dot(barycentric)

			normal := a.Normal.MulScalar(barycentric.X).
				Add(b.Normal.MulScalar(barycentric.Y)).
				Add(c.Normal.MulScalar(barycentric.Z))

			texcoord := a.Texcoord.MulScalar(barycentric.X).
				Add(b.Texcoord.MulScalar(barycentric.Y)).
				Add(c.Texcoord.MulScalar(barycentric.Z))

			finalColor := shadePixel(lighting, tex, normal, texcoord)

			if finalColor.A > 10 { // FIXME
				if pixelDepth > target.Depth.At(x, y) {
					continue
				}

				target.Color.SetNRGBA(x, y, finalColor)
				target.Depth.Set(x, y, pixelDepth)
			}
		}
	}
}

func transformToFaceDir(v lm.Vector3, facedir uint8) lm.Vector3 {
	axis := (facedir >> 2) & 0x7
	dir := facedir & 0x3

	// Left click with screwdriver
	switch dir {
	case 0: // no-op
	case 1:
		v = v.RotateXZ(lm.Radians(-90))
	case 2:
		v = v.RotateXZ(lm.Radians(180))
	case 3:
		v = v.RotateXZ(lm.Radians(90))
	}

	// Right click with screwdriver
	switch axis {
	case 0: // no-op
	case 1:
		v = v.RotateYZ(lm.Radians(90))
	case 2:
		v = v.RotateYZ(lm.Radians(-90))
	case 3:
		v = v.RotateXY(lm.Radians(-90))
	case 4:
		v = v.RotateXY(lm.Radians(90))
	case 5:
		v = v.RotateXY(lm.Radians(180))
	}

	return v
}

func (r *NodeRasterizer) createMesh(node RenderableNode, nodeDef *game.NodeDefinition) *mesh.Model {
	switch {
	case nodeDef.DrawType.IsLiquid():
		return mesh.Cube(node.HiddenFaces)
	default:
		return nodeDef.Model
	}
}

func (r *NodeRasterizer) Render(node RenderableNode, nodeDef *game.NodeDefinition) *RenderBuffer {
	if nodeDef.DrawType == game.DrawTypeAirlike || nodeDef.Model == nil || len(nodeDef.Textures) == 0 {
		return nil
	}

	if target, ok := r.cache[node]; ok {
		return target
	}

	rect := image.Rect(0, 0, BaseResolution, BaseResolution+BaseResolution/8)
	target := NewRenderBuffer(rect)

	model := r.createMesh(node, nodeDef)

	for j, mesh := range model.Meshes {
		triangleCount := len(mesh.Vertices) / 3

		for i := 0; i < triangleCount; i++ {
			vertexA := mesh.Vertices[i*3]
			vertexB := mesh.Vertices[i*3+1]
			vertexC := mesh.Vertices[i*3+2]

			if nodeDef.ParamType2 == game.ParamType2FaceDir {
				vertexA.Position = transformToFaceDir(vertexA.Position, node.Param2)
				vertexB.Position = transformToFaceDir(vertexB.Position, node.Param2)
				vertexC.Position = transformToFaceDir(vertexC.Position, node.Param2)
				vertexA.Normal = transformToFaceDir(vertexA.Normal, node.Param2)
				vertexB.Normal = transformToFaceDir(vertexB.Normal, node.Param2)
				vertexC.Normal = transformToFaceDir(vertexC.Normal, node.Param2)
			}

			vertexA.Position.Z = -vertexA.Position.Z
			vertexB.Position.Z = -vertexB.Position.Z
			vertexC.Position.Z = -vertexC.Position.Z

			vertexA.Position.X = -vertexA.Position.X
			vertexB.Position.X = -vertexB.Position.X
			vertexC.Position.X = -vertexC.Position.X

			r.drawTriangle(target, nodeDef.Textures[j], node.Light, vertexA, vertexB, vertexC)
		}
	}

	r.cache[node] = target

	return target
}
