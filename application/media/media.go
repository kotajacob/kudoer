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
)

const MaxPixels = 512

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
// The named file (sha1sum.jpeg) or an error will be returned.
func (m *MediaStore) StorePic(src io.Reader) (string, error) {
	img, err := imaging.Decode(src, imaging.AutoOrientation(true))
	if err != nil {
		return "", fmt.Errorf("failed decoding image: %v", err)
	}

	// Resize the image.
	img = imaging.Fill(
		img,
		MaxPixels,
		MaxPixels,
		imaging.Center,
		imaging.Lanczos,
	)

	return m.store(img)
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
