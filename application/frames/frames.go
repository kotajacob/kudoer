package frames

const Count = 2

// Validate returns true if a given frame is valid.
func Validate(key int) bool {
	return key < Count
}
