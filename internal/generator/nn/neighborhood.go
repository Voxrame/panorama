package nn

import (
	"github.com/lord-server/panorama/pkg/geom"
	"github.com/lord-server/panorama/pkg/world"
)

type BlockNeighborhood struct {
	blocks [27]*world.MapBlock
}

var neighborhoodCenter = geom.BlockPosition{X: 1, Y: 1, Z: 1}

func blockIndex(pos geom.BlockPosition) int {
	return pos.Z*9 + pos.Y*3 + pos.X
}

func (b *BlockNeighborhood) FetchBlock(w *world.World, posOffset, worldPos geom.BlockPosition) {
	block, err := w.GetBlock(worldPos.Add(posOffset))

	if err != nil {
		return
	}

	b.SetBlock(neighborhoodCenter.Add(posOffset), block)
}

func (b *BlockNeighborhood) SetBlock(pos geom.BlockPosition, block *world.MapBlock) {
	b.blocks[blockIndex(pos)] = block
}

func (b *BlockNeighborhood) getBlockByNodePos(pos geom.NodePosition) *world.MapBlock {
	blockPos := geom.BlockPosition{
		X: pos.X/geom.BlockSize + neighborhoodCenter.X,
		Y: pos.Y/geom.BlockSize + neighborhoodCenter.Y,
		Z: pos.Z/geom.BlockSize + neighborhoodCenter.Z,
	}

	return b.blocks[blockIndex(blockPos)]
}

func (b *BlockNeighborhood) GetNode(pos geom.NodePosition) (string, uint8, uint8) {
	block := b.getBlockByNodePos(pos)

	if block == nil {
		return "ignore", 0, 0
	}

	node := block.GetNode(geom.NodePosition{
		X: pos.X % geom.BlockSize,
		Y: pos.Y % geom.BlockSize,
		Z: pos.Z % geom.BlockSize,
	})

	name := block.ResolveName(node.ID)

	return name, node.Param1, node.Param2
}

func (b *BlockNeighborhood) GetParam1(pos geom.NodePosition) uint8 {
	block := b.getBlockByNodePos(pos)

	if block == nil {
		return 0
	}

	node := block.GetNode(geom.NodePosition{
		X: pos.X % geom.BlockSize,
		Y: pos.Y % geom.BlockSize,
		Z: pos.Z % geom.BlockSize,
	})

	return node.Param1
}
