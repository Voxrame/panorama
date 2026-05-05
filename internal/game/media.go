package game

import (
	"image"
	"image/color"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/Voxrame/panorama/pkg/imageutil"
	"github.com/Voxrame/panorama/pkg/mesh"
	"go.uber.org/zap"
)

type MediaCache struct {
	images     map[string]*image.NRGBA
	models     map[string]*mesh.Model
	dummyImage *image.NRGBA
}

func NewMediaCache() *MediaCache {
	dummyImage := image.NewNRGBA(image.Rect(0, 0, 2, 2))

	dummyImage.SetNRGBA(0, 0, color.NRGBA{255, 0, 255, 255})
	dummyImage.SetNRGBA(0, 1, color.NRGBA{0, 0, 0, 255})
	dummyImage.SetNRGBA(1, 0, color.NRGBA{0, 0, 0, 255})
	dummyImage.SetNRGBA(1, 1, color.NRGBA{255, 0, 255, 255})

	return &MediaCache{
		images:     make(map[string]*image.NRGBA),
		models:     make(map[string]*mesh.Model),
		dummyImage: dummyImage,
	}
}

func (m *MediaCache) fetchMedia(path string) error {
	return filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// FIXME: pass logger explicitly
			zap.S().Warnw("Encountered error while fetching media", "error", err, "dir_entry", d.Name())

			return nil
		}

		if !d.Type().IsRegular() {
			return nil
		}

		basePath := filepath.Base(path)

		switch filepath.Ext(path) {
		case ".png":
			img, _ := imageutil.LoadPNG(path)
			m.images[basePath] = img

		case ".obj":
			model, err := mesh.LoadOBJ(path)
			if err != nil {
				return err
			}

			m.models[basePath] = &model
		}

		return nil
	})
}

func (m *MediaCache) Image(name string) *image.NRGBA {
	// FIXME: resolve modifiers
	baseName := strings.Split(name, "^")[0]

	if img, ok := m.images[baseName]; ok {
		return img
	} else {
		return m.dummyImage
	}
}

func (m *MediaCache) Mesh(name string) *mesh.Model {
	if model, ok := m.models[name]; ok {
		return model
	} else {
		return nil
	}
}
