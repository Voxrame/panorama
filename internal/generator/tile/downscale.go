package tile

import (
	"image"
	"image/draw"
	"log/slog"
	"sort"

	"github.com/nfnt/resize"

	"github.com/Voxrame/panorama/pkg/imageutil"
	"github.com/Voxrame/panorama/pkg/lm"
)

func uniquePositions(input []TilePosition) []TilePosition {
	// Slices with zero or one element always contain unique elements
	if len(input) < 2 {
		return input
	}

	// Sort positions by their coordinates
	sort.Slice(input, func(i, j int) bool {
		if input[i].X < input[j].X {
			return true
		}

		if input[i].X > input[j].X {
			return false
		}

		if input[i].Y < input[j].Y {
			return true
		}

		if input[i].Y > input[j].Y {
			return false
		}

		return false
	})

	// Loop over the slice and skip repeating elements
	j := 1

	for i := 1; i < len(input); i++ {
		// Skip element if it repeats
		if input[i] == input[i-1] {
			continue
		}

		// Rewrite repeated elements with unique ones
		input[j] = input[i]
		j++
	}

	return input[:j]
}

// downscalePositions produces downscaled images for given zoom level and returns a list of produced tile positions
func (t *Tiler) downscalePositions(zoom int, positions []TilePosition) []TilePosition {
	const quadrantSize = 128

	var nextPositions []TilePosition

	for _, pos := range positions {
		target := image.NewNRGBA(image.Rect(0, 0, 256, 256))

		for quadrantY := 0; quadrantY < 2; quadrantY++ {
			for quadrantX := 0; quadrantX < 2; quadrantX++ {
				source, err := imageutil.LoadPNG(t.tilePath(pos.X*2+quadrantX, pos.Y*2+quadrantY, zoom-1))
				if err != nil {
					continue
				}

				quadrant := resize.Resize(quadrantSize, quadrantSize, source, resize.Lanczos3)

				targetX := quadrantX * quadrantSize
				targetY := quadrantY * quadrantSize
				draw.Draw(target, image.Rect(targetX, targetY, targetX+quadrantSize, targetY+quadrantSize), quadrant, image.Pt(0, 0), draw.Src)
			}
		}

		imagePath := t.tilePath(pos.X, pos.Y, zoom)

		err := imageutil.SavePNG(target, imagePath)
		if err != nil {
			slog.Error("unable to save image", "err", err, "path", imagePath)
		}

		nextPositions = append(nextPositions, TilePosition{
			X: lm.FloorDiv(pos.X, 2),
			Y: lm.FloorDiv(pos.Y, 2),
		})
	}

	nextPositions = uniquePositions(nextPositions)

	return nextPositions
}
