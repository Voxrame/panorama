package render

import (
	"github.com/lord-server/panorama/internal/game"
	"github.com/lord-server/panorama/internal/spatial"
	"github.com/lord-server/panorama/internal/world"
)

type NodeEntry struct {
	Name string
	Def  game.NodeDefinition
}

type resolvedBlock struct {
	block *world.MapBlock
	table []NodeEntry
}

type BlockNeighborhood struct {
	blocks      [27]*resolvedBlock
	ignoreEntry NodeEntry
}

var neighborhoodCenter = spatial.BlockPosition{X: 1, Y: 1, Z: 1}

func NewBlockNeighborhood(g *game.Game) BlockNeighborhood {
	return BlockNeighborhood{
		ignoreEntry: NodeEntry{
			Name: "ignore",
			Def:  g.NodeDef("ignore"),
		},
	}
}

func buildResolvedTable(block *world.MapBlock, g *game.Game) []NodeEntry {
	maxID := uint16(0)
	for id := range block.MappingIter() {
		if id > maxID {
			maxID = id
		}
	}
	table := make([]NodeEntry, maxID+1)
	for id, name := range block.MappingIter() {
		table[id] = NodeEntry{Name: name, Def: g.NodeDef(name)}
	}
	return table
}

func blockIndex(pos spatial.BlockPosition) int {
	return pos.Z*9 + pos.Y*3 + pos.X
}

func (b *BlockNeighborhood) FetchBlock(w *world.World, g *game.Game, posOffset, worldPos spatial.BlockPosition) {
	block, err := w.GetBlock(worldPos.Add(posOffset))

	if err != nil {
		return
	}

	b.SetBlock(neighborhoodCenter.Add(posOffset), block, g)
}

func (b *BlockNeighborhood) SetBlock(pos spatial.BlockPosition, block *world.MapBlock, g *game.Game) {
	if block == nil {
		b.blocks[blockIndex(pos)] = nil
		return
	}

	b.blocks[blockIndex(pos)] = &resolvedBlock{
		block: block,
		table: buildResolvedTable(block, g),
	}
}

func (b *BlockNeighborhood) getBlockByNodePos(pos spatial.NodePosition) *resolvedBlock {
	blockPos := spatial.BlockPosition{
		X: pos.X/spatial.BlockSize + neighborhoodCenter.X,
		Y: pos.Y/spatial.BlockSize + neighborhoodCenter.Y,
		Z: pos.Z/spatial.BlockSize + neighborhoodCenter.Z,
	}

	return b.blocks[blockIndex(blockPos)]
}

func (b *BlockNeighborhood) GetNode(pos spatial.NodePosition) (NodeEntry, uint8, uint8) {
	rb := b.getBlockByNodePos(pos)

	if rb == nil {
		return b.ignoreEntry, 0, 0
	}

	node := rb.block.GetNode(spatial.NodePosition{
		X: pos.X % spatial.BlockSize,
		Y: pos.Y % spatial.BlockSize,
		Z: pos.Z % spatial.BlockSize,
	})

	if int(node.ID) < len(rb.table) {
		return rb.table[node.ID], node.Param1, node.Param2
	}

	return b.ignoreEntry, node.Param1, node.Param2
}

func (b *BlockNeighborhood) GetParam1(pos spatial.NodePosition) uint8 {
	rb := b.getBlockByNodePos(pos)

	if rb == nil {
		return 0
	}

	node := rb.block.GetNode(spatial.NodePosition{
		X: pos.X % spatial.BlockSize,
		Y: pos.Y % spatial.BlockSize,
		Z: pos.Z % spatial.BlockSize,
	})

	return node.Param1
}
