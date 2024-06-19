package emoji

import "fmt"

// Emoji represents an emoji used for a kudo. The Key is an integer, which is
// used in the forms and stored in our database. The Value is the actual
// corresponding unicode symbol.
type Emoji struct {
	Key   int
	Value string
}

var all = []Emoji{
	{
		Key:   1,
		Value: "🤮",
	},
	{
		Key:   2,
		Value: "🫠",
	},
	{
		Key:   3,
		Value: "🤔",
	},
	{
		Key:   4,
		Value: "😐",
	},
	{
		Key:   5,
		Value: "🥰",
	},
	{
		Key:   6,
		Value: "🤩",
	},
}

var lookup map[int]string

func init() {
	lookup = make(map[int]string, len(all))
	for _, e := range all {
		lookup[e.Key] = e.Value
	}
}

// List returns a list of all Emoji.
func List() []Emoji {
	return all
}

// Value returns the actual unicode symbol for a given Emoji key.
func Value(key int) (string, error) {
	if value, ok := lookup[key]; ok {
		return value, nil
	}
	return "", fmt.Errorf("invalid emoji %v\n", key)
}
