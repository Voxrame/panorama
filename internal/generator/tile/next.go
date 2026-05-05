package tile

import (
	"github.com/Voxrame/panorama/internal/game"
	"github.com/Voxrame/panorama/pkg/geom"
	"github.com/Voxrame/panorama/pkg/world"
)

type NextTiler struct {
}

func (t *NextTiler) FullRender(game *game.Game, world *world.World, workers int, region geom.Region, createRenderer CreateRendererFunc) {

}
