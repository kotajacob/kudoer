// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package ui

import (
	"embed"
	"html/template"
	"io/fs"
	"path/filepath"
	"time"
)

const baseTMPL = "base.tmpl"

//go:embed "base.tmpl" "pages" "partials" "static"
var EFS embed.FS

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
				"Date": Date,
			}).
			ParseFS(EFS, files...)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}
	return cache, nil
}

func Date(t time.Time) string {
	return t.Format("January 2, 2006")
}
