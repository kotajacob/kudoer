// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package media

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"image"
	"io"
	"os"
	"path/filepath"

	"image/jpeg"
	_ "image/jpeg"
	_ "image/png"

	"github.com/disintegration/imaging"
	"github.com/muesli/smartcrop"
	"github.com/muesli/smartcrop/nfnt"
)

type MediaStore struct {
	msn string
}

func Open(msn string) (*MediaStore, error) {
	m := MediaStore{
		msn: msn,
	}
	err := os.MkdirAll(msn, 0700)
	return &m, err
}

func (m *MediaStore) Dir() string {
	return m.msn
}

// StorePic stores a picture.
// The filename is in the format of sha1sum.jpeg. Both a 512px and a 128px
// variant are created and returned in that order.
func (m *MediaStore) StorePic(src io.Reader) (string, string, error) {
	img, err := imaging.Decode(src, imaging.AutoOrientation(true))
	if err != nil {
		return "", "", fmt.Errorf("failed decoding image: %v", err)
	}
	maxSize := max(img.Bounds().Dx(), img.Bounds().Dy())
	analyzer := smartcrop.NewAnalyzer(nfnt.NewDefaultResizer())
	crop, err := analyzer.FindBestCrop(img, maxSize, maxSize)
	if err != nil {
		return "", "", err
	}

	img512, err := resize(img, crop, 512)
	if err != nil {
		return "", "", fmt.Errorf("failed resizing image: %v", err)
	}
	img128, err := resize(img, crop, 128)
	if err != nil {
		return "", "", fmt.Errorf("failed resizing image: %v", err)
	}

	filename512, err := m.store(img512)
	if err != nil {
		return "", "", fmt.Errorf("failed storing image: %v", err)
	}
	filename128, err := m.store(img128)
	if err != nil {
		return "", "", fmt.Errorf("failed storing image: %v", err)
	}
	return filename512, filename128, nil
}

// resize an image to a specific profile picture size.
func resize(img image.Image, crop image.Rectangle, size int) (image.Image, error) {
	// Crop to a square around the content.
	img = imaging.Crop(img, crop)

	// Scale / shrink to the correct final size.
	img = imaging.Fill(
		img,
		size,
		size,
		imaging.Center,
		imaging.Lanczos,
	)
	return img, nil
}

// store an image as a jpeg with a name containing a sha1sum of itself.
func (m *MediaStore) store(img image.Image) (string, error) {
	var buf bytes.Buffer
	err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})
	if err != nil {
		return "", fmt.Errorf("failed encoding image as jpeg: %v", err)
	}

	// Calculate hash for the filename.
	h := sha1.New()
	r := bytes.NewReader(buf.Bytes())
	if _, err := io.Copy(h, r); err != nil {
		return "", fmt.Errorf("failed calculating hash for image: %v", err)
	}
	r.Seek(0, 0)

	name := fmt.Sprintf("%x.jpeg", h.Sum(nil))
	f, err := os.Create(filepath.Join(m.msn, name))
	if err != nil {
		return "", fmt.Errorf("failed creating file to store image: %v", err)
	}

	if _, err := io.Copy(f, r); err != nil {
		return "", fmt.Errorf("failed writing image: %v", err)
	}
	return name, f.Close()
}

// DeletePic deletes a picture.
func (m *MediaStore) DeletePic(name string) error {
	return os.Remove(filepath.Join(m.msn, name))
}
