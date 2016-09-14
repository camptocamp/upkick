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

// NewUpkick returns a new Upkick handler
func NewUpkick(version string) (*Upkick, error) {
	u := &Upkick{}
	err := u.setup(version)
	return u, err
}

// GetImages returns a slice of Image
func (u *Upkick) GetImages() (images map[string]*image.Image, err error) {
	log.Debug("Getting images")
	containers, err := u.Client.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		err = errors.Wrap(err, "failed to list containers")
		return
	}

	images = make(map[string]*image.Image)

	for _, c := range containers {
		cont, err := u.Client.ContainerInspect(context.Background(), c.ID)
		if err != nil {
			msg := fmt.Sprintf("failed to inspect container %s", c.ID)
			return images, errors.Wrap(err, msg)
		}

		i, ok := images[c.Image]
		if !ok {
			images[cont.Config.Image] = &image.Image{
				ID: cont.Config.Image,
			}
			i = images[cont.Config.Image]
			i.Hashes = make(map[string]*image.Hash)
		}
		h, ok := i.Hashes[c.ImageID]
		if !ok {
			i.Hashes[c.ImageID] = &image.Hash{}
			h = i.Hashes[c.ImageID]
		}
		log.Debugf("Adding %s with hash %s to %s", c.ID, c.ImageID, cont.Config.Image)
		h.Containers = append(h.Containers, c.ID)
	}

	return
}

// Pull pulls the newest version of an image
func (u *Upkick) Pull(i *image.Image) (err error) {
	log.Debugf("Pulling Image %s", i)

	_, err = u.Client.ImagePull(context.Background(), i.ID, types.ImagePullOptions{})
	if err != nil {
		msg := fmt.Sprintf("failed to pull image %s", i.ID)
		return errors.Wrap(err, msg)
	}

	img, _, err := u.Client.ImageInspectWithRaw(context.Background(), i.ID)
	if err != nil {
		msg := fmt.Sprintf("failed to inspect image %s", i.ID)
		return errors.Wrap(err, msg)
	}

	i.Hash = img.ID
	log.Infof("Image %s updated to %v", i, i.Hash)

	return
}

// Kick stops and removes all containers
// using an obsolete version of the Image
func (u *Upkick) Kick(i *image.Image) (err error) {
	log.Debugf("Kicking containers for Image %s", i)

	for hash, hashS := range i.Hashes {
		if hash == i.Hash {
			// Already up-to-date
			log.Debugf("Not kicking containers for up-to-date hash %s", hash)
			continue
		}

		for _, c := range hashS.Containers {
			log.Debugf("Stopping container %s", c)
			timeout := 10 * time.Second
			err = u.Client.ContainerStop(context.Background(), c, &timeout)
			if err != nil {
				msg := fmt.Sprintf("failed to stop container %s", c)
				return errors.Wrap(err, msg)
			}

			log.Debugf("Removing container %s", c)
			err = u.Client.ContainerRemove(context.Background(), c, types.ContainerRemoveOptions{})
			if err != nil {
				msg := fmt.Sprintf("failed to remove container %s", c)
				return errors.Wrap(err, msg)
			}
		}
	}

	return
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
