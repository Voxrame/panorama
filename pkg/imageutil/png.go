package imageutil

import (
	"image"
	"image/draw"
	"image/png"
	"log"
	"os"
	"path/filepath"
)

func toNRGBA(img image.Image) *image.NRGBA {
	dst := image.NewNRGBA(img.Bounds())
	draw.Draw(dst, img.Bounds(), img, img.Bounds().Min, draw.Src)

	return dst
}

func LoadPNG(path string) (*image.NRGBA, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("error closing file: %v", err)
		}
	}()

	img, err := png.Decode(file)
	if err != nil {
		return nil, err
	}

	return toNRGBA(img), nil
}

func SavePNG(img *image.NRGBA, name string) error {
	err := os.MkdirAll(filepath.Dir(name), os.ModePerm)
	if err != nil {
		return err
	}

	file, err := os.Create(name)
	if err != nil {
		return err
	}

	encoder := png.Encoder{
		CompressionLevel: png.BestCompression,
	}
	if err := encoder.Encode(file, img); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		log.Printf("error closing file: %v", err)
		return err
	}

	return nil
}
