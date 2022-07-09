package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"gopkg.in/gographics/imagick.v2/imagick"
	"os"
	"strings"
	"time"
)

const (
	ImgMaxSide       = 1280
	ThumbnailMaxSide = 500
	ThumbnailWidth   = 500
	ThumbnailHeight  = 400
)

func ProcessImages(blob []byte) (string, string, error) {
	imagick.Initialize()
	defer imagick.Terminate()

	mw := imagick.NewMagickWand()
	if err := mw.ReadImageBlob(blob); err != nil {
		return "", "", err
	}

	format := mw.GetImageFormat()
	if format != "PNG" && format != "JPEG" && format != "BMP" && format != "WEBP" {
		return "", "", errors.New("unsupported image format")
	}

	// Setting output format.
	if err := mw.SetImageFormat("JPEG"); err != nil {
		return "", "", err
	}

	var (
		imgBase64       string
		thumbnailBase64 string
	)

	if format != "GIF" {
		// Only image that is not GIF should be processed.
		resWidth, resHeight := getImageRes(mw, ImgMaxSide)
		if err := resizeAndCompress(mw, resWidth, resHeight); err != nil {
			return "", "", err
		}
		imgBase64 = toBase64(mw)

		mw.WriteImage("image2.jpg")
	}

	// Processing thumbnail of any image.
	tResWidth, tResHeight := getThumbnailRes(mw, ThumbnailMaxSide)
	if err := resizeAndCompress(mw, tResWidth, tResHeight); err != nil {
		return "", "", err
	}
	if err := cropImage(mw, ThumbnailWidth, ThumbnailHeight); err != nil {
		return "", "", err
	}
	thumbnailBase64 = toBase64(mw)

	mw.WriteImage("image3.jpg")

	return imgBase64, thumbnailBase64, nil
}

func toBase64(mw *imagick.MagickWand) string {
	str := fmt.Sprintf("data:image/%s;base64,", strings.ToLower(mw.GetImageFormat()))
	str += base64.StdEncoding.EncodeToString(mw.GetImageBlob())
	return str
}

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
	return nil
}

func cropImage(mw *imagick.MagickWand, width uint, height uint) error {
	imgWidth := mw.GetImageWidth()
	imgHeight := mw.GetImageHeight()

	return mw.CropImage(width, height, int((imgWidth-width)/2), int((imgHeight-height)/2))
}

func main() {
	start := time.Now()

	b, _ := os.ReadFile("image1.jpg")
	ProcessImages(b)

	fmt.Println(time.Since(start))
}
