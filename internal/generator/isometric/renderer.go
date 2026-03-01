package isometric

import (
	"image"
	"math"

	"github.com/lord-server/panorama/internal/game"
	"github.com/lord-server/panorama/internal/generator/light"
	"github.com/lord-server/panorama/internal/generator/nn"
	"github.com/lord-server/panorama/internal/generator/rasterizer"
	"github.com/lord-server/panorama/internal/generator/tile"
	"github.com/lord-server/panorama/internal/world"
	"github.com/lord-server/panorama/pkg/geom"
	"github.com/lord-server/panorama/pkg/lm"
	"github.com/lord-server/panorama/pkg/mesh"
)

var (
	YOffsetCoef     = int(math.Round(rasterizer.BaseResolution * (1 + math.Sqrt2) / 4))
	TileBlockWidth  = geom.BlockSize * rasterizer.BaseResolution
	TileBlockHeight = rasterizer.BaseResolution/2*geom.BlockSize - 1 + YOffsetCoef*geom.BlockSize
)

type IsometricRenderer struct {
	nr rasterizer.NodeRasterizer

	region geom.Region
	game   *game.Game
}

func NewRenderer(region geom.Region, game *game.Game) *IsometricRenderer {
	return &IsometricRenderer{
		nr:     rasterizer.New(lm.DimetricProjection()),
		region: region,
		game:   game,
	}
}

func (r *IsometricRenderer) renderNode(
	target *rasterizer.RenderBuffer,
	pos geom.NodePosition,
	worldPos geom.NodePosition,
	neighborhood *nn.BlockNeighborhood,
	offset image.Point,
	depthOffset float64,
) {
	name, param1, param2 := neighborhood.GetNode(pos)

	// Fast path: checking for air immediately is faster than fetching NodeDefinition
	if name == "air" {
		return
	}

	nodeDef := r.game.NodeDef(name)

	needsAlphaBlending := nodeDef.DrawType != game.DrawTypeNormal

	maxParam1, hiddenFaces := r.estimateVisibility(nodeDef, neighborhood, param1, pos)

	// Make underground edges visible (otherwise the edge becomes oddly thin and
	// that doesn't look good)
	if r.region.IsAtEdge(worldPos) && maxParam1 == light.ZeroIntensity {
		maxParam1 = light.MapEdgeIntensity
	}

	renderableNode := rasterizer.RenderableNode{
		Name:        name,
		Light:       light.Decode(maxParam1),
		Param2:      param2,
		HiddenFaces: hiddenFaces,
	}
	renderedNode := r.nr.Render(renderableNode, &nodeDef)

	depthOffset = -float64(pos.Z+pos.X)/math.Sqrt2 - 0.5*(float64(pos.Y)) + depthOffset
	if needsAlphaBlending {
		target.OverlayDepthAwareWithAlpha(renderedNode, offset, depthOffset)
	} else {
		target.OverlayDepthAware(renderedNode, offset, depthOffset)
	}
}

func (r *IsometricRenderer) estimateVisibility(
	nodeDef game.NodeDefinition,
	neighborhood *nn.BlockNeighborhood,
	param1 uint8,
	pos geom.NodePosition,
) (uint8, mesh.CubeFaces) {
	// Estimate lighting by sampling neighboring nodes and using the brightest one
	neighborOffsets := []geom.NodePosition{
		{X: 1, Y: 0, Z: 0},
		{X: 0, Y: 1, Z: 0},
		{X: 0, Y: 0, Z: 1},
	}

	neighborFaces := []mesh.CubeFaces{
		mesh.CubeFaceEast,
		mesh.CubeFaceTop,
		mesh.CubeFaceNorth,
	}

	maxParam1 := param1
	hiddenFaces := mesh.CubeFaces(0)

	for i, offset := range neighborOffsets {
		neighborPos := pos.Add(offset)
		neighborName, param1, _ := neighborhood.GetNode(neighborPos)

		if param1 > maxParam1 {
			maxParam1 = param1
		}

		// Compute visibility for stacked liquids
		if nodeDef.DrawType.IsLiquid() {
			hiddenFaces |= mesh.CubeFaceWest | mesh.CubeFaceDown | mesh.CubeFaceSouth

			neighborNodeDef := r.game.NodeDef(neighborName)
			if neighborNodeDef.DrawType.IsLiquid() {
				hiddenFaces |= neighborFaces[i]
			}
		}
	}

	return maxParam1, hiddenFaces
}

func (r *IsometricRenderer) renderBlock(
	target *rasterizer.RenderBuffer,
	blockPos geom.BlockPosition,
	neighborhood *nn.BlockNeighborhood,
	offset image.Point,
	depthOffset float64,
) {
	rect := image.Rect(0, 0, TileBlockWidth, TileBlockHeight)

	// FIXME: nodes must define their origin points
	originX, originY := rect.Dx()/2-rasterizer.BaseResolution/2, rect.Dy()/2+rasterizer.BaseResolution/4+2

	for z := geom.BlockSize - 1; z >= 0; z-- {
		for y := geom.BlockSize - 1; y >= 0; y-- {
			for x := geom.BlockSize - 1; x >= 0; x-- {
				nodePos := geom.NodePosition{X: x, Y: y, Z: z}
				nodeWorldPos := blockPos.AddNode(nodePos)

				if !r.region.Intersects(nodeWorldPos.Region()) {
					continue
				}

				offset := image.Point{
					X: originX + rasterizer.BaseResolution*(z-x)/2 + offset.X,
					Y: originY + rasterizer.BaseResolution*(z+x)/4 + offset.Y - YOffsetCoef*y,
				}

				r.renderNode(target, nodePos, nodeWorldPos, neighborhood, offset, depthOffset)
			}
		}
	}
}

func (r *IsometricRenderer) RenderTile(
	tilePos tile.TilePosition,
	world *world.World,
	game *game.Game,
) *rasterizer.RenderBuffer {
	tilePos.Y *= 2

	rect := image.Rect(0, 0, TileBlockWidth, TileBlockWidth)
	target := rasterizer.NewRenderBuffer(rect)

	centerX := tilePos.Y - tilePos.X
	centerY := 0
	centerZ := tilePos.Y + tilePos.X

	yMin := int(math.Floor(float64(r.region.YBounds.Min) / float64(geom.BlockSize)))
	yMax := int(math.Ceil(float64(r.region.YBounds.Max) / float64(geom.BlockSize)))

	for i := yMin; i < yMax; i++ {
		for z := -3; z <= 3; z++ {
			for x := -3; x <= 3; x++ {
				blockPos := geom.BlockPosition{
					X: centerX + x + i,
					Y: centerY + i,
					Z: centerZ + z + i,
				}

				neighborhood := nn.BlockNeighborhood{}

				neighborhood.FetchBlock(world, geom.BlockPosition{X: 0, Y: 0, Z: 0}, blockPos)
				neighborhood.FetchBlock(world, geom.BlockPosition{X: 1, Y: 0, Z: 0}, blockPos)
				neighborhood.FetchBlock(world, geom.BlockPosition{X: 0, Y: 1, Z: 0}, blockPos)
				neighborhood.FetchBlock(world, geom.BlockPosition{X: 0, Y: 0, Z: 1}, blockPos)

				offset := image.Point{
					X: rasterizer.BaseResolution * (z - x) / 2 * geom.BlockSize,
					Y: (rasterizer.BaseResolution*(z+x+2*i)/4 - i*YOffsetCoef) * geom.BlockSize,
				}

				depthOffset := (-float64(z+x+2*i)/math.Sqrt2 - 0.5*float64(i)) * geom.BlockSize
				r.renderBlock(target, blockPos, &neighborhood, offset, depthOffset)
			}
		}
	}

	return target
}

func (r *IsometricRenderer) ProjectRegion(region geom.Region) geom.ProjectedRegion {
	xMin := int(math.Floor(float64((region.ZBounds.Min - region.XBounds.Max)) / 2 / geom.BlockSize))
	xMax := int(math.Ceil(float64((region.ZBounds.Max - region.XBounds.Min)) / 2 / geom.BlockSize))

	yMin := int(math.Floor((float64(region.ZBounds.Min+region.XBounds.Min+2*region.YBounds.Max)/4 -
		float64(region.YBounds.Max*YOffsetCoef)/rasterizer.BaseResolution) / geom.BlockSize))
	yMax := int(math.Ceil((float64(region.ZBounds.Max+region.XBounds.Max+2*region.YBounds.Min)/4 -
		float64(region.YBounds.Min*YOffsetCoef)/rasterizer.BaseResolution) / geom.BlockSize))

	return geom.ProjectedRegion{
		XBounds: geom.Bounds{
			Min: xMin,
			Max: xMax,
		},
		YBounds: geom.Bounds{
			Min: yMin,
			Max: yMax,
		},
	}
}
