# kudoer
A site for giving kudos.

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
