package main

type image struct {
	tag string
}

// getImages returns a slice of image
func getImages() (images []image) {
	return
}

// String returns the string representation of an image
func (i *image) String() string {
	return i.tag
}

// pull updates an image
func (i *image) pull() error {
	return nil
}

// kick stops and removes all containers
// using an obsolete version of the image
func (i *image) kick() error {
	return nil
}
