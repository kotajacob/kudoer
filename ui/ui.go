// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package ui

import (
	"embed"
	"html/template"
	"io/fs"
	"path/filepath"
)

const baseTMPL = "base.tmpl"

//go:embed "base.tmpl" "pages"
var EFS embed.FS

func Templates() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := fs.Glob(EFS, "pages/*.tmpl")
	if err != nil {
		panic(err)
	}

	for _, page := range pages {
		name := filepath.Base(page)
		files := []string{baseTMPL}
		files = append(files, page)

		ts, err := template.New(baseTMPL).ParseFS(EFS, files...)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}
	return cache, nil
}
