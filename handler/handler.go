package handler

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"

	"github.com/camptocamp/upkick/config"
	"github.com/camptocamp/upkick/image"
	"github.com/camptocamp/upkick/metrics"
)

var blacklist = []string{
	"rancher/agent",
	"rancher/agent-instance",
	"camptocamp/upkick",
}

// Upkick is an upkick handler
type Upkick struct {
	Client   *docker.Client
	Config   *config.Config
	Hostname string
	Metrics  *metrics.PrometheusMetrics
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
	containers, err := u.Client.ContainerList(context.Background(), types.ContainerListOptions{
		All: true,
	})
	if err != nil {
		err = errors.Wrap(err, "failed to list containers")
		return
	}

	images = make(map[string]*image.Image)

	containersTotal := make(map[string]int)
	blacklistedTags := make(map[string]int)
	blacklistedContainers := make(map[string]int)

	for _, c := range containers {
		cont, err := u.Client.ContainerInspect(context.Background(), c.ID)
		if err != nil {
			log.Errorf("failed to inspect container %s: %v", c.ID, err)
			continue
		}

		tag := cont.Config.Image
		containersTotal[tag]++

		if blacklistedTag(tag) {
			log.Debugf("Ignoring blacklisted image tag %s", tag)
			blacklistedTags[tag]++
			continue
		}

		if blacklistedContainer(cont) {
			log.Debugf("Ignoring blacklisted container %s", cont.ID)
			blacklistedContainers[tag]++
			continue
		}

		i, ok := images[c.Image]
		if !ok {
			images[tag] = &image.Image{
				ID: tag,
			}
			i = images[tag]
			i.Hashes = make(map[string]*image.Hash)
		}
		h, ok := i.Hashes[c.ImageID]
		if !ok {
			i.Hashes[c.ImageID] = &image.Hash{}
			h = i.Hashes[c.ImageID]
		}
		log.Debugf("Adding %s with hash %s to %s", c.ID, c.ImageID, tag)
		h.Containers = append(h.Containers, c.ID)
	}

	var m *metrics.Metric
	for _, i := range images {
		tag := i.ID
		m = u.Metrics.NewMetric("upkick_containers", "gauge")
		m.NewEvent(&metrics.Event{
			Value: strconv.Itoa(containersTotal[tag]),
			Labels: map[string]string{
				"what":  "total",
				"image": tag,
			},
		})
		m.NewEvent(&metrics.Event{
			Value: strconv.Itoa(blacklistedTags[tag]),
			Labels: map[string]string{
				"what":  "blacklisted_tag",
				"image": tag,
			},
		})
		m.NewEvent(&metrics.Event{
			Value: strconv.Itoa(blacklistedContainers[tag]),
			Labels: map[string]string{
				"what":  "blacklisted_container",
				"image": tag,
			},
		})
	}

	return
}

// Pull pulls the newest version of an image
func (u *Upkick) Pull(i *image.Image) (err error) {
	log.Debugf("Pulling Image %s", i)

	var pullOut io.ReadCloser
	pullOut, err = u.Client.ImagePull(context.Background(), i.ID, types.ImagePullOptions{})
	if err != nil {
		msg := fmt.Sprintf("failed to pull image %s", i.ID)
		return errors.Wrap(err, msg)
	}

	// Wait for the image to be fully pulled
	io.Copy(ioutil.Discard, pullOut)
	pullOut.Close()

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

	var noup int
	var outWarn int
	var upOK int
	var upNOK int

	for hash, hashS := range i.Hashes {
		if hash == i.Hash {
			// Already up-to-date
			log.Debugf("Not kicking containers for up-to-date hash %s", hash)
			noup += len(hashS.Containers)
			continue
		}

		for _, c := range hashS.Containers {
			cont, err := u.Client.ContainerInspect(context.Background(), c)
			if err != nil {
				log.Errorf("failed to inspect container %s: %v", c, err)
				continue
			}

			warnOnly, warnLabel := cont.Config.Labels["io.upkick.warn_only"]

			if u.Config.Warn && !(warnLabel && warnOnly == "false") {
				log.Warnf("Container %s uses an out-of-date image", c)
				outWarn++
				continue
			}
			if cont.State.Running {
				log.Infof("Stopping container %s", c)
				timeout := 10 * time.Second
				err = u.Client.ContainerStop(context.Background(), c, &timeout)
				if err != nil {
					upNOK++
					log.Errorf("failed to stop container %s: %v", c, err)
					continue
				}
			} else {
				log.Infof("Container %s already stopped", c)
			}

			log.Infof("Removing container %s", c)
			err = u.Client.ContainerRemove(context.Background(), c, types.ContainerRemoveOptions{})
			if err != nil {
				upNOK++
				log.Errorf("failed to remove container %s: %v", c, err)
				continue
			}
			upOK++
		}
	}

	m := u.Metrics.NewMetric("upkick_containers", "gauge")
	m.NewEvent(&metrics.Event{
		Value: strconv.Itoa(noup),
		Labels: map[string]string{
			"what":  "up_to_date",
			"image": i.ID,
		},
	})
	m.NewEvent(&metrics.Event{
		Value: strconv.Itoa(upOK),
		Labels: map[string]string{
			"what":  "updated",
			"image": i.ID,
		},
	})
	m.NewEvent(&metrics.Event{
		Value: strconv.Itoa(upNOK),
		Labels: map[string]string{
			"what":  "update_failed",
			"image": i.ID,
		},
	})
	m.NewEvent(&metrics.Event{
		Value: strconv.Itoa(outWarn),
		Labels: map[string]string{
			"what":  "not_updated",
			"image": i.ID,
		},
	})

	return
}

// PushMetrics pushes metrics for the handler
func (u *Upkick) PushMetrics() {
	/* Metrics per node:
	 *
	 * - States (gauges):
	 *   - NOUP:     up-to-date containers
	 *   - UP-OK:    container successfully updated
	 *   - UP-NOK:   container failed to update
	 *   - OUT-WARN: container is out-of-date but not updated
	 *
	 * - Image timestamp per hash (counter)
	 */
	u.Metrics.Push()
}

func (u *Upkick) setup(version string) (err error) {
	u.Config = config.LoadConfig(version)

	err = u.setupLoglevel()
	if err != nil {
		return errors.Wrap(err, "failed to setup log level")
	}

	err = u.getHostname()
	if err != nil {
		return errors.Wrap(err, "failed to get hostname")
	}

	err = u.setupDocker()
	if err != nil {
		return errors.Wrap(err, "failed to setup Docker")
	}

	err = u.setupMetrics()
	if err != nil {
		return errors.Wrap(err, "failed to setup metrics")
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

func (u *Upkick) getHostname() (err error) {
	if u.Config.HostnameFromRancher {
		resp, err := http.Get("http://rancher-metadata/latest/self/host/name")
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		u.Hostname = string(body)
	} else {
		u.Hostname, err = os.Hostname()
	}
	return
}

func (u *Upkick) setupDocker() (err error) {
	u.Client, err = docker.NewClient(u.Config.Docker.Endpoint, "", nil, nil)
	return
}

func (u *Upkick) setupMetrics() (err error) {
	u.Metrics = metrics.NewMetrics(u.Hostname, u.Config.Metrics.PushgatewayURL)
	return
}

func blacklistedTag(tag string) bool {
	baseImage := strings.Split(tag, ":")[0]

	for _, b := range blacklist {
		if baseImage == b {
			return true
		}
	}

	return false
}

func blacklistedContainer(cont types.ContainerJSON) bool {
	if l, ok := cont.Config.Labels["io.upkick.warn_only"]; ok && l == "true" {
		return true
	}

	return false
}
