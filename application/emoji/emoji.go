package emoji

// Emoji represents an emoji used for a kudo. The Key is an integer, which is
// used in the forms, to reference the svgs, and stored in our database. Alt
// is the alt text to be added to the emoji images.
type Emoji struct {
	Key int
	Alt string
}

var all = []Emoji{
	{
		Key: 0,
		Alt: "",
	},
	{
		Key: 1,
		Alt: "",
	},
	{
		Key: 2,
		Alt: "",
	},
	{
		Key: 3,
		Alt: "",
	},
	{
		Key: 4,
		Alt: "",
	},
	{
		Key: 5,
		Alt: "",
	},
	{
		Key: 6,
		Alt: "",
	},
	{
		Key: 7,
		Alt: "",
	},
	{
		Key: 8,
		Alt: "",
	},
	{
		Key: 9,
		Alt: "",
	},
	{
		Key: 10,
		Alt: "",
	},
	{
		Key: 11,
		Alt: "",
	},
	{
		Key: 12,
		Alt: "",
	},
	{
		Key: 13,
		Alt: "",
	},
	{
		Key: 14,
		Alt: "",
	},
	{
		Key: 15,
		Alt: "",
	},
	{
		Key: 16,
		Alt: "",
	},
	{
		Key: 17,
		Alt: "",
	},
}

var lookup map[int]string

func init() {
	lookup = make(map[int]string, len(all))
	for _, e := range all {
		lookup[e.Key] = e.Alt
	}
}

// List returns a list of all Emoji.
func List() []Emoji {
	return all
}

// Validate returns true if a given emoji ID is valid.
func Validate(key int) bool {
	_, ok := lookup[key]
	return ok
}
