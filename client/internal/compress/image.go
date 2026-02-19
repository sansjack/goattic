package compress

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"strings"

	_ "golang.org/x/image/webp"
)

func Image(in, out, ext string) (bool, error) {
	f, err := os.Open(in)
	if err != nil {
		return false, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return false, fmt.Errorf("decode image: %w", err)
	}

	tmp, err := os.CreateTemp("", "goattic-img-*")
	if err != nil {
		return false, err
	}
	defer tmp.Close()
	tmpName := tmp.Name()

	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg":
		if err := jpeg.Encode(tmp, img, &jpeg.Options{Quality: 80}); err != nil {
			os.Remove(tmpName)
			return false, fmt.Errorf("encode jpeg: %w", err)
		}
	case ".png":
		enc := png.Encoder{CompressionLevel: png.BestCompression}
		if err := enc.Encode(tmp, img); err != nil {
			os.Remove(tmpName)
			return false, fmt.Errorf("encode png: %w", err)
		}
	default:
		os.Remove(tmpName)
		return false, nil
	}
	tmp.Close()

	origInfo, _ := os.Stat(in)
	compInfo, _ := os.Stat(tmpName)
	if compInfo.Size() >= origInfo.Size() {
		os.Remove(tmpName)
		return false, nil
	}

	if err := os.Rename(tmpName, out); err != nil {
		os.Remove(tmpName)
		return false, err
	}

	return true, nil
}
