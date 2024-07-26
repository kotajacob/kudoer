// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package ui

import (
	"crypto/sha1"
	"embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"path"
	"path/filepath"

	"git.sr.ht/~kota/kudoer/application/emoji"
	"github.com/oklog/ulid"
)

const baseTMPL = "base.tmpl"

//go:embed "base.tmpl" "pages" "partials" "static"
var EFS embed.FS

type StaticFiles struct {
	// ToHash maps static filenames to the hashed version.
	ToHash map[string]string

	// FromHash maps hashed filenames back to their originals.
	FromHash map[string]string
}

var statics StaticFiles

func init() {
	var err error
	statics, err = Statics()
	if err != nil {
		panic(err)
	}
}

func Statics() (StaticFiles, error) {
	entries, err := fs.ReadDir(EFS, "static")
	if err != nil {
		return StaticFiles{}, err
	}

	to := make(map[string]string, len(entries))
	from := make(map[string]string, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			return StaticFiles{}, fmt.Errorf("directories cannot be served from static")
		}

		f, err := EFS.Open(filepath.Join("static", entry.Name()))
		if err != nil {
			return StaticFiles{}, err
		}
		defer f.Close()

		h := sha1.New()
		if _, err := io.Copy(h, f); err != nil {
			return StaticFiles{}, err
		}
		plain := path.Join("/static/", entry.Name())
		hashed := fmt.Sprintf("/static/%x.%v", h.Sum(nil), entry.Name())
		to[plain] = hashed
		from[hashed] = plain
	}
	return StaticFiles{ToHash: to, FromHash: from}, nil
}

func Templates() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	partials, err := fs.Glob(EFS, "partials/*.tmpl")
	if err != nil {
		panic(err)
	}

	pages, err := fs.Glob(EFS, "pages/*.tmpl")
	if err != nil {
		panic(err)
	}

	for _, page := range pages {
		name := filepath.Base(page)
		files := []string{baseTMPL}
		files = append(files, partials...)
		files = append(files, page)

		ts, err := template.New(baseTMPL).
			Funcs(template.FuncMap{
				"Date":     Date,
				"ToHash":   ToHash,
				"FromHash": FromHash,
				"EmojiAlt": emoji.Alt,
			}).
			ParseFS(EFS, files...)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}
	return cache, nil
}

func Date(id ulid.ULID) string {
	return ulid.Time(id.Time()).Format("January 2, 2006")
}

// ToHash maps static filenames to the hashed version.
func ToHash(f string) string {
	return statics.ToHash[f]
}

// FromHash maps hashed filenames back to their originals.
func FromHash(f string) string {
	return statics.FromHash[f]
}
