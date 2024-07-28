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
	"image/png"
	_ "image/png"

	"github.com/nfnt/resize"
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
	img, format, err := image.Decode(src)
	if err != nil {
		return "", err
	}
	img = resize.Resize(
		MaxPixels,
		MaxPixels,
		img,
		resize.NearestNeighbor,
	)

	var b bytes.Buffer
	switch format {
	case "jpeg":
		err := jpeg.Encode(&b, img, &jpeg.Options{Quality: 95})
		if err != nil {
			return "", err
		}
	case "png":
		err := png.Encode(&b, img)
		if err != nil {
			return "", err
		}
	default:
		return "", image.ErrFormat
	}

	h := sha1.New()
	r := bytes.NewReader(b.Bytes())
	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}
	r.Seek(0, 0)

	name := fmt.Sprintf("%x.%v", h.Sum(nil), format)
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
