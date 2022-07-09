package main

import (
	"errors"
	"fmt"
	"gopkg.in/gographics/imagick.v2/imagick"
	"os"
	"time"
)

const (
	ImgMaxSide       = 1280
	ThumbnailMaxSide = 300
)

func ProcessImages(blob []byte) (string, string, error) {
	imagick.Initialize()
	defer imagick.Terminate()

	mw := imagick.NewMagickWand()
	if err := mw.ReadImageBlob(blob); err != nil {
		return "", "", err
	}

	var (
		imgBlob       []byte
		thumbnailBlob []byte
	)

	format := mw.GetImageFormat()
	if format == "PNG" || format == "JPEG" || format == "BMP" || format == "WEBP" {
		resWidth, resHeight := getImageRes(mw, ImgMaxSide)

		// Processing main image.
		if err := resizeAndCompress(mw, resWidth, resHeight); err != nil {
			return "", "", err
		}

		// Copying blob.
		tmp := mw.GetImageBlob()
		imgBlob = make([]byte, len(tmp))
		copy(imgBlob, tmp)

		// Processing thumbnail image.
		tResWidth, tResHeight := getImageRes(mw, ThumbnailMaxSide)
		if err := resizeAndCompress(mw, tResWidth, tResHeight); err != nil {
			return "", "", err
		}
		if err := cropImage(mw); err != nil {
			return "", "", err
		}

		// No copy is needed.
		thumbnailBlob = mw.GetImageBlob()
	} else if format == "GIF" {
		imgBlob = blob

		// Processing thumbnail image.
		tResWidth, tResHeight := getImageRes(mw, ImgMaxSide)
		if err := resizeAndCompress(mw, tResWidth, tResHeight); err != nil {
			return "", "", err
		}
		if err := cropImage(mw); err != nil {
			return "", "", err
		}

		// No copy is needed.
		thumbnailBlob = mw.GetImageBlob()
	} else {
		return "", "", errors.New("unsupported image format")
	}

	os.WriteFile("image2.jpg", imgBlob, 0777)
	os.WriteFile("image3.jpg", thumbnailBlob, 0777)

	return "", "", nil
}

// getThumbnailRes returns resized resolution for thumbnail (given width and height should not be bigger than maxSide).
func getThumbnailRes(mw *imagick.MagickWand, maxSide uint) (uint, uint) {
	width := mw.GetImageWidth()
	height := mw.GetImageHeight()

	if width < height {
		ratio := float32(height) / float32(width)
		return maxSide, uint(float32(maxSide) * ratio)
	} else {
		ratio := float32(width) / float32(height)
		return uint(float32(maxSide) * ratio), maxSide
	}
}

// getImageRes returns resized resolution for main image.
func getImageRes(mw *imagick.MagickWand, maxSide uint) (uint, uint) {
	width := mw.GetImageWidth()
	height := mw.GetImageHeight()

	if width > height && width > maxSide {
		ratio := float32(height) / float32(width)
		return maxSide, uint(float32(maxSide) * ratio)
	} else if height > maxSide {
		ratio := float32(width) / float32(height)
		return uint(float32(maxSide) * ratio), maxSide
	}

	return width, height
}

func resizeAndCompress(mw *imagick.MagickWand, resizedWith uint, resizedHeight uint) error {
	if err := mw.ResizeImage(resizedWith, resizedHeight, imagick.FILTER_LANCZOS, 1); err != nil {
		return err
	}
	if err := mw.SetImageCompressionQuality(85); err != nil {
		return err
	}
	if err := mw.SetImageFormat("JPG"); err != nil {
		return err
	}

	return nil
}

func cropImage(mw *imagick.MagickWand) error {
	width := mw.GetImageWidth()
	height := mw.GetImageHeight()

}

func main() {
	start := time.Now()

	b, _ := os.ReadFile("image1.png")
	ProcessImages(b)

	fmt.Println(time.Since(start))
}
