package image

// Image is a Docker image
type Image struct {
	ID     string
	Hash   string
	Hashes map[string]*Hash
}

// Hash is a specific Image
type Hash struct {
	Containers []string
}

// String returns the string representation of an Image
func (i *Image) String() string {
	return i.ID
}
