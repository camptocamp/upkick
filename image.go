package main

import (
	"fmt"
	"time"

	"github.com/docker/engine-api/types"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

type image struct {
	handler *handler
	id      string
	hash    string
	hashes  map[string]*imageHash
}

type imageHash struct {
	containers []string
}

// getImages returns a slice of image
func (h *handler) getImages() (images map[string]*image, err error) {
	fmt.Println("Getting images")
	containers, err := h.Client.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		err = errors.Wrap(err, "failed to list containers")
		return
	}

	images = make(map[string]*image)

	for _, c := range containers {
		i, ok := images[c.Image]
		if !ok {
			images[c.Image] = &image{
				handler: h,
				id:      c.Image,
			}
			i = images[c.Image]
			i.hashes = make(map[string]*imageHash)
		}
		h, ok := i.hashes[c.ImageID]
		if !ok {
			i.hashes[c.ImageID] = &imageHash{}
			h = i.hashes[c.ImageID]
		}
		fmt.Printf("Adding %s with hash %s to %s", c.ID, c.ImageID, c.Image)
		h.containers = append(h.containers, c.ID)
	}

	return
}

// String returns the string representation of an image
func (i *image) String() string {
	return i.id
}

// pull updates an image
func (i *image) pull() error {
	fmt.Printf("Pulling image %s\n", i)
	_, _ = i.handler.Client.ImagePull(context.Background(), i.id, types.ImagePullOptions{})
	img, _, _ := i.handler.Client.ImageInspectWithRaw(context.Background(), i.id)
	i.hash = img.ID
	fmt.Printf("Image %s updated to %v\n", i, i.hash)
	return nil
}

// kick stops and removes all containers
// using an obsolete version of the image
func (i *image) kick() error {
	fmt.Printf("Kicking containers for image %s\n", i)

	for h, hh := range i.hashes {
		if h == i.hash {
			// Already up-to-date
			fmt.Printf("Not kicking containers for up-to-date hash %s\n", h)
			continue
		}

		for _, c := range hh.containers {
			fmt.Printf("Kicking container %s\n", c)
			timeout := 10 * time.Second
			_ = i.handler.Client.ContainerStop(context.Background(), c, &timeout)
			_ = i.handler.Client.ContainerRemove(context.Background(), c, types.ContainerRemoveOptions{})
		}
	}

	return nil
}
