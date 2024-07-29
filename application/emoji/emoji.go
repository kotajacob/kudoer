package emoji

import "math/rand"

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
		Alt: "A pair of eyes, glancing slightly to the left.",
	},
	{
		Key: 1,
		Alt: "A yellow face, smiling and drooling as though thinking of something delicious.",
	},
	{
		Key: 2,
		Alt: "A flame, as produced when something is on fire.",
	},
	{
		Key: 3,
		Alt: "A yellow face with a big grin, uplifted eyebrows, and smiling eyes, each shedding a tear from laughing so hard.",
	},
	{
		Key: 4,
		Alt: "A yellow face with simple, open eyes and a flat, closed mouth.",
	},
	{
		Key: 5,
		Alt: "A red face with an angry expression: frowning mouth with eyes and eyebrows scrunched downward.",
	},
	{
		Key: 6,
		Alt: "A yellow face with simple open eyes showing clenched teeth.",
	},
	{
		Key: 7,
		Alt: "A yellow face with an open mouth wailing and streams of heavy tears flowing from closed eyes.",
	},
	{
		Key: 8,
		Alt: "A person with arms crossed forming an ‘X’ to indicate ‘no’ or ‘no good’.",
	},
	{
		Key: 9,
		Alt: "A yellow face with furrowed eyebrows looking upwards with thumb and index finger resting on its chin.",
	},
	{
		Key: 10,
		Alt: "A yellow face with a broad, open smile, showing upper teeth on most platforms, with stars for eyes, as if seeing a beloved celebrity.",
	},
	{
		Key: 11,
		Alt: "A yellow face with scrunched, X-shaped eyes spewing bright-green vomit.",
	},
	{
		Key: 12,
		Alt: "A yellow face with an open mouth, the top of its head exploding in the shape of a brain-like mushroom cloud.",
	},
	{
		Key: 13,
		Alt: "A yellow face with smiling eyes, a closed smile, rosy cheeks, and several hearts floating around its head.",
	},
	{
		Key: 14,
		Alt: "A yellow face with eyes closed and mouth wide open covered by a hand, mid yawn.",
	},
	{
		Key: 15,
		Alt: "A gold star.",
	},
	{
		Key: 16,
		Alt: "A yellow smiley face melting into a puddle.",
	},
	{
		Key: 17,
		Alt: "A classic red love heart emoji.",
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

// Shuffle returns a shuffled list of all emoji
func Shuffle() []Emoji {
	var shuffled []Emoji
	shuffled = append(shuffled, all...)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})
	return shuffled
}

// Validate returns true if a given emoji ID is valid.
func Validate(key int) bool {
	_, ok := lookup[key]
	return ok
}

// Alt returns the alt text for a given emoji ID.
func Alt(key int) string {
	return lookup[key]
}
