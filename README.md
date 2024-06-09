# kudoer
A site for giving kudos.

## compile
A `go` compiler is required to compile this application. Check `go.mod` for the
oldest supported version of [go](https://go.dev/). Then run `make` to compile
the project.

## license
GNU AGPL version 3 or later, see LICENSE.

## schema
```go
type user struct {
	id      int
	name    string
	email   string
	created time.Time

	reviews []int
}

type kudo struct {
	id      int
	author  int
	rating  string // emoji
	body    string
	reviews []int
}

type item struct {
	id           int
	names        []string
	descriptions []string
	images       []string
	reviews      []int
}
```
