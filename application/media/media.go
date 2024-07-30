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

	"github.com/nfnt/resize"
	orientation "github.com/takumakei/exif-orientation"
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
// The named file (sha1sum.imgformat) or an error will be returned.
func (m *MediaStore) StorePic(src io.ReadSeeker) (string, error) {
	img, _, err := image.Decode(src)
	if err != nil {
		return "", err
	}

	// Rotate the image using its EXIF data if needed.
	_, err = src.Seek(0, 0)
	if err != nil {
		return "", err
	}
	o, err := orientation.Read(src)
	if err != nil {
		return "", err
	}
	img = orientation.Normalize(img, o)

	// Resize the image.
	img = resize.Resize(
		MaxPixels,
		MaxPixels,
		img,
		resize.NearestNeighbor,
	)

	// Encode the image as jpeg.
	var b bytes.Buffer
	err = jpeg.Encode(&b, img, &jpeg.Options{Quality: 90})
	if err != nil {
		return "", err
	}

	// Calculate its hash for the filename.
	h := sha1.New()
	r := bytes.NewReader(b.Bytes())
	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}
	r.Seek(0, 0)

	name := fmt.Sprintf("%x.jpeg", h.Sum(nil))
	f, err := os.Create(filepath.Join(m.msn, name))
	if err != nil {
		return "", err
	}

	if _, err := io.Copy(f, r); err != nil {
		return "", err
	}
	return name, f.Close()
}

// DeletePic deletes a picture.
func (m *MediaStore) DeletePic(name string) error {
	return os.Remove(filepath.Join(m.msn, name))
}
