package handler

import (
	"context"
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"

	docker "github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"

	"github.com/camptocamp/upkick/config"
	"github.com/camptocamp/upkick/image"
)

// Upkick is an upkick handler
type Upkick struct {
	*docker.Client
	*config.Config
}

func NewUpkick(version string) (*Upkick, error) {
	u := &Upkick{}
	err := u.setup(version)
	return u, err
}

func (u *Upkick) setup(version string) (err error) {
	u.Config = config.LoadConfig(version)

	err = u.setupLoglevel()
	if err != nil {
		return errors.Wrap(err, "failed to setup log level")
	}

	err = u.setupDocker()
	if err != nil {
		return errors.Wrap(err, "failed to setup Docker")
	}

	return
}

func (u *Upkick) setupLoglevel() (err error) {
	switch u.Config.Loglevel {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	case "panic":
		log.SetLevel(log.PanicLevel)
	default:
		errMsg := fmt.Sprintf("Wrong log level '%v'", u.Config.Loglevel)
		err = errors.New(errMsg)
	}

	if u.Config.JSON {
		log.SetFormatter(&log.JSONFormatter{})
	}

	return
}

func (u *Upkick) setupDocker() (err error) {
	u.Client, err = docker.NewClient(u.Config.Docker.Endpoint, "", nil, nil)
	return
}

// GetImages returns a slice of Image
func (h *Upkick) GetImages() (images map[string]*image.Image, err error) {
	log.Debug("Getting images")
	containers, err := h.Client.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		err = errors.Wrap(err, "failed to list containers")
		return
	}

	images = make(map[string]*image.Image)

	for _, c := range containers {
		cont, _ := h.Client.ContainerInspect(context.Background(), c.ID)

		i, ok := images[c.Image]
		if !ok {
			images[cont.Config.Image] = &image.Image{
				ID: cont.Config.Image,
			}
			i = images[cont.Config.Image]
			i.Hashes = make(map[string]*image.ImageHash)
		}
		h, ok := i.Hashes[c.ImageID]
		if !ok {
			i.Hashes[c.ImageID] = &image.ImageHash{}
			h = i.Hashes[c.ImageID]
		}
		log.Debugf("Adding %s with hash %s to %s", c.ID, c.ImageID, cont.Config.Image)
		h.Containers = append(h.Containers, c.ID)
	}

	return
}

func (h *Upkick) Pull(i *image.Image) (err error) {
	log.Debugf("Pulling Image %s", i)
	_, _ = h.Client.ImagePull(context.Background(), i.ID, types.ImagePullOptions{})
	img, _, _ := h.Client.ImageInspectWithRaw(context.Background(), i.ID)
	i.Hash = img.ID
	log.Infof("Image %s updated to %v", i, i.Hash)
	return nil
}

// kick stops and removes all containers
// using an obsolete version of the Image
func (h *Upkick) Kick(i *image.Image) error {
	log.Debugf("Kicking containers for Image %s", i)

	for hash, hashS := range i.Hashes {
		if hash == i.Hash {
			// Already up-to-date
			log.Debugf("Not kicking containers for up-to-date hash %s", h)
			continue
		}

		for _, c := range hashS.Containers {
			log.Debugf("Kicking container %s", c)
			timeout := 10 * time.Second
			log.Debugf("Stopping container %s", c)
			_ = h.Client.ContainerStop(context.Background(), c, &timeout)
			//			log.Debugf("Removing container %s", c)
			//			_ = h.Client.ContainerRemove(context.Background(), c, types.ContainerRemoveOptions{})
		}
	}

	return nil
}
