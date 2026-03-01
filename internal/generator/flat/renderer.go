package flat

import (
	"image"
	"log/slog"

	"github.com/lord-server/panorama/internal/game"
	"github.com/lord-server/panorama/internal/generator/rasterizer"
	"github.com/lord-server/panorama/internal/generator/tile"
	"github.com/lord-server/panorama/pkg/geom"
	"github.com/lord-server/panorama/pkg/lm"
	"github.com/lord-server/panorama/pkg/world"
	"github.com/lord-server/panorama/pkg/world/selector"
)

type FlatRenderer struct {
	nr rasterizer.NodeRasterizer

	region geom.Region
	game   *game.Game
}

func NewRenderer(region geom.Region, game *game.Game) *FlatRenderer {
	return &FlatRenderer{
		nr:     rasterizer.New(lm.DimetricProjection()),
		region: region,
		game:   game,
	}
}

func (r *FlatRenderer) RenderTile(
	tilePos tile.TilePosition,
	wd *world.World,
	game *game.Game,
) *rasterizer.RenderBuffer {
	rect := image.Rect(0, 0, 256, 256)
	target := rasterizer.NewRenderBuffer(rect)

	selector := selector.BlocksAlongY{
		X: tilePos.X,
		Z: tilePos.Y,
	}

	err := wd.GetBlocks(selector, func(pos geom.BlockPosition, block *world.MapBlock) error {
		return nil
	})
	if err != nil {
		slog.Error("unable to get blocks", "error", err)
	}

	return target
}

func (r *FlatRenderer) ProjectRegion(region geom.Region) geom.ProjectedRegion {
	return geom.ProjectedRegion{
		XBounds: geom.Bounds{
			Min: region.XBounds.Min / 16,
			Max: region.XBounds.Max / 16,
		},
		YBounds: geom.Bounds{
			Min: region.ZBounds.Min / 16,
			Max: region.ZBounds.Max / 16,
		},
	}
}
