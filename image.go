package main

import (
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/docker/engine-api/types"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

type image struct {
	handler *Upkick
	id      string
	hash    string
	hashes  map[string]*imageHash
}

type imageHash struct {
	containers []string
}

// getImages returns a slice of image
func (h *Upkick) getImages() (images map[string]*image, err error) {
	log.Debug("Getting images")
	containers, err := h.Client.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		err = errors.Wrap(err, "failed to list containers")
		return
	}

	images = make(map[string]*image)

	for _, c := range containers {
		cont, _ := h.Client.ContainerInspect(context.Background(), c.ID)

		i, ok := images[c.Image]
		if !ok {
			images[cont.Config.Image] = &image{
				handler: h,
				id:      cont.Config.Image,
			}
			i = images[cont.Config.Image]
			i.hashes = make(map[string]*imageHash)
		}
		h, ok := i.hashes[c.ImageID]
		if !ok {
			i.hashes[c.ImageID] = &imageHash{}
			h = i.hashes[c.ImageID]
		}
		log.Debugf("Adding %s with hash %s to %s", c.ID, c.ImageID, cont.Config.Image)
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
	log.Debugf("Pulling image %s", i)
	_, _ = i.handler.Client.ImagePull(context.Background(), i.id, types.ImagePullOptions{})
	img, _, _ := i.handler.Client.ImageInspectWithRaw(context.Background(), i.id)
	i.hash = img.ID
	log.Infof("Image %s updated to %v", i, i.hash)
	return nil
}

// kick stops and removes all containers
// using an obsolete version of the image
func (i *image) kick() error {
	log.Debugf("Kicking containers for image %s", i)

	for h, hh := range i.hashes {
		if h == i.hash {
			// Already up-to-date
			log.Debugf("Not kicking containers for up-to-date hash %s", h)
			continue
		}

		for _, c := range hh.containers {
			log.Debugf("Kicking container %s", c)
			timeout := 10 * time.Second
			log.Debugf("Stopping container %s", c)
			_ = i.handler.Client.ContainerStop(context.Background(), c, &timeout)
			//			log.Debugf("Removing container %s", c)
			//			_ = i.handler.Client.ContainerRemove(context.Background(), c, types.ContainerRemoveOptions{})
		}
	}

	return nil
}
