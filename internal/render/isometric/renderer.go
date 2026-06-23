package isometric

import (
	"image"
	"math"

	"github.com/lord-server/panorama/internal/game"
	"github.com/lord-server/panorama/internal/lm"
	"github.com/lord-server/panorama/internal/mesh"
	"github.com/lord-server/panorama/internal/raster"
	"github.com/lord-server/panorama/internal/render"
	"github.com/lord-server/panorama/internal/spatial"
	"github.com/lord-server/panorama/internal/world"
)

var (
	YOffsetCoef     = int(math.Round(render.BaseResolution * (1 + math.Sqrt2) / 4))
	TileBlockWidth  = spatial.BlockSize * render.BaseResolution
	TileBlockHeight = render.BaseResolution/2*spatial.BlockSize - 1 + YOffsetCoef*spatial.BlockSize
)

type IsometricRenderer struct {
	nr render.NodeRasterizer

	region spatial.Region
	game   *game.Game
}

func NewRenderer(region spatial.Region, game *game.Game) *IsometricRenderer {
	return &IsometricRenderer{
		nr:     render.NewNodeRasterizer(lm.DimetricProjection()),
		region: region,
		game:   game,
	}
}

func (r *IsometricRenderer) renderNode(
	target *raster.RenderBuffer,
	pos spatial.NodePosition,
	worldPos spatial.NodePosition,
	neighborhood *render.BlockNeighborhood,
	offset image.Point,
	depthOffset float64,
) {
	entry, param1, param2 := neighborhood.GetNode(pos)

	if entry.Def.DrawType == game.DrawTypeAirlike {
		return
	}

	nodeDef := entry.Def

	needsAlphaBlending := true
	if nodeDef.DrawType == game.DrawTypeNormal {
		needsAlphaBlending = false
	}

	// Estimate lighting by sampling neighboring nodes and using the brightest one
	neighborOffsets := []spatial.NodePosition{
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
		neighborEntry, param1, _ := neighborhood.GetNode(neighborPos)
		if param1 > maxParam1 {
			maxParam1 = param1
		}

		// Compute visibility for stacked liquids
		if nodeDef.DrawType.IsLiquid() {
			hiddenFaces |= mesh.CubeFaceWest | mesh.CubeFaceDown | mesh.CubeFaceSouth

			if neighborEntry.Def.DrawType.IsLiquid() {
				hiddenFaces |= neighborFaces[i]
			}
		}
	}

	// Make underground edges visible (otherwise the edge becomes oddly thin and
	// that doesn't look good)
	if r.region.IsAtEdge(worldPos) && maxParam1 == render.ZeroIntensity {
		maxParam1 = render.MapEdgeIntensity
	}

	renderableNode := render.RenderableNode{
		Name:        entry.Name,
		Light:       render.DecodeLight(maxParam1),
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

func (r *IsometricRenderer) renderBlock(
	target *raster.RenderBuffer,
	blockPos spatial.BlockPosition,
	neighborhood *render.BlockNeighborhood,
	offset image.Point,
	depthOffset float64,
) {
	rect := image.Rect(0, 0, TileBlockWidth, TileBlockHeight)

	// FIXME: nodes must define their origin points
	originX, originY := rect.Dx()/2-render.BaseResolution/2, rect.Dy()/2+render.BaseResolution/4+2

	for z := spatial.BlockSize - 1; z >= 0; z-- {
		for y := spatial.BlockSize - 1; y >= 0; y-- {
			for x := spatial.BlockSize - 1; x >= 0; x-- {
				nodePos := spatial.NodePosition{X: x, Y: y, Z: z}
				nodeWorldPos := blockPos.AddNode(nodePos)

				if !r.region.Intersects(nodeWorldPos.Region()) {
					continue
				}

				offset := image.Point{
					X: originX + render.BaseResolution*(z-x)/2 + offset.X,
					Y: originY + render.BaseResolution*(z+x)/4 + offset.Y - YOffsetCoef*y,
				}

				r.renderNode(target, nodePos, nodeWorldPos, neighborhood, offset, depthOffset)
			}
		}
	}
}

func (r *IsometricRenderer) RenderTile(
	tilePos render.TilePosition,
	world *world.World,
	game *game.Game,
) *raster.RenderBuffer {
	tilePos.Y *= 2

	rect := image.Rect(0, 0, TileBlockWidth, TileBlockWidth)
	target := raster.NewRenderBuffer(rect)

	centerX := tilePos.Y - tilePos.X
	centerY := 0
	centerZ := tilePos.Y + tilePos.X

	yMin := int(math.Floor(float64(r.region.YBounds.Min) / float64(spatial.BlockSize)))
	yMax := int(math.Ceil(float64(r.region.YBounds.Max) / float64(spatial.BlockSize)))

	for i := yMin; i < yMax; i++ {
		for z := -3; z <= 3; z++ {
			for x := -3; x <= 3; x++ {
				blockPos := spatial.BlockPosition{
					X: centerX + x + i,
					Y: centerY + i,
					Z: centerZ + z + i,
				}

			neighborhood := render.NewBlockNeighborhood(game)

			neighborhood.FetchBlock(world, game, spatial.BlockPosition{X: 0, Y: 0, Z: 0}, blockPos)
			neighborhood.FetchBlock(world, game, spatial.BlockPosition{X: 1, Y: 0, Z: 0}, blockPos)
			neighborhood.FetchBlock(world, game, spatial.BlockPosition{X: 0, Y: 1, Z: 0}, blockPos)
			neighborhood.FetchBlock(world, game, spatial.BlockPosition{X: 0, Y: 0, Z: 1}, blockPos)

				offset := image.Point{
					X: render.BaseResolution * (z - x) / 2 * spatial.BlockSize,
					Y: (render.BaseResolution*(z+x+2*i)/4 - i*YOffsetCoef) * spatial.BlockSize,
				}

				depthOffset := (-float64(z+x+2*i)/math.Sqrt2 - 0.5*float64(i)) * spatial.BlockSize
				r.renderBlock(target, blockPos, &neighborhood, offset, depthOffset)
			}
		}
	}

	return target
}

func (r *IsometricRenderer) ProjectRegion(region spatial.Region) spatial.ProjectedRegion {
	xMin := int(math.Floor(float64((region.ZBounds.Min - region.XBounds.Max)) / 2 / spatial.BlockSize))
	xMax := int(math.Ceil(float64((region.ZBounds.Max - region.XBounds.Min)) / 2 / spatial.BlockSize))

	yMin := int(math.Floor((float64(region.ZBounds.Min+region.XBounds.Min+2*region.YBounds.Max)/4 -
		float64(region.YBounds.Max*YOffsetCoef)/render.BaseResolution) / spatial.BlockSize))
	yMax := int(math.Ceil((float64(region.ZBounds.Max+region.XBounds.Max+2*region.YBounds.Min)/4 -
		float64(region.YBounds.Min*YOffsetCoef)/render.BaseResolution) / spatial.BlockSize))

	return spatial.ProjectedRegion{
		XBounds: spatial.Bounds{
			Min: xMin,
			Max: xMax,
		},
		YBounds: spatial.Bounds{
			Min: yMin,
			Max: yMax,
		},
	}
}
