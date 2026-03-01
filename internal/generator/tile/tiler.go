package tile

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/lord-server/panorama/internal/game"
	"github.com/lord-server/panorama/internal/generator/rasterizer"
	"github.com/lord-server/panorama/pkg/geom"
	"github.com/lord-server/panorama/pkg/imageutil"
	"github.com/lord-server/panorama/pkg/lm"
	"github.com/lord-server/panorama/pkg/world"
)

type TilePosition struct {
	X, Y int
}

type Renderer interface {
	RenderTile(pos TilePosition, w *world.World, game *game.Game) *rasterizer.RenderBuffer
	ProjectRegion(region geom.Region) geom.ProjectedRegion
}

type Tiler struct {
	region     geom.Region
	zoomLevels int
	tilesPath  string
}

func NewTiler(region geom.Region, zoomLevels int, tilesPath string) Tiler {
	return Tiler{
		region:     region,
		zoomLevels: zoomLevels,
		tilesPath:  tilesPath,
	}
}

func (t *Tiler) tilePath(x, y, zoom int) string {
	return fmt.Sprintf("%v/%v/%v/%v.png", t.tilesPath, -zoom, x, y)
}

func (t *Tiler) worker(wg *sync.WaitGroup, game *game.Game, world *world.World, renderer Renderer, positions <-chan TilePosition) {
	for position := range positions {
		output := renderer.RenderTile(position, world, game)
		// Don't save empty tiles
		if !output.Dirty {
			continue
		}

		tilePath := t.tilePath(position.X, position.Y, 0)

		err := imageutil.SavePNG(output.Color, tilePath)
		if err != nil {
			return
		}

		slog.Info("saved", "path", tilePath)
	}

	wg.Done()
}

type CreateRendererFunc func() Renderer

func (t *Tiler) FullRender(game *game.Game, world *world.World, workers int, region geom.Region, createRenderer CreateRendererFunc) {
	var wg sync.WaitGroup

	positions := make(chan TilePosition)
	projectedRegion := geom.ProjectedRegion{}

	for i := 0; i < workers; i++ {
		wg.Add(1)

		renderer := createRenderer()
		projectedRegion = renderer.ProjectRegion(region)

		go t.worker(&wg, game, world, renderer, positions)
	}

	for x := projectedRegion.XBounds.Min; x < projectedRegion.XBounds.Max; x++ {
		err := os.MkdirAll(fmt.Sprintf("%v/%v", path.Join(t.tilesPath, "0"), x), os.ModePerm)
		if err != nil {
			panic(err)
		}

		for y := projectedRegion.YBounds.Min; y < projectedRegion.YBounds.Max; y++ {
			positions <- TilePosition{X: x, Y: y}
		}
	}

	close(positions)

	wg.Wait()
}

// DownscaleTiles rescales high-resolution tiles into lower resolution ones until it reaches adequate zoom level
func (t *Tiler) DownscaleTiles() {
	slog.Info("downscaling", "zoomLevels", t.zoomLevels)

	tileDir, err := filepath.Abs(path.Join(t.tilesPath, "0"))
	if err != nil {
		panic(err)
	}

	// Collect tile positions
	var positions []TilePosition

	err = filepath.WalkDir(tileDir, func(path string, d fs.DirEntry, _ error) error {
		if d == nil || d.IsDir() {
			return nil
		}

		dir, file := filepath.Split(path)

		y, err := strconv.Atoi(strings.TrimSuffix(file, filepath.Ext(file)))
		if err != nil {
			slog.Warn("skipped file due to error", "path", path, "err", err)

			return nil
		}

		x, err := strconv.Atoi(filepath.Base(dir))
		if err != nil {
			slog.Warn("skipped file due to error", "path", path, "err", err)

			return nil
		}

		positions = append(positions, TilePosition{
			X: lm.FloorDiv(x, 2),
			Y: lm.FloorDiv(y, 2),
		})

		return nil
	})

	if err != nil {
		panic(err)
	}

	positions = uniquePositions(positions)

	for zoom := 1; zoom <= t.zoomLevels; zoom++ {
		slog.Info("rescaling tiles", "zoom", zoom)
		positions = t.downscalePositions(zoom, positions)
	}
}
