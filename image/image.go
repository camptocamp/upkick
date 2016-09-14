package image

type Image struct {
	ID     string
	Hash   string
	Hashes map[string]*ImageHash
}

type ImageHash struct {
	Containers []string
}

// String returns the string representation of an Image
func (i *Image) String() string {
	return i.ID
}
